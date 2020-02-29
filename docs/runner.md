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
