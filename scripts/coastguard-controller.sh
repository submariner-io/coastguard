#!/bin/bash
set -e -x

trap "exit 1" SIGTERM SIGINT

if [ "${COASTGUARD_DEBUG}" == "true" ]; then
    DEBUG="--debug -v=9"
else
    DEBUG="-v=4"
fi

exec coastguard-controller ${DEBUG} -alsologtostderr