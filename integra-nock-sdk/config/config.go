package config

const (

	// Common for both user and organizations
	
	CHECK_FOR_UPDATES      = "http://localhost:5000/chaincode/checkforupdates"
	CHAINCODE_LOGS         = "http://localhost:5000/chaincode/logs"
	INSTALL_CHAINCODE_LIST = "http://localhost:5000/chaincode/list"
	CHAINCODE_UPDATE       = "http://localhost:5000/chaincode/update"
	ORG_LIST               = "http://localhost:5000/organization/list"
	ORG                    = "http://localhost:5000/single/organization"
	GET_MODIFIED_CONFIG    = "http://localhost:5000/organization/modifiedconfig"
	SAVE_ORG_SIGN          = "http://localhost:5000/organization/savesign"
	LOGIN                  = "http://localhost:5000/user/login"
	USER_DASHBOARD         = "http://localhost:5000/user/home"
	JOIN_STATUS_UPDATE		   = "http://localhost:5000/organization/join/status_update"
	SIGN_STATUS_UPDATE		   = "http://localhost:5000/organization/sign/status_update"

	ORDERER_ENDPOINT = "orderer.example.com"
	CHANNEL_ID = "mychannel"
	CC_NAME = "example_cc"

	// ========== env variables for user1 and organization 1 ================
	CONFIG_PROVIDER_PATH = "./connection-org1.yaml"
	ORG_NAME = "Org1"
	PEER    = "peer0.org1.example.com"
	PEER_NAME = PEER

	CA_INSTANCE = "ca.org1.example.com"
	ORG_ADMIN = "org1admin"
	SECRET = "org1adminpw"

	// ========== env variables for user2 and organization 2 ================
	// CONFIG_PROVIDER_PATH = "./connection-org1.yaml"
	// ORG_NAME = "Org2"
	// PEER    = "peer0.org2.example.com"
	// PEER_NAME = PEER

	// CA_INSTANCE = "ca.org2.example.com"
	// ORG_ADMIN = "org2admin"
	// SECRET = "org2adminpw"

	// ========== env variables for user3 and organization 3 ================
	HOST = "peer0.org3.example.com"
	PORT = "11051"
	CORE_PEER_LOCALMSPID = "Org3MSP"

	// CONFIG_PROVIDER_PATH = "./connection-org3.yaml"
	// PEER = "peer0.org3.example.com"
	// ORG_NAME = "Org3"
	// PEER_NAME = PEER
	// CA_INSTANCE = "ca.org3.example.com"
	// ORG_ADMIN = "org3admin"
	// SECRET = "org3adminpw"

	
)
