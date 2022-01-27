apiVersion: apps.gitlab.com/v1beta1
kind: GitLab
metadata:
  name: gitlab
spec:
  chart:
    version: "$GITLAB_CHART_VERSION"
    values:
      global:
        hosts:
          domain: $GITLAB_OPERATOR_DOMAIN
        pages:
          enabled: true
        ingress:
          configureCertmanager: false
          tls:
            secretName:  custom-gitlab-tls
        shell:
          port: 32022
      gitlab:
        gitlab-pages:
          ingress:
            tls:
              secretName: custom-pages-tls
        gitlab-shell:
          minReplicas: 1
          maxReplicas: 1
        gitlab-exporter:
          enabled: true 
        webservice:
          minReplicas: 1
          maxReplicas: 1
      nginx-ingress:
        controller:
          ingressClassResource:
            enabled: true
          service:
            nodePorts:
              # https port value below must match the KinD config file:
              #   nodes[0].extraPortMappings[0].containerPort
              https: 32443
          replicaCount: 1
          minAvailable: 1
        defaultBackend:
          replicaCount: 1
      registry:
        hpa:
          minReplicas: 1
          maxReplicas: 1