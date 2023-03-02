#!/bin/bash

# This script generates the message to be placed in
# the description of a Release.
#
# This is meant to be used after the Release has been created via CI.
#
# It provides:
#   1. A link to compare the changes between the previous
#      version and the target version.
#   2. A table mapping each supported Chart version to its
#      corresponding GitLab version.
#
# Usage:
#   ./generate_release_message.sh <previous version> <target version>

set -eo pipefail

previousVersion="${1:-none}"
version="${2:-none}"

function usage() {
  if [ "${version}" = 'none' ] || [ "${previousVersion}" = 'none' ]; then
    echo 'Usage: ./script <previous version> <target version>'
    exit 1
  fi
}

function get_versions_from_message() {
  tag_message=$(git show "${version}" --quiet | grep 'supports GitLab Charts')

  echo "${tag_message}" | awk '{print $7 " " $8 " " $9}' | sed 's/,//g'
}

function find_gitlab_version() {
  chart_version="${1}"

  helm repo update > /dev/null 2>&1

  helm search repo gitlab/gitlab -l -o json \
    | jq -r --arg version "${chart_version}" \
        '.[] | select(.name=="gitlab/gitlab") | select(.version==$version) | .app_version'
}

function get_version_map() {
  for tag in $(get_versions_from_message); do
    gitlabVersion=$(find_gitlab_version "${tag}")
    printf "%s | %s\n" "${tag}" "${gitlabVersion}"
  done
}

function generate() {
  printf '
## Changelog
[Changes](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/compare/%s...%s?from_project_id=18899486)

## Version mapping
Chart version | GitLab version
-|-
%s
' "${previousVersion}" "${version}" "$(get_version_map)"
}

usage
generate
