package regsdk

import (
	"log"
	config "integra-nock-sdk/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var sdk *fabsdk.FabricSDK
var configProvider = configImpl.FromFile(config.CONFIG_PROVIDER_PATH)
var mspClient *msp.Client

func RegSdk() (*resmgmt.Client, *fabsdk.FabricSDK, context.ClientProvider, *msp.Client, string) {
	var err error
	log.Println("Setup the SDK")
	sdk, err = fabsdk.New(configProvider)
	if err != nil {
		panic("error while initialising sdk configuration" + err.Error())
	}

	mspClient, err := msp.New(sdk.Context(), msp.WithCAInstance(config.CA_INSTANCE))
	if err != nil {
		panic("error while initialising sdk configuration" + err.Error())
	}
	log.Println("MSP CLIENT @@@@@@@@@@@@@", mspClient)

	err = mspClient.Enroll(config.ORG_ADMIN, msp.WithSecret(config.SECRET))
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("msp client created")

	// // create client for resource management in fabric
	log.Println("**** Setup resmgmt client for Orderer")
	clCtx := sdk.Context(fabsdk.WithUser(config.ORG_ADMIN), fabsdk.WithOrg(config.ORG_NAME))
	resmgmtClient, err := resmgmt.New(sdk.Context(fabsdk.WithUser(config.ORG_ADMIN), fabsdk.WithOrg(config.ORG_NAME)))
	if err != nil {
		panic(err)
	}

	log.Println("clctx", clCtx, "ctx", "prposed")
	log.Println("RESMGMT CLIENT ########################## ", resmgmtClient)
	log.Println("SDK ########################## ", sdk)

	user := config.ORG_ADMIN

	return resmgmtClient, sdk, clCtx, mspClient, user
}
