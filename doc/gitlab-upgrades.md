# Upgrading GitLab

The GitLab Operator is capable of managing upgrades between versions of GitLab. This document includes background context on how the upgrade flow works under the hood, along with instructions to perform a GitLab upgrade.

## How the Operator handles GitLab upgrades

At the beginning of the controller reconcile loop, the Operator checks if the current version matches the desired version.
- If these versions match, then the regular reconcile loop executes, ensuring objects exist that satisfy the configuration provided in the CR spec.
- If these versions _do not_ match, the regular reconcile loop still executes, but an additional branch of logic executes to handle the upgrade flow.

The upgrade flow behaves like this:

1. The controller reconciles all Deployments.
   - The Webservice and Sidekiq Deployments are reconciled but are "paused". This means that the "old" pods stay up until the new Deployments are unpaused.
1. Pre-migrations run.
   - This effectively just runs the Migrations job, but skips post-deployment migrations.
1. The controller unpauses the Webservice and Sidekiq Deployments.
1. The controller waits for the new Webservice and Sidekiq pods to be running.
1. Post-migrations run.
   - This runs the Migrations job (without skipping post-deployment migrations).
1. The controller performs a rolling update on the Webservice and Sidekiq Deployments.
1. The controller waits for the restarted Webservice and Sidekiq pods to be running.

In future reconcile loops, this branch of logic will be skipped because the desired version (from `spec.chart.version`) will match the current version (from `status.version`).

## How to update GitLab

Below are the steps to upgrade a GitLab instance using the GitLab Operator.

### Step 1

Update your GitLab CR's `spec.chart.version` field to a new version. For example:

```diff
apiVersion: apps.gitlab.com/v1beta1
kind: GitLab
metadata:
  name: gitlab
spec:
  chart:
-   version: "5.0.6"
+   version: "5.1.1"
    values:
      ...
```

### Step 2

Apply your modified GitLab CR to the cluster:

```
$ kubectl -n gitlab-system apply -f mygitlab.yaml
```

You should see the following message:

> gitlab.apps.gitlab.com/gitlab created

### Step 3

You can watch the progress via the controller logs:

```
$ kubectl -n gitlab-system logs deployment/gitlab-controller-manager -c manager -f
2021-09-14T20:59:12.342Z        INFO    controllers.GitLab      Reconciling GitLab    {"gitlab": "gitlab-system/gitlab"}
2021-09-14T20:59:12.344Z        DEBUG   controllers.GitLab      version information   {"gitlab": "gitlab-system/gitlab", "upgrade": true, "current version": "", "desired version": "5.0.6"}
2021-09-14T20:59:18.168Z        INFO    controllers.GitLab      reconciling Webservice and Sidekiq Deployments (paused) {"gitlab": "gitlab-system/gitlab"}
...
```

You will see log entries following the upgrade flow outlined above.

You can also view the GitLab CR status in the cluster:

```shell
$ kubectl -n gitlab-system get gitlab
NAME     STATUS        VERSION
gitlab   Preparing     5.2.4
```

When the application is ready and upgraded to the new version, you'll see it reflected in the `STATUS` column.

```shell
$ kubectl -n gitlab-system get gitlab
NAME     STATUS      VERSION
gitlab   Running     5.2.4
```

Status conditions on the GitLab object itself present more detailed information about the application.
