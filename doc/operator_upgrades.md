# Upgrading the Operator

Below are instructions to upgrade the GitLab Operator.

## Step 1: Identify the desired version of the Operator

To use a released version, the tags will look something like `X.Y.Z-betaN`. See our
[tags](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/tags) for the full list.

If you wish to use a development version, you can find the commit on `master` that you'd like to use,
and copy the first 8 characters of that commit SHA. This aligns with the tag we apply to each
image built in the `master` pipelines (using `$CI_COMMIT_SHORT_SHA`).

Picking a specific SHA or tag is more reliable than using the `latest` tag, which is overridden with each commit to `master`.

Note that by default, the Operator deployment manifest specifies `imagePullPolicy=Always`. This ensures that if the tag
`latest` is used, deleting the pod will pull `latest` again and pull in the latest version of the image under that tag.

## Step 2: Update the Operator deployment with the desired version

The next step is to instruct the Operator deployment to use the desired version of the Operator image. This can be done
multiple ways - the simplest would be to run:

```shell
TAG=abcd1234 make deploy_operator
```

This will instruct `kustomize` to patch the [Operator deployment manifest](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/blob/master/config/manager/manager.yaml) with the desired
tag and send that Deployment manifest to the cluster.

Alternatively, you can edit the Deployment in the cluster directly and enter the desired image tag.

## Step 3: Confirm that the new version of the Operator becomes the leader

The Deployment should create a new ReplicaSet with this change, which will spawn a new Operator pod. Meanwhile, the previous
Operator pod will start to shut down, giving up its leader status. When this happens, the new Operator pod will become the leader.

If the new version of the Operator contains updated logic, you should see it start taking action on the resources in the namespace.

Keep an eye on the logs from the new Operator pod. If you notice any errors, check our
[issue tracker](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues) to see if the issue is known. If not,
open a new issue.
