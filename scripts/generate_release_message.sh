#!/bin/sh

# This script generates the message to be placed in
# the description of a Release.
#
# It provides:
#   1. A link to the changelog.
#   2. A table mapping each supported Chart version to its
#      corresponding GitLab version.
#
# Dependencies:
# - git
# - helm
#
# Usage:
#   ./generate_release_message.sh <target version>

set -e

version="${1:-none}"

usage() {
  if [ "${version}" = 'none' ]; then
    echo 'Usage: ./script <target version>'
    exit 1
  fi
}

get_versions_from_message() {
  tag_message=$(git show "${version}" --quiet | grep 'supports GitLab Charts')

  echo "${tag_message}" | awk '{print $7 " " $8 " " $9}' | sed 's/,//g'
}

find_gitlab_version() {
  chart_version="${1}"

  if [ -f "./charts/gitlab-${chart_version}.tgz" ]; then
    tar -Oxf "./charts/gitlab-${chart_version}.tgz" 'gitlab/Chart.yaml' | yq eval '.appVersion' -
  else
    helm search repo gitlab/gitlab -l -o table \
      | grep "${chart_version}" \
      | awk '{print $3}'
  fi
}

get_version_map() {
  helm repo update > /dev/null 2>&1

  for tag in $(get_versions_from_message); do
    gitlabVersion=$(find_gitlab_version "${tag}")
    printf "%s | %s\n" "${tag}" "${gitlabVersion}"
  done
}

generate() {
  printf '
## Changelog
[CHANGELOG.md](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/blob/master/CHANGELOG.md)

## Version mapping
Chart version | GitLab version
-|-
%s
' "$(get_version_map)"
}

usage
generate
