#!/usr/bin/env ruby

# The push_to_redhat.rb script will retag a number of container images
# and push them to Red Hat for certification tests. Each image has a
# unique ID and a pull secret used to access the Red Hat registry. These
# values are found in the REDHAT_SECRETS_JSON variable with is an encoded
# JSON string.

require 'json'
require 'digest'

$GITLAB_REGISTRY = ENV['GITLAB_REGISTRY_BASE_URL'] ||
                   File.dirname(ENV['CI_REGISTRY_IMAGE']) ||
                   'registry.gitlab.com/gitlab-org'
$REDHAT_REGISTRY = ENV['REDHAT_REGISTRY_HOSTNAME'] ||
                   'scan.connect.redhat.com'
$IMAGE_VERSION_VAR = { 'gitlab-operator' => 'OPERATOR_VERSION' }

def retag_image(name, version, proj_id)
  gitlab_tag = "#{version}"
  redhat_tag = gitlab_tag # we using matching tags in the operator.
  new_container_name = "#{$REDHAT_REGISTRY}/#{proj_id}/#{name}:#{redhat_tag}"

  puts "Retagging #{$GITLAB_REGISTRY}/#{name}:#{gitlab_tag} to #{new_container_name}"
  %x(docker tag #{$GITLAB_REGISTRY}/#{name}:#{gitlab_tag} #{new_container_name})
  new_container_name
end

def set_credentials(secret)
  puts "Setting credentials (secret SHA1=#{Digest::SHA1.hexdigest secret})"
  %x(echo '#{secret}' | docker login -u unused --password-stdin #{$REDHAT_REGISTRY})
end

def pull_image(image)
  puts "Pulling #{image}"
  %x(docker pull #{image})
end

def push_image(image)
  puts "Pushing #{image}"
  %x(docker push #{image})
end

def is_regular_tag
  ENV['CI_COMMIT_TAG'] || ENV['GITLAB_TAG']
end

if ARGV.length < 1
  puts "Need to specify a version (i.e. v13.5.4)"
  exit 1
end

version = ARGV[0]

# pull in the secrets used to auth with Red Hat registries (CI var)
begin
  if ENV.include? 'REDHAT_SECRETS_JSON'
    secrets = JSON.parse(ENV['REDHAT_SECRETS_JSON'])
  else
    puts "REDHAT_SECRETS_JSON must be defined in the environment"
    exit(1)
  end
rescue => e
  puts "Unable to parse JSON: #{e.message}"
  puts e.backtrace
  raise
end

puts "Using #{version} as the docker tag to pull"

errors = []
$IMAGE_VERSION_VAR.keys.each do |name|
  if secrets.has_key? name
    # pull the image from the GitLab registry
    response = pull_image("#{$GITLAB_REGISTRY}/#{name}:#{version}")
    if response.empty?
      puts "Skipping #{$GITLAB_REGISTRY}/#{name}:#{version} (Not Found)"
      errors << "#{name}: image not found with #{version} tag"
      next
    end

    # retag the image with the Red Hat registry information
    container_name = retag_image(name, version, secrets[name]['id'])

    # each image has separate creds, so need to re auth
    result = set_credentials(secrets[name]['secret']).chomp
    if result != 'Login Succeeded'
      puts "***** Failed to authenticate to registry for #{name} *****"
      puts "#{result}\n"
      errors << "#{name}: Unable to authentcate to registry (bad secret?)"
      next
    end

    # push image to Red Hat and display the response received
    puts push_image(container_name)
  else
    # let someone know that there was not a secret for a specific image
    puts "No entry for #{name} in secrets file"
    errors << "#{name}: No secret listed in $REDHAT_SECRETS_JSON"
  end
end

# display the collected errors in the CI job output
unless errors.empty?
  puts "\n\nThe following errors have been collected:"
  errors.each { |err|
    puts "\t- #{err}"
  }
  exit(1)
end