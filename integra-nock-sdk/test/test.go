package test

import (
	"log"
	"path/filepath"
	"time"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	lcpackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	// /"github.com/hyperledger/fabric/common/policydsl"
)

const (
	channelID = "mychannel"
	// orgName   = "Org1"
	// orgMsp    = "Org1MSP"
	orgMsp  = "Org2MSP"
	orgName = "Org2"
	peer2   = "peer0.org2.example.com"
	peer1   = "peer0.org1.example.com"
)

var (
	ccID = "example_coc"
)

func CreateCCLifecycle(orgResMgmt *resmgmt.Client, sdk *fabsdk.FabricSDK) {
	// Package cc
	duration := time.Duration(8) * time.Second
	label, ccPkg := packageCC()
	// packageID := "example_coc_1:e70924c48a015d1e879c8d827024f688575b479b9b7c50ee268b310b5cd1c73a" //lcpackager.ComputePackageID(label, ccPkg)
	packageID := lcpackager.ComputePackageID(label, ccPkg)

	log.Println(lcpackager.ComputePackageID(label, ccPkg))

	log.Println("Package ID %s", packageID)
	// Install cc
	installCC(label, ccPkg, orgResMgmt)

	time.Sleep(duration)
	// Get installed cc package
	getInstalledCCPackage(packageID, ccPkg, orgResMgmt)

	time.Sleep(duration)

	// Query installed cc
	queryInstalled(label, packageID, orgResMgmt)

	// time.Sleep(duration)

	// Approve cc
	approveCC(packageID, orgResMgmt)

	time.Sleep(duration)

	//Query approve cc
	queryApprovedCC(orgResMgmt)

	time.Sleep(duration)

	// Check commit readiness
	checkCCCommitReadiness(orgResMgmt)

	time.Sleep(duration)

	// Commit cc
	commitCC(orgResMgmt)

	time.Sleep(duration)
	log.Println("=========== Here ==========")
	// Query committed cc
	queryCommittedCC(orgResMgmt)

	time.Sleep(duration)
	// Init cc
	initCC(sdk)

}

func getLcDeployPath() string {
	const ccPath = "fixtures/testdata/go/src/github.com/example_cc"
	return filepath.Join(ccPath)
}

func packageCC() (string, []byte) {
	desc := &lcpackager.Descriptor{
		Path:  getLcDeployPath(),
		Type:  pb.ChaincodeSpec_GOLANG,
		Label: "example_coc_1",
	}
	ccPkg, err := lcpackager.NewCCPackage(desc)
	if err != nil {
		log.Fatalf(" Fatal : %v", err)
	}

	// _, err = generateTarGz([]*Descriptor{desc})
	// if err == nil {
	// 	t.Fatal("generateTarGz call failed")
	// }

	return desc.Label, ccPkg
}

func installCC(label string, ccPkg []byte, orgResMgmt *resmgmt.Client) {
	installCCReq := resmgmt.LifecycleInstallCCRequest{
		Label:   label,
		Package: ccPkg,
	}

	packageID := lcpackager.ComputePackageID(installCCReq.Label, installCCReq.Package)

	resp, err := orgResMgmt.LifecycleInstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Fatalf("installCC Fatal : %v", err)
	}
	log.Println("packageID %s", packageID)
	log.Println("installCC %v ", resp)
}

func getInstalledCCPackage(packageID string, ccPkg []byte, orgResMgmt *resmgmt.Client) {
	resp, err := orgResMgmt.LifecycleGetInstalledCCPackage(packageID, resmgmt.WithTargetEndpoints(peer2), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Fatalf("getInstalledCCPackage Fatal : %v", err)
	}
	log.Println("getInstalledCCPackage  %v ", resp[0])
}

func queryInstalled(label string, packageID string, orgResMgmt *resmgmt.Client) {
	resp, err := orgResMgmt.LifecycleQueryInstalledCC(resmgmt.WithTargetEndpoints(peer2), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Fatalf(" Fatal queryInstalled : %v", err)
	}

	log.Println("queryInstalled %s", resp[0].PackageID)
}

