#!/bin/bash
# Usage: scripts/tag_release.sh '<version>'

VERSION=$1
CHART_VERSIONS=$(paste -sd, CHART_VERSIONS)
MSG="Version ${VERSION} - supports GitLab Charts ${CHART_VERSIONS//,/, }"

git tag "${VERSION}" -m \"${MSG}\"

echo "tag ${VERSION} created, finish by pushing: 'git push origin ${VERSION}'"
