#!/bin/bash
set -e
MAX_REGISTER_ATTEMPTS=30

for i in $(seq 1 "${MAX_REGISTER_ATTEMPTS}"); do
    echo "Registration attempt ${i} of ${MAX_REGISTER_ATTEMPTS}"
    /entrypoint register --non-interactive

    retval=$?

    if [ ${retval} = 0 ]; then
    break
    elif [ ${i} = ${MAX_REGISTER_ATTEMPTS} ]; then
    exit 1
    fi

    sleep 5 
done

exit 0
