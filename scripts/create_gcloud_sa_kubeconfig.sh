#!/bin/bash -e

SANAME='gitlab-operator-gke-cluster'
PROJECT='cloud-native-182609'
CLUSTERNAME='gitlab-operator'
CLUSTERZONE='europe-west3-a'

gcloud iam service-accounts create $SANAME

gcloud projects add-iam-policy-binding $PROJECT \
  --member=serviceAccount:$SANAME@$PROJECT.iam.gserviceaccount.com \
  --role=roles/container.admin  # container.admin is needed to create cluster-scope objects

gcloud projects add-iam-policy-binding $PROJECT \
  --member=serviceAccount:$SANAME@$PROJECT.iam.gserviceaccount.com \
  --role=roles/dns.admin  # dns.admin is needed to create DNS records for LetsEncrypt

gcloud iam service-accounts keys create gsa-key.json \
  --iam-account=$SANAME@$PROJECT.iam.gserviceaccount.com

ENDPOINT=$(gcloud container clusters describe $CLUSTERNAME \
    --zone=$CLUSTERZONE --format="value(endpoint)")

CACERT=$(gcloud container clusters describe $CLUSTERNAME \
    --zone=$CLUSTERZONE --format="value(masterAuth.clusterCaCertificate)")

(
cat <<EOF
apiVersion: v1
kind: Config
clusters:
- name: $CLUSTERNAME
  cluster:
    server: https://$ENDPOINT
    certificate-authority-data: $CACERT
users:
- name: $SANAME
  user:
    auth-provider:
      name: gcp
contexts:
- context:
    cluster: $CLUSTERNAME
    user: $SANAME
  name: $CLUSTERNAME-ci
current-context: $CLUSTERNAME-ci
EOF
) > gsa-kubeconfig.yaml
