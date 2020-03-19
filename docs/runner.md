# Runner

GitLab runner is a project that would be used to run your continuous integration jobs and sends the results results back to GitLab. A runner has to be registered to GitLab to be able to be used.

The gitlab-operator can be used to deployed and register a runner against a GitLab instance created by the operator or any other GitLab instance reachable over the network.

When registering to a GitLab instance that was not created by the operator, be sure to provide the URL of the instance and the runner registration token which can be found from within your GitLab web interface.

However, when registering to a gitlab instance deployed by the operator, normally to the same namespace, providing a name only would be sufficient to register the runner.

The output below shows a GitLab runner pod which was able to successfully register:

```
$ kubectl -n operators logs example-runner-8464dc856-78c2c
+ cp /scripts/config.toml /etc/gitlab-runner/
+ /entrypoint register --non-interactive --executor kubernetes
Runtime platform                                    arch=amd64 os=linux pid=7 revision=1b659122 version=12.8.0
Running in system-mode.                            

Registering runner... succeeded                     runner=ETCOAAWi
Runner registered successfully. Feel free to start it, but if it's running already the config should be automatically reloaded!
+ /entrypoint run --user=gitlab-runner --working-directory=/home/gitlab-runner
Runtime platform                                    arch=amd64 os=linux pid=18 revision=1b659122 version=12.8.0
Starting multi-runner from /etc/gitlab-runner/config.toml...  builds=0
Running in system-mode.                            

Configuration loaded                                builds=0
listen_address not defined, metrics & debug endpoints disabled  builds=0
[session_server].listen_address not defined, session endpoints disabled  builds=0
$
```

## Clean up
Whenever deleting a runner, prior to deleting, it will de-register from the GitLab instance to prevent any further jobs being sent to a non-existent  runner.

As a result, no further clean up is required.
