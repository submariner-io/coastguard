#!/bin/bash
set -e -x

trap "exit 1" SIGTERM SIGINT

if [ "${COASTGUARD_DEBUG}" == "true" ]; then
    DEBUG="-v=9"
else
    DEBUG=""
fi

exec coastguard-controller ${DEBUG} -alsologtostderr
