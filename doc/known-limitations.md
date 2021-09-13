## Known limitations

Below are known limitations of the GitLab Operator.

### Multiple instances of Webservice not supported on OpenShift

Multiple Webservice instances are problematic on OpenShift. The Ingresses report "All hosts are taken by other resources" when using NGINX Ingress Operator.

Related: [#160](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/160)

### Certain components not supported

Below is a list of unsupported components:

- GitLab Shell [#58](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/58) is unable to provide SSH access.
- Praefect: [#136](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/136)
- Pages: [#138](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/138)
- KAS: [#139](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/139)
- Mailroom: [#140](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/140)
