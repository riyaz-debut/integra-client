#!/bin/sh

ORG_NAME="$1"
SOURCE="$2"
ORG_SOURCE="$3"
FINAL="$4"


# for std out
jq -s '.[0] * {"channel_group":{"groups":{"Application":{"groups": {"'${ORG_NAME}'":.[1]}}}}}' ${SOURCE} ${ORG_SOURCE}


# for out to file
jq -s '.[0] * {"channel_group":{"groups":{"Application":{"groups": {"'${ORG_NAME}'":.[1]}}}}}' ${SOURCE} ${ORG_SOURCE} > ${FINAL}