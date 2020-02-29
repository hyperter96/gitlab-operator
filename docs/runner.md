The runner resource can be deployed and linked to a Gitlab.com or any other Gitlab instance that was not created using the operator.

In such a case, be sure to provide the URL of the instance and the runner registration token.

However, when registering to a gitlab instance deployed to the same namespace that was also created by the operator, providing a name only would be sufficient to register the runner.

Note: The runner will default to the Kubernetes executor.

```
$ kubectl -n operators get runner
NAME      AGE
example   8s
$
```

```
$ kubectl -n operators get po -l app.kubernetes.io/part-of=runner
NAME                              READY   STATUS    RESTARTS   AGE
example-runner-8464dc856-78c2c    1/1     Running   0          10m
$
```

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
