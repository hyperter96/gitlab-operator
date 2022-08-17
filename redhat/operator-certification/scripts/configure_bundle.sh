#!/bin/sh

set -eu

OPENSHIFT_MIN=${OPENSHIFT_MIN:-"4.6"}
OPENSHIFT_MAX=${OPENSHIFT_MAX:-"4.10"}

BUNDLE_DIR=${BUNDLE_DIR:-"."}

BUNDLE_DIR=$(realpath ${BUNDLE_DIR})

YQ=${YQ:-yq}


adjust_annotations(){
    local version_range="v${OPENSHIFT_MIN}-v${OPENSHIFT_MAX}"
    ${YQ} eval -i '.annotations["com.redhat.openshift.versions"]="'$version_range'"' ${BUNDLE_DIR}/metadata/annotations.yaml
}

adjust_csv(){
    local csv_files=$(grep -l 'kind: ClusterServiceVersion' ${BUNDLE_DIR}/manifests/*.yaml)
    for csv in $csv_files
    do
        ${YQ} eval -i '.metadata.annotations["olm.properties"]="[{\"type\": \"olm.maxOpenShiftVersion\", \"value\": \"'${OPENSHIFT_MAX}'\"}]"' $csv
    done
}

for cmd in $@
do
    $cmd 
done
