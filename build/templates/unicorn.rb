# This file should be equivalent to `unicorn.rb` from:
# * gitlab-foss: https://gitlab.com/gitlab-org/gitlab-foss/blob/master/config/unicorn.rb.example
# * omnibus: https://gitlab.com/gitlab-org/omnibus-gitlab/blob/master/files/gitlab-cookbooks/gitlab/templates/default/unicorn.rb.erb
worker_processes 2
working_directory "/srv/gitlab"
listen "0.0.0.0:8080", :tcp_nopush => true
timeout 60
pid "/home/git/unicorn.pid"
preload_app true

require_relative "/srv/gitlab/lib/gitlab/cluster/lifecycle_events"

before_exec do |server|
  # Signal application hooks that we're about to restart
  Gitlab::Cluster::LifecycleEvents.do_before_master_restart
end

run_once = true
before_fork do |server, worker|
  if run_once
    # There is a difference between Puma and Unicorn:
    # - Puma calls before_fork once when booting up master process
    # - Unicorn runs before_fork whenever new work is spawned
    # To unify this behavior we call before_fork only once (we use
    # this callback for deleting Prometheus files so for our purposes
    # it makes sense to align behavior with Puma)
    run_once = false

    # Signal application hooks that we're about to fork
    Gitlab::Cluster::LifecycleEvents.do_before_fork
  end

  # The following is only recommended for memory/DB-constrained
  # installations.  It is not needed if your system can house
  # twice as many worker_processes as you have configured.
  #
  # This allows a new master process to incrementally
  # phase out the old master process with SIGTTOU to avoid a
  # thundering herd (especially in the "preload_app false" case)
  # when doing a transparent upgrade.  The last worker spawned
  # will then kill off the old master process with a SIGQUIT.
  old_pid = "#{server.config[:pid]}.oldbin"
  if old_pid != server.pid
    begin
      sig = (worker.nr + 1) >= server.worker_processes ? :QUIT : :TTOU
      Process.kill(sig, File.read(old_pid).to_i)
    rescue Errno::ENOENT, Errno::ESRCH
    end
  end
  #
  # Throttle the master from forking too quickly by sleeping.  Due
  # to the implementation of standard Unix signal handlers, this
  # helps (but does not completely) prevent identical, repeated signals
  # from being lost when the receiving process is busy.
  # sleep 1
end

after_fork do |server, worker|
  # Signal application hooks of worker start
  Gitlab::Cluster::LifecycleEvents.do_worker_start

  # per-process listener ports for debugging/admin/migrations
  # addr = "127.0.0.1:#{9293 + worker.nr}"
  # server.listen(addr, :tries => -1, :delay => 5, :tcp_nopush => true)
end

ENV['GITLAB_UNICORN_MEMORY_MIN'] = (700 * 1 << 20).to_s
ENV['GITLAB_UNICORN_MEMORY_MAX'] = (1024 * 1 << 20).to_s
