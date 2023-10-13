# #!/bin/sh

CORE_PEER_LOCALMSPID="$1"
HOST="$2"
PORT="$3"
CURRENT_CONFIG="$4"
MODIFY_JSON_PATH="$5"
# CURRENT_CONFIG_PB="$6"
# MODIFY_CONFIG_PATH_PB="$7"
# CHANNEL="$8"

# CONFIG_UPDATE_PB_PATH="$9"

# CONFIG_UPDATE_JSON_PATH="$10"







jq '.channel_group.groups.Application.groups.'${CORE_PEER_LOCALMSPID}'.values += {"AnchorPeers":{"mod_policy": "Admins","value":{"anchor_peers": [{"host": "'$HOST'","port": '$PORT'}]},"version": "0"}}' ${CURRENT_CONFIG} > ${MODIFY_JSON_PATH}















# # for std out
# jq -s '.[0] * {"channel_group":{"groups":{"Application":{"groups": {"'${ORG_NAME}'":.[1]}}}}}' ${SOURCE} ${ORG_SOURCE}


# # for out to file
# jq -s '.[0] * {"channel_group":{"groups":{"Application":{"groups": {"'${ORG_NAME}'":.[1]}}}}}' ${SOURCE} ${ORG_SOURCE} > ${FINAL}