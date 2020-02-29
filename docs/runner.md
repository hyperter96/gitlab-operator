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
