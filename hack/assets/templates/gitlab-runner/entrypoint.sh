#!/bin/bash
set -e
mkdir -p ~/.gitlab-runner/
cp /scripts/config.toml ~/.gitlab-runner/

# Register the runner
if [[ -f /secrets/accesskey && -f /secrets/secretkey ]]; then
    export CACHE_S3_ACCESS_KEY=$(cat /secrets/accesskey)
    export CACHE_S3_SECRET_KEY=$(cat /secrets/secretkey)
fi

if [[ -f /secrets/gcs-applicaton-credentials-file ]]; then
    export GOOGLE_APPLICATION_CREDENTIALS="/secrets/gcs-applicaton-credentials-file"
else
    if [[ -f /secrets/gcs-access-id && -f /secrets/gcs-private-key ]]; then
    export CACHE_GCS_ACCESS_ID=$(cat /secrets/gcs-access-id)
    # echo -e used to make private key multiline (in google json auth key private key is oneline with \n)
    export CACHE_GCS_PRIVATE_KEY=$(echo -e $(cat /secrets/gcs-private-key))
    fi
fi

if ! sh /scripts/register-runner; then
    exit 1
fi

# Start the runner
exec /entrypoint run --user=gitlab-runner --listen-address=0.0.0.0
