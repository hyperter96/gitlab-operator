<!--
NOTICE: This Issue tracker is for the GitLab Operator, not the GitLab Rails application.

Support: Please do not raise support issues for GitLab.com on this tracker. See https://about.gitlab.com/support/
-->

## Summary

(Summarize the bug encountered, concisely as possible)

## Steps to reproduce

(Please provide the steps to reproduce the issue)

## Configuration used

(Please provide a _sanitized_ version of the GitLab Custom Resource `spec` or any other relevant configuration used wrapped in a YAML code block (starting with ` ```yaml `). Make sure to remove/change any sensitive or identifying information.)

```yaml
(Paste sanitized configuration here)
```

## Current behavior

(What you're experiencing happening)

## Expected behavior

(What you're expecting to happen)

## Versions

- Operator: (tagged version | branch | hash `git rev-parse HEAD`)
- Platform:
  - Cloud: (GKE | AKS | EKS | ?)
  - Self-hosted: (OpenShift | Minikube | Rancher RKE | ?)
- Kubernetes: (`kubectl version`)
  - Client:
  - Server:

## Relevant logs

(Please provide any relevant log snippets you have collected, using code blocks (```) to format)
<!--
  Logs for the operator can be collected with the following command:

      kubectl logs -l control-plane=controller-manager
-->

<!--
  If there are a number of logs to attach or the logs are very long, it is
  suggested that each code blocked log be enclosed in a <details></details> block.
  See: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/details
-->
