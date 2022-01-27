apiVersion: apps.gitlab.com/v1beta1
kind: GitLab
metadata:
  name: gitlab
spec:
  chart:
    version: "$GITLAB_CHART_VERSION"
    values:
      certmanager-issuer:
        email: $GITLAB_ACME_EMAIL
      global:
        hosts:
          domain: $GITLAB_OPERATOR_DOMAIN
        pages:
          enabled: true
      gitlab:
        gitlab-pages:
          ingress:
            tls:
              secretName: custom-pages-tls
