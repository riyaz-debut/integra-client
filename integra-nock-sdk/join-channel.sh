# #!/bin/sh

# CORE_PEER_LOCALMSPID="$1"
# HOST="$2"
# PORT="$3"
CURRENT_CONFIG="$1"
MODIFY_JSON_PATH="$2"
CURRENT_CONFIG_PB="$3"
MODIFY_CONFIG_PATH_PB="$4"
CHANNEL="$5"

CONFIG_UPDATE_PB_PATH="$6"

CONFIG_UPDATE_JSON_PATH="$7"

ENVELOPE_JSON="$8"

ENVELOPE_PB_ANCHOR_TX="$9"





echo $ENVELOPE_JSON $ENVELOPE_PB > join-channel-files/check.txt




 configtxlator proto_encode --input "${CURRENT_CONFIG}" --type common.Config > ${CURRENT_CONFIG_PB}

 configtxlator proto_encode --input "${MODIFY_JSON_PATH}" --type common.Config > ${MODIFY_CONFIG_PATH_PB}

 configtxlator compute_update --channel_id "${CHANNEL}" --original ${CURRENT_CONFIG_PB} --updated ${MODIFY_CONFIG_PATH_PB} > ${CONFIG_UPDATE_PB_PATH}

 configtxlator proto_decode --input ${CONFIG_UPDATE_PB_PATH} --type common.ConfigUpdate > ${CONFIG_UPDATE_JSON_PATH}

 echo '{"payload":{"header":{"channel_header":{"channel_id":"'$CHANNEL'", "type":2}},"data":{"config_update":'$(cat join-channel-files/config_update.json)'}}}' | jq . > ${ENVELOPE_JSON}


  configtxlator proto_encode --input ${ENVELOPE_JSON} --type common.Envelope > ${ENVELOPE_PB_ANCHOR_TX}.pb



  configtxlator proto_encode --input ${ENVELOPE_JSON} --type common.Envelope > ${ENVELOPE_PB_ANCHOR_TX}.tx


# configtxlator proto_encode --input ${ENVELOPE_JSON} --type common.Envelope > "join-channel-files/orgAnchor.tx"








# # for std out
# jq -s '.[0] * {"channel_group":{"groups":{"Application":{"groups": {"'${ORG_NAME}'":.[1]}}}}}' ${SOURCE} ${ORG_SOURCE}


# # for out to file
# jq -s '.[0] * {"channel_group":{"groups":{"Application":{"groups": {"'${ORG_NAME}'":.[1]}}}}}' ${SOURCE} ${ORG_SOURCE} > ${FINAL}