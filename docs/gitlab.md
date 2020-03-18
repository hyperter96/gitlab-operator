# GitLab

The GitLab resource created by the operator, can be queried like any other native kubernetes resources such as pods and deployments as shown below. In my example, there is only once gitlab instance called `example`:

```
$ kubectl -n operators get gitlab
NAME      AGE
example   18h
$
```

The GitLab instance will deploy the database, redis instance and the GitLab application pod all named with the instance name as a prefix.  

```
$ kubectl -n operators get po -l app.kubernetes.io/managed-by=gitlab-operator
NAME                              READY   STATUS    RESTARTS   AGE
example-database-0                1/1     Running   0          26h
example-gitlab-6485b58488-sbk55   1/1     Running   0          26h
example-redis-0                   1/1     Running   0          26h
$
```

## Configurations

```
$ kubectl -n operators get gitlab example -o yaml
apiVersion: gitlab.com/v1beta1
kind: Gitlab
metadata:
  creationTimestamp: "2020-03-17T12:47:11Z"
  generation: 1
  name: example
  namespace: operators
  resourceVersion: "3865544"
  selfLink: /apis/gitlab.com/v1beta1/namespaces/operators/gitlabs/example
  uid: f198c447-f8ce-4c50-88f7-93ebad382203
spec:
  certificate: gitlab-tls
  email:
    authentication: login
    domain: smtp.gmail.com
    enable: true
    enableStartTLS: true
    host: smtp.gmail.com
    opensslVerifyMode: peer
    password: *****************
    port: 587
    tls: false
    username: <email_addr_for_gitlab>@gmail.com
  enterprise: false
  externalURL: gitlab.mydomain.tld
  registry:
    enable: true
    externalURL: registry.mydomain.tld
  replicas: 1
  volumes:
    config:
      capacity: 2Gi
      persist: true
    data:
      capacity: 10Gi
      persist: false
    database:
      capacity: 8Gi
      persist: true
    redis:
      capacity: 4Gi
      persist: true
    registry:
      capacity: 10Gi
      persist: true
$
```
