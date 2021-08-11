#!/bin/sh -e

# Collects the last 3 minor versions of the GitLab chart and prints them to stdout.
# Can be used to update the CHART_VERSIONS file.

GITLAB_CHART="gitlab/gitlab"

scripts_dir="$(dirname "$0")"
. "${scripts_dir}/add_gitlab_repo.sh"

chart_versions() {
    # Find all the applicable charts and return a list of "app vers:chart vers" entries
    $HELM search repo ${GITLAB_CHART} -l 2>/dev/null | \
        awk -v CHART="${GITLAB_CHART}" '{if ( match($1, "^" CHART "$") ){ print $3 ":" $2}}'
}

previous_minor() {
    # this will subtract 1 from the minor version.
    # if minor-1 is -1 return previous major-1 with now minor
    echo "$1" | awk -F'.' '{if (($2 - 1) == "-1")
                              print ($1 - 1) ".";
                            else
                              print $1 "." ($2 - 1)}'
}

# Retrieve a list of GitLab charts and determine versions to fetch
target_versions=""
next_version=""
for version_pair in $(chart_versions); do
    # Unpack the versions in to variables
    IFS=':' read -r gitlab_version chart_version <<EOF
$version_pair
EOF

    # Pick the first chart version if nothing selected yet
    if [ -z "$target_versions" ]; then
        target_versions=$chart_version
        next_version=$(previous_minor "$gitlab_version")
        continue
    fi

    if [ -n "$(expr "$gitlab_version" : "\($next_version\)")" ]; then
        target_versions="${target_versions}:$chart_version"
        next_version=$(previous_minor "$gitlab_version")
    fi

    # Only need to target 3 versions
    # shellcheck disable=SC2046
    if [ $(echo "${target_versions}" | awk -F: '{print NF - 1}') -eq 2 ]; then
        break
    fi
done

# Update the CHART_VERSIONS file
rm CHART_VERSIONS
for version in $(echo "${target_versions}" | tr ':' ' '); do
  echo "${version}" >> CHART_VERSIONS
done
