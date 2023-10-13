package client_route

import (
	"log"
	"path/filepath"

	clientController "integra-nock-sdk/client/client_controller"
	reg "integra-nock-sdk/regsdk"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	lcpackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"

	// configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/gin-gonic/gin"
	_ "github.com/spacemonkeygo/openssl"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var chConfigPath = "/home/riyaz/projects/integra/integra-client-backend/integra-nock-sdk/downloadFiles/envelopConfig.pb"

var resmgmtClient, sdkreg, clCtx, mspClient, user = reg.RegSdk()

//register new user route
func Login(c *gin.Context) {
	response := clientController.UserLogin(c, resmgmtClient, clCtx, mspClient, user)
	log.Println("response in login client route: ", response.Data)
	c.JSON(response.Status, response)
}

//get logged in user details
func UserDashboard(c *gin.Context) {
	log.Println("conetxft of client route of userDashboard fx: ", c)
	response := clientController.CurrentUser(c)
	log.Println("response in login client route: ", response.Data)
	c.JSON(response.Status, response)
}

//get logged in user details CheckApi
func TestApi(c *gin.Context) {
	log.Println("in testApi fx: ", c)
	response := "sucessfully run the 1st test api"
	c.JSON(200, response)
}

//get logged in user details CheckApi
func CheckApi(c *gin.Context) {
	log.Println("in checkApi fx: ", c)
	response := "sucessfully run the 2nd test api"
	c.JSON(200, response)
}

//install and approve chaincode function
func ChaincodeInstall(c *gin.Context) {
	label, ccPkg := packageCC()
	response := clientController.InstallCC(label, ccPkg, resmgmtClient, c)
	c.IndentedJSON(response.Status, response)
}

// get list of chaincode installed
func ChaincodeList(c *gin.Context) {
	response := clientController.ListChaincodes(resmgmtClient)
	c.IndentedJSON(response.Status, response)
}

func packageCC() (string, []byte) {
	desc := &lcpackager.Descriptor{
		Path: getLcDeployPath(),
		Type: pb.ChaincodeSpec_GOLANG,

		Label: "example_cc",
	}
	ccPkg, err := lcpackager.NewCCPackage(desc)
	if err != nil {
		log.Fatalf(" Fatal : %v", err)
	}
	return desc.Label, ccPkg
}

func getLcDeployPath() string {
	const ccPath = "fixtures/testdata/go/src/github.com/example_cc"
	return filepath.Join(ccPath)
}

//get chaincode checkupdate by id
func ChaincodeUpdateCheck(c *gin.Context) {
	response := clientController.CcUpdateCheck()
	log.Println("chaincode info is ", response)
	c.IndentedJSON(response.Status, response)

}

//install update
func InstallUpdate(c *gin.Context) {
	id := c.Param("chaincodeid")
	response := clientController.UpdateInstallation(c, resmgmtClient, id)
	log.Println("chaincode info is ", response)
	c.IndentedJSON(response.Status, response)

}

// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ Organization @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

//func get list of Org except the login one
func ListOrganizations(c *gin.Context) {
	response := clientController.GetOrganizations()
	c.IndentedJSON(response.Status, response)

}

//createSignature
func SignOrganization(c *gin.Context) {
	//fetching id of org to be signed from params
	id := c.Param("orgid")
	response := clientController.CreateSignature(resmgmtClient, chConfigPath, mspClient, user, id)
	c.IndentedJSON(response.Status, response)
}