func approveCC(packageID string, orgResMgmt *resmgmt.Client) {
	// ccPolicy := policydsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})
	approveCCReq := resmgmt.LifecycleApproveCCRequest{
		Name:              ccID,
		Version:           "1",
		PackageID:         packageID,
		Sequence:          1,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		//SignaturePolicy:   ccPolicy,
		InitRequired: true,
	}

	txnID, err := orgResMgmt.LifecycleApproveCC(channelID, approveCCReq, resmgmt.WithTargetEndpoints(peer2), resmgmt.WithOrdererEndpoint("orderer.example.com"), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Fatalf(" approveCC Fatal : %v", err)
	}
	log.Println(" approveCC %v", txnID)
}

func queryApprovedCC(orgResMgmt *resmgmt.Client) {
	queryApprovedCCReq := resmgmt.LifecycleQueryApprovedCCRequest{
		Name:     ccID,
		Sequence: 1,
	}
	resp, err := orgResMgmt.LifecycleQueryApprovedCC(channelID, queryApprovedCCReq, resmgmt.WithTargetEndpoints(peer2), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Fatalf(" Fatal : %v", err)
	}

	log.Println("Navpreet Response", resp)
}

func checkCCCommitReadiness(orgResMgmt *resmgmt.Client) {
	// ccPolicy := policydsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})
	req := resmgmt.LifecycleCheckCCCommitReadinessRequest{
		Name:              ccID,
		Version:           "1",
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		//SignaturePolicy:   ccPolicy,
		Sequence:     1,
		InitRequired: true,
	}
	resp, err := orgResMgmt.LifecycleCheckCCCommitReadiness(channelID, req, resmgmt.WithTargetEndpoints(peer2), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Fatalf("checkCCCommitReadiness Fatal : %v", err)
	}

	log.Println(resp)

	// ccPolicy = policydsl.SignedByAnyMember([]string{"Org2MSP", "Org2MSP"})
	// req = resmgmt.LifecycleCheckCCCommitReadinessRequest{
	// 	Name:              ccID,
	// 	Version:           "1",
	// 	EndorsementPlugin: "escc",
	// 	ValidationPlugin:  "vscc",
	// 	SignaturePolicy:   ccPolicy,
	// 	Sequence:          1,
	// 	InitRequired:      true,
	// }
	// resp, err = orgResMgmt.LifecycleCheckCCCommitReadiness(channelID, req, resmgmt.WithTargetEndpoints(peer2), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	// if err != nil {
	// 	log.Fatalf(" Fatal : %v", err)
	// }

	// log.Println(resp)
}

func commitCC(orgResMgmt *resmgmt.Client) {
	// ccPolicy := policydsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})
	req := resmgmt.LifecycleCommitCCRequest{
		Name:              ccID,
		Version:           "1",
		Sequence:          1,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		// SignaturePolicy:   ccPolicy,
		InitRequired: true,
	}
	txnID, err := orgResMgmt.LifecycleCommitCC(channelID, req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithTargetEndpoints(peer1, peer2), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		log.Fatalf("commitCC Fatal : %v", err)
	}

	log.Println(txnID)
}

func queryCommittedCC(orgResMgmt *resmgmt.Client) {
	req := resmgmt.LifecycleQueryCommittedCCRequest{
		Name: ccID,
	}
	resp, err := orgResMgmt.LifecycleQueryCommittedCC(channelID, req, resmgmt.WithTargetEndpoints(peer2), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Fatalf(" Fatal : %v", err)
	}

	log.Println(resp)
}

func initCC(sdk *fabsdk.FabricSDK) {
	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser("User1"), fabsdk.WithOrg(orgName))
	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)
	if err != nil {
		log.Fatalf("Failed to create new channel client: %s", err)
	}

	// init
	var initArgs = [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte("200")}

	_, err = client.Execute(channel.Request{ChaincodeID: ccID, Fcn: "init", Args: initArgs, IsInit: true},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		log.Fatalf("Failed to init: %s", err)
	}

}

// var c * gin.Context
// router := gin.Default()
// router.GET("/", createCCLifecycle)
// router.Run("localhost:4000")
