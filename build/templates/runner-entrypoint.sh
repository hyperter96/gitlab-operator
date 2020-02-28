#!/bin/bash

set -xe

cp /scripts/config.toml /etc/gitlab-runner/

# Register the runner
/entrypoint register --non-interactive --executor kubernetes

# Start the runner
/entrypoint run --user=gitlab-runner --working-directory=/home/gitlab-runner