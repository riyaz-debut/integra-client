package client_contoller

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	config "integra-nock-sdk/config"
	"integra-nock-sdk/utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/protolator"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"

	// "os from urllib.parse import urlparse"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"

	// "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	fabAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"

	// configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	lcpackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var userID uint
var userToken string
var orgId int
var orgName string
var orgMsp string

// =========================== USER PART ==============================
// /login with user credentials
func UserLogin(context *gin.Context, resmgmtClient *resmgmt.Client, clCtx contextApi.ClientProvider, mspClient *msp.Client, userName string) utils.Response {
	var input LoginInput
	log.Println("input of login client controller: ", input)
	if err := context.ShouldBindJSON(&input); err != nil {
		log.Println("getting error in mapping json data to input variable")
		response := utils.Response{
			Status:  422,
			Message: "Invalid credentials",
			Err:     err,
		}
		return response
	}

	user := User{}

	user.UserName = input.UserName
	user.Password = input.Password
	log.Println("user of login client controller: ", user)

	userBuffer, _ := json.Marshal(user)
	log.Println("userBuffer of login client controller: ", userBuffer)

	//calling sub api
	apiResp, err := ApiCall("POST", config.LOGIN, userBuffer, "")
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "Internal server error",
			Err:     err,
		}
		return response

	}

	respData, _ := ioutil.ReadAll(apiResp.Body)

	type Organizations struct {
		Id             int    `gorm:"primaryKey;autoIncrement"`
		Name           string `json:"name"`
		MspId          string `json:"msp_id"`
		PeersCount     int    `json:"peers_count"`
		Config         string `json:"file" gorm:"type:text"`
		ModifiedConfig string `json:"modified_config" gorm:"type:text"`
		Join_Status    int    `json:"join_status"`
		CreatedAt      string `json:"created_at"`
		UpdatedAt      string `json:"updated_at" gorm:"autoUpdateTime"`
	}

	type ApiResponse struct {
		Status  int           `json:"status,omitempty"`
		Message string        `json:"message,omitempty"`
		Data    string        `json:"data,omitempty"`
		Role    string        `json:"role,omitempty"`
		OrgData Organizations `json:"org_data,omitempty"`
		Err     error         `json:"err,omitempty"`
	}

	var apiResponse ApiResponse
	err = json.Unmarshal(respData, &apiResponse)
	if err != nil {
		log.Println("error unmarshalling client side login info", err)
		response := utils.Response{
			Status:  500,
			Message: "Internal server error",
			Err:     err,
		}
		log.Println("response", response)
		return response
	}

	userToken = apiResponse.Data
	fmt.Println("login response Body:", string(respData))
	log.Println("join sttaus :", apiResponse.OrgData.Join_Status)

	joinStatus := apiResponse.OrgData.Join_Status

	id := apiResponse.OrgData.Id
	log.Println("id in User login fx :", id)

	if joinStatus == 0 {
		log.Println("Need to add sign org")
	} else if joinStatus == 1 {
		log.Println("Signed need to save channel config")
	} else if joinStatus == 2 {
		log.Println("ready to join")
		JoinChannel(id, resmgmtClient, clCtx, mspClient, userName)
	} else {
		log.Println("Successfully joined cahnnel")
	}

	response := utils.Response{
		Status:  200,
		Message: apiResponse.Message,
		Data:    apiResponse.Data,
		Role:    apiResponse.Role,
		OrgData: apiResponse.OrgData,
	}
	return response
}

//user home dashboard
func CurrentUser(body *gin.Context) utils.Response {

	var byte []byte

	apiResp, err := ApiCall("GET", config.USER_DASHBOARD, byte, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}

	respData, _ := ioutil.ReadAll(apiResp.Body)
	if err != nil {
		fmt.Println("Can't readAll resp body", err)
	}

	type User struct {
		Id       uint   `json:"id"`
		UserName string `json:"user_name"`
		OrgId    int    `json:"org_id"`
		OrgName  string `json:"org_name"`
		OrgMsp   string `json:"org_msp"`
	}

	type ApiResponse struct {
		Status  int    `json:"status,omitempty"`
		Message string `json:"message,omitempty"`
		Data    User   `json:"data,omitempty"`
		Err     error  `json:"err,omitempty"`
	}
	var apiResponse ApiResponse

	err = json.Unmarshal(respData, &apiResponse)
	if err != nil {
		log.Println("error unmarshalling ", err)
		response := utils.Response{Status: 500, Message: "error unmarshalling modified config", Err: err}
		log.Println("response", response)
		return response
	}

	//setting user id and org id from current user
	orgId = int(uint(apiResponse.Data.OrgId))

	orgName = apiResponse.Data.OrgName
	orgMsp = apiResponse.Data.OrgMsp
	// orgId = int(orgId)
	userID = apiResponse.Data.Id

	response := utils.Response{
		Status:  apiResponse.Status,
		Message: apiResponse.Message,
		Data:    apiResponse.Data,
		Err:     apiResponse.Err,
	}
	return response

}

// ============================ CHAINCODE PART ========================

type chaincode struct {
	Id       int    `json:"id"`
	CcName   string `json:"name"`
	Label    string `json:"label"`
	Version  string `json:"version"`
	Sequence int    `json:"sequence"`
	OrgName  string `json:"org_name"`
	OrgId    int    `json:"org_id"`
	OrgMsp   string `json:"msp_id"`
}

type ChaincodeLog struct {
	Id        int       `json:"id" gorm:"primary key;autoincrement"`
	CcId      int       `json:"cc_id"`
	Name      string    `json:"name"`
	Label     string    `json:"label"`
	Version   string    `json:"version"`
	Sequence  int       `json:"sequence"`
	OrgId     int       `json:"org_id"`
	OrgName   string    `json:"org_name"`
	OrgMsp    string    `json:"msp_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ChaincodeCommits struct {
	Id int `json:"id" gorm:"primary key;autoincrement"`
	// ccID int       `json:"cc_id"`
	Name      string    `json:"name"`
	Label     string    `json:"label"`
	Version   string    `json:"version"`
	Sequence  int       `json:"sequence"`
	Status    int       `json:"status,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"autoUpdateTime"`
}

func InstallCC(label string, ccPkg []byte, orgResMgmt *resmgmt.Client, body *gin.Context) utils.Response {
	log.Println("logged in org id:", orgId, "org_name:", orgName, "org_msp:", orgMsp)
	version := "1"
	sequence := 1
	chaincodeID := 1

	duration := time.Duration(8) * time.Second
	// Getting payload obiects
	var chaincodes chaincode

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := body.BindJSON(&chaincodes); err != nil {

	} else {
		log.Println("payload from backend  @@@@@@@@@@@@@@@", chaincodes)

	}

	installCCReq := resmgmt.LifecycleInstallCCRequest{
		Label:   label,
		Package: ccPkg,
	}
	packageID := lcpackager.ComputePackageID(installCCReq.Label, installCCReq.Package)

	resp, err := orgResMgmt.LifecycleInstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Println("installCC Fatal :", err)

		response := utils.Response{
			Status:  500,
			Message: "error installing chaincode",
			Err:     err,
		}
		return response
	}
	log.Println("packageID   @@@@@@@@@@@@ ", packageID)
	log.Println("installCC ##############  ", resp)

	time.Sleep(duration)

	//getInstalledCCPackage
	log.Println("######################################")
	log.Println("getInstalledCCPackage")
	log.Println("######################################")

	respGet, err := orgResMgmt.LifecycleGetInstalledCCPackage(packageID, resmgmt.WithTargetEndpoints(config.PEER_NAME), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Println("getInstalledCCPackage Fatal @@@@@@@@@@@@@@@@@@@@@ :", err)

		response := utils.Response{
			Status:  404,
			Message: "error getting installed package",
			Err:     err,
		}
		return response
	}
	log.Println("getInstalledCCPackage   @@@@@@@@@@@@@@@@@@@@@@@  ", respGet[0])

	time.Sleep(duration)

	// Query installed cc

	log.Println("######################################")
	log.Println("Query installed cc")
	log.Println("######################################")

	respQuery, err := orgResMgmt.LifecycleQueryInstalledCC(resmgmt.WithTargetEndpoints(config.PEER_NAME), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Println(" Fatal queryInstalled @@@@@@@@@@@@@@@@@@: ", err)
		response := utils.Response{
			Status:  500,
			Message: "error querying installed package",
			Err:     err,
		}
		return response
	}

	log.Println("queryInstalled @@@@@@@@@@@@@@@@@@@@@@ ", respQuery[0].PackageID)

	log.Println("######################################")
	log.Println("Approve cc")
	log.Println("######################################")

	approveCCReq := resmgmt.LifecycleApproveCCRequest{
		Name:              config.CC_NAME,
		Version:           version,
		PackageID:         packageID,
		Sequence:          int64(sequence),
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		InitRequired:      true,
	}

	txnID, err := orgResMgmt.LifecycleApproveCC(config.CHANNEL_ID, approveCCReq, resmgmt.WithTargetEndpoints(config.PEER_NAME), resmgmt.WithOrdererEndpoint("orderer.example.com"), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Println(" approveCC Fatal : ", err)
		response := utils.Response{
			Status:  500,
			Message: "error approving installed package",
			Err:     err,
		}
		return response
	}
	log.Println(" approveCC @@@@@@@@@@@@@@@@ ", txnID)

	time.Sleep(duration)

	log.Println("######################################")
	log.Println("queryApprovedCC")
	log.Println("######################################")

	queryApprovedCCReq := resmgmt.LifecycleQueryApprovedCCRequest{
		Name:     config.CC_NAME,
		Sequence: approveCCReq.Sequence,
	}
	respApprove, err := orgResMgmt.LifecycleQueryApprovedCC(config.CHANNEL_ID, queryApprovedCCReq, resmgmt.WithTargetEndpoints(config.PEER_NAME), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Println(" Fatal :", err)
		response := utils.Response{
			Status:  500,
			Message: "error querying approved installed package",
			Err:     err,
		}
		return response
	}

	log.Println("Response @@@@@@@@@@@@@@@@", respApprove)

	//chaincode params into struct
	chaincodeData := &chaincode{Id: chaincodeID, CcName: config.CC_NAME, Label: label, Version: version, Sequence: sequence, OrgName: orgName, OrgId: orgId, OrgMsp: orgMsp}

	chaincodeBuffer, _ := json.Marshal(chaincodeData)

	//calling api to store chaincode commit status in db

	apiResp, err := ApiCall("POST", config.CHAINCODE_LOGS, chaincodeBuffer, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}

	//fetching response from calling api and unmarshalling
	respData, _ := ioutil.ReadAll(apiResp.Body)
	log.Println("body received ", string(respData))

	type ApiResponse struct {
		Status  int          `json:"status,omitempty"`
		Message string       `json:"message,omitempty"`
		Data    ChaincodeLog `json:"data,omitempty"`
		Err     error        `json:"err,omitempty"`
	}

	var apiResponse ApiResponse
	err = json.Unmarshal(respData, &apiResponse)
	if err != nil {
		log.Println("error unmarshalling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error unmarhsalling ",
			Err:     err,
		}
		log.Println("response", response)
		return response
	}

	fmt.Println("response Body:", string(respData))

	response := utils.Response{
		Status:  apiResponse.Status,
		Message: apiResponse.Message,
		Data:    apiResponse.Data,
		Err:     apiResponse.Err,
	}
	return response

}

func ListChaincodes(orgResMgmt *resmgmt.Client) utils.Response {
	log.Println("######################################")

	log.Println("Query installed cc")
	log.Println("######################################")

	// type GetOrgId struct {
	// 	OrgId int `json:"org_id"`
	// }
	var idByte []byte

	// orgId := GetOrgId{OrgId: orgId}

	// orgData, _ := json.Marshal(orgId)

	//calling api to get list of org from db
	apiResp, err := ApiCall("POST", config.INSTALL_CHAINCODE_LIST, idByte, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}

	log.Println("response from chaincode/status from admin", apiResp)
	respData, _ := ioutil.ReadAll(apiResp.Body)
	if err != nil {
		fmt.Println("Can't readAll resp body", err)
		response := utils.Response{
			Status:  500,
			Message: "error converting response data to bytes array",
			// Data: ,
			Err: err,
		}
		return response
	}

	type ChaincodeLists struct {
		Id        int    `json:"id"`
		CC_ID     int    `json:"cc_id"`
		Name      string `json:"name"`
		Label     string `json:"label"`
		Version   string `json:"version"`
		Sequence  int    `json:"sequence"`
		Status    int    `json:"status"`
		Url       string `json:"url"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	//fetching called api response into
	type Data struct {
		Status  int              `json:"status"`
		Message string           `json:"message"`
		Data    []ChaincodeLists `json:"data"`
		Err     error            `json:"err,omitempty"`
	}

	//unmarshalling resp data
	var chaincodeList Data
	err = json.Unmarshal(respData, &chaincodeList)
	if err != nil {
		log.Println("error unmarshalling ", err)
		response := utils.Response{Status: 500, Message: "error unmarshalling organizations data", Err: err}
		log.Println("response", response)
		return response
	}

	response := utils.Response{Status: chaincodeList.Status, Message: chaincodeList.Message, Data: chaincodeList.Data, Err: chaincodeList.Err}
	log.Println("response", response)
	return response

}

// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

//for checking whether chaincode update is available or not for particular organization

func CcUpdateCheck() utils.Response {
	log.Println("token for user ", userToken)

	//org id into struct
	type OrgInfo struct {
		OrgId int `json:"org_id"`
	}

	orginfo := &OrgInfo{OrgId: orgId}

	//marshalling org id
	orgIdByte, _ := json.Marshal(orginfo)

	apiResp, err := ApiCall("POST", config.CHECK_FOR_UPDATES, orgIdByte, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}

	respData, _ := ioutil.ReadAll(apiResp.Body)

	fmt.Println("response Body:", string(respData))

	//struct to receieve response data from called api
	type ApiData struct {
		Status  int         `json:"status,omitempty"`
		Message string      `json:"message,omitempty"`
		Data    interface{} `json:"data,omitempty"`
		Err     error       `json:"err,omitempty"`
	}

	//unmarshalling api resp data
	var apiresponse ApiData
	err = json.Unmarshal(respData, &apiresponse)
	if err != nil {
		log.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		log.Println("error unmarshalling ", err)
		log.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")

		response := utils.Response{Status: 500, Message: "error unmarshalling response", Err: err}

		log.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		log.Println("response", response)
		log.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		return response
	}

	response := utils.Response{
		Status:  apiresponse.Status,
		Message: apiresponse.Message,
		Data:    apiresponse.Data,
		Err:     apiresponse.Err,
	}
	return response

}

//function for installing chaincode  update
func UpdateInstallation(body *gin.Context, orgResMgmt *resmgmt.Client, chaincodeid string) utils.Response {

	//converting org id into int
	chaincodeId, err := strconv.Atoi(chaincodeid)
	if err != nil {
		response := utils.Response{Status: 500, Message: "error converting org id to int", Err: err}
		return response
	}

	//adding orgid struct
	type ChaincodeInfo struct {
		ChaincodeID int `json:"cc_id"`
	}

	//adding org id into struct
	chaincodeIdStruct := &ChaincodeInfo{ChaincodeID: chaincodeId}

	//marshalling data
	chaincodeIdByte, _ := json.Marshal(chaincodeIdStruct)

	//calling sub api
	apiResp, err := ApiCall("POST", config.CHAINCODE_UPDATE, chaincodeIdByte, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}

	// struct for fetching update data from called api
	type ChaincodeUpdates struct {
		Id        int    `json:"id"`
		CC_ID     int    `json:"cc_id"`
		Name      string `json:"name"`
		Label     string `json:"label"`
		Version   string `json:"version"`
		Sequence  int    `json:"sequence"`
		Status    int    `json:"status"`
		Url       string `json:"url"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	// receiving api response data
	type Data struct {
		Status  int              `json:"status,omitempty"`
		Message string           `json:"message,omitempty"`
		Data    ChaincodeUpdates `json:"data,omitempty"`
		Err     error            `json:"err,omitempty"`
	}

	respData, err := ioutil.ReadAll(apiResp.Body)
	if err != nil {
		fmt.Println("Can't readAll resp body", err)
	}

	//unmarshalling resp data
	var chaincode_updates Data
	err = json.Unmarshal(respData, &chaincode_updates)

	if err != nil {
		fmt.Println("Can't unmarshal the byte array", err)
	}

	//new update paarmeters

	NewCcName := chaincode_updates.Data.Name
	NewCcVersion := chaincode_updates.Data.Version
	NewCcSequence := chaincode_updates.Data.Sequence
	NewCcLabel := chaincode_updates.Data.Label

	if _, err := os.Stat("public/chaincodes"); err == nil {
		fmt.Printf("File exists\n")
	} else {
		// create directory
		if err := os.MkdirAll("public/chaincodes", os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	//connvert chaincode_id to int
	fileUrl := chaincode_updates.Data.Url
	filename := filepath.Join("public/chaincodes/" + chaincode_updates.Data.Name + "_" + chaincode_updates.Data.Version)
	//chaincode zipfile extraction path
	forDest := "public/chaincodes"
	zipExtractPath := filepath.Join(forDest)
	log.Println("zipeextractpath", zipExtractPath)
	GetZipFile(filename, fileUrl, zipExtractPath)

	log.Println("chaincode-name", chaincode_updates.Data.Name)

	//getting new chaincode extracted path
	extractedCcPath := "public/chaincodes/" + chaincode_updates.Data.Name

	log.Println("checking after extraction")

	NewCclabel, NewCcPkg := packageCC(extractedCcPath, NewCcLabel)

	duration := time.Duration(8) * time.Second

	installCCReq := resmgmt.LifecycleInstallCCRequest{
		Label:   NewCclabel,
		Package: NewCcPkg,
	}
	packageID := lcpackager.ComputePackageID(installCCReq.Label, installCCReq.Package)
	resp, err := orgResMgmt.LifecycleInstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Println("installCC Fatal :", err)

		response := utils.Response{
			Status:  500,
			Message: "error installing chaincode",
			Err:     err,
		}
		return response
	}
	log.Println("packageID   @@@@@@@@@@@@ ", packageID)
	log.Println("installCC ##############  ", resp)

	time.Sleep(duration)

	//getInstalledCCPackage
	log.Println("######################################")
	log.Println("getInstalledCCPackage")
	log.Println("######################################")

	log.Println("peer name :", config.PEER_NAME)
	respGet, err := orgResMgmt.LifecycleGetInstalledCCPackage(packageID, resmgmt.WithTargetEndpoints(config.PEER_NAME), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Println("getInstalledCCPackage Fatal @@@@@@@@@@@@@@@@@@@@@ :", err)

		response := utils.Response{
			Status:  404,
			Message: "error getting installed package",
			Err:     err,
		}
		return response
	}
	log.Println("getInstalledCCPackage   @@@@@@@@@@@@@@@@@@@@@@@ ", respGet[0])

	time.Sleep(duration)
	// Query installed cc

	log.Println("######################################")
	log.Println("Query installed cc")
	log.Println("######################################")

	respQuery, err := orgResMgmt.LifecycleQueryInstalledCC(resmgmt.WithTargetEndpoints(config.PEER_NAME), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Println(" Fatal queryInstalled @@@@@@@@@@@@@@@@@@:", err)
		response := utils.Response{
			Status:  500,
			Message: "error querying installed package",
			Err:     err,
		}
		return response
	}

	log.Println("queryInstalled @@@@@@@@@@@@@@@@@@@@@@ ", respQuery[0].PackageID)

	log.Println("######################################")
	log.Println("Approve cc")
	log.Println("######################################")

	approveCCReq := resmgmt.LifecycleApproveCCRequest{
		Name:              NewCcName,
		Version:           NewCcVersion,
		PackageID:         packageID,
		Sequence:          int64(NewCcSequence),
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		InitRequired:      true,
	}

	txnID, err := orgResMgmt.LifecycleApproveCC(config.CHANNEL_ID, approveCCReq, resmgmt.WithTargetEndpoints(config.PEER_NAME), resmgmt.WithOrdererEndpoint("orderer.example.com"), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Println(" approveCC Fatal : ", err)
		response := utils.Response{
			Status:  500,
			Message: "error approving installed package",
			Err:     err,
		}
		return response
	}
	log.Println(" approveCC @@@@@@@@@@@@@@@@ ", txnID)

	time.Sleep(duration)

	log.Println("######################################")
	log.Println("queryApprovedCC")
	log.Println("######################################")

	queryApprovedCCReq := resmgmt.LifecycleQueryApprovedCCRequest{
		Name:     NewCcName,
		Sequence: approveCCReq.Sequence,
	}
	respApprove, err := orgResMgmt.LifecycleQueryApprovedCC(config.CHANNEL_ID, queryApprovedCCReq, resmgmt.WithTargetEndpoints(config.PEER_NAME), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		log.Println(" Fatal :", err)
		response := utils.Response{
			Status:  500,
			Message: "error querying approved installed package",
			Err:     err,
		}
		return response
	}

	log.Println("Response @@@@@@@@@@@@@@@@", respApprove)

	chaincodeData := &chaincode{Id: chaincodeId, CcName: NewCcName, Label: NewCcLabel, Version: NewCcVersion, Sequence: NewCcSequence, OrgName: orgName, OrgId: orgId, OrgMsp: orgMsp}

	chaincodeBuffer, _ := json.Marshal(chaincodeData)

	apiResp, err = ApiCall("POST", config.CHAINCODE_LOGS, chaincodeBuffer, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}

	log.Println("response from chaincode/status from admin", apiResp)
	respData, _ = ioutil.ReadAll(apiResp.Body)

	fmt.Println("response Body:", string(respData))

	type ApiResponse struct {
		Status  int              `json:"status,omitempty"`
		Message string           `json:"message,omitempty"`
		Data    ChaincodeUpdates `json:"data,omitempty"`
		Err     error            `json:"err,omitempty"`
	}

	var apiResponse ApiResponse
	err = json.Unmarshal(respData, &apiResponse)
	if err != nil {
		log.Println("error unmarshalling ", err)
		response := utils.Response{Status: 500, Message: "error unmarshalling api", Err: err}
		log.Println("response", response)
		return response
	}

	response := utils.Response{
		Status:  apiResponse.Status,
		Message: apiResponse.Message,
		Data:    apiResponse.Data,
		Err:     apiResponse.Err,
	}

	return response

}

//get desired chaincode path
func getLcDeployPath(cc_path string) string {
	ccPath := cc_path
	return filepath.Join(ccPath)
}

//packaging chaincode
func packageCC(cc_path string, label string) (string, []byte) {
	desc := &lcpackager.Descriptor{
		Path: getLcDeployPath(cc_path),
		Type: pb.ChaincodeSpec_GOLANG,

		Label: label,
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

// =================================== ORGANIZATION PART =============================================
// use this method
func SignByOrg(resmgmtClient *resmgmt.Client, clCtx contextApi.ClientProvider, body *gin.Context, id string) utils.Response {

	//for later
	//need logged in org id from logged in details later for dynamic logged in user id
	loggedInId := orgId

	//org_Id of new org to be signed
	var org_Id int
	if i, err := strconv.Atoi(id); err == nil {
		fmt.Printf("i=%d, type: %T\n", i, i)
		org_Id = i
	} else {
		response := utils.Response{Status: 500, Message: "error converting org id to int", Err: err}
		log.Println("response", response)
		return response
	}

	type OrgId struct {
		OrgId int `json:"org_id"`
	}

	orgId := OrgId{OrgId: org_Id}
	orgIdByte, err := json.Marshal(orgId)
	if err != nil {
		log.Println("err marshalling org id ", err)
		response := utils.Response{Status: 500, Message: "error marshalling org id", Err: err}
		log.Println("response", response)
		return response
	}

	//calling api to fetch data from db
	apiResp, err := ApiCall("POST", config.GET_MODIFIED_CONFIG, orgIdByte, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}
	respData, _ := ioutil.ReadAll(apiResp.Body)
	if err != nil {
		fmt.Println("Can't readAll resp body", err)
	}
	log.Println("respData of modified config :", respData)
	type ApiData struct {
		Status         int    `json:"status,omitempty"`
		Message        string `json:"message,omitempty"`
		ModifiedConfig string `json:"data,omitempty"`
		Err            error  `json:"err,omitempty"`
	}

	//unmarshalling resp data
	var modifiedData ApiData
	err = json.Unmarshal(respData, &modifiedData)
	if err != nil {
		log.Println("error unmarshalling ", err)
		response := utils.Response{Status: 500, Message: "error unmarshalling modified config", Err: err}
		log.Println("response", response)
		return response
	}

	modifiedConfig := modifiedData.ModifiedConfig
	log.Println("modifiedConfig data after marshaling:", modifiedConfig)

	signOrg, err := signConfigUpdate(resmgmtClient, clCtx, config.CHANNEL_ID, modifiedConfig)
	if err != nil {
		panic(err)
	}

	log.Println("come out of signConfigUpdate", signOrg)

	var buffOrg bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buffOrg, signOrg); err != nil {
		log.Println("error while deep marshaliing", err)
		response := utils.Response{Status: 500, Message: "error unmarshalling api", Err: err}
		return response
	}
	signatureString := buffOrg.String()
	log.Println("signatureString :", signatureString)

	//callng api to store signatures in db
	type SignatureData struct {
		OrgId      int    `json:"org_id"`
		SigningOrg int    `json:"signingorg_id"`
		Signatures string `json:"signatures"`
	}

	signData := SignatureData{OrgId: orgId.OrgId, SigningOrg: loggedInId, Signatures: signatureString}
	signDataByte, _ := json.Marshal(signData)

	log.Println("org to be signed id is ", orgId.OrgId)

	//calling api to fetch data from db
	apiResp, err = ApiCall("POST", config.SAVE_ORG_SIGN, signDataByte, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}
	respData, _ = ioutil.ReadAll(apiResp.Body)
	type OrgSignature struct {
		Id int `gorm:"primaryKey;autoIncrement"`
		// ChaincodeId int       `json:"chaincode_id"`
		OrgId     int    `json:"org_id"`
		OrgMsp    string `json:"org_msp"`
		SignbyId  int    `json:"signby_id"`
		Signature string `json:"signature" gorm:"type:text"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	}

	fmt.Println("response Body:", string(respData))

	type ApiResponse struct {
		Status  int          `json:"status,omitempty"`
		Message string       `json:"message,omitempty"`
		Data    OrgSignature `json:"data,omitempty"`
		Err     error        `json:"err,omitempty"`
	}

	var apiResponse ApiResponse
	err = json.Unmarshal(respData, &apiResponse)
	if err != nil {
		log.Println("error unmarshalling ", err)
		response := utils.Response{Status: 500, Message: "error unmarshalling api", Err: err}
		log.Println("response", response)
		return response
	}

	orgIdString := strconv.Itoa(orgId.OrgId)
	fmt.Println(orgIdString)

	loggedInIdString := strconv.Itoa(loggedInId)
	fmt.Println(loggedInIdString)

	// add signorg details to db
	err = ioutil.WriteFile("signed_org"+"_"+orgIdString+"_"+loggedInIdString+".json", buffOrg.Bytes(), 0777)
	if err != nil {
		panic(err)
	}

	response := utils.Response{
		Status:  apiResponse.Status,
		Message: apiResponse.Message,
		Data:    apiResponse.Data,
		Err:     apiResponse.Err,
	}
	return response
}

type Organizations struct {
	Id         int       `gorm:"primaryKey;autoIncrement"`
	Name       string    `json:"name"`
	MspId      string    `json:"msp_id"`
	PeersCount int       `json:"peers_count"`
	Config     string    `json:"file" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

////get Organizations list except the login one
func GetOrganizations() utils.Response {
	log.Println("user token is ", userToken)
	log.Println("org id ", orgId)

	type GetOrgId struct {
		OrgId int `json:"org_id"`
	}

	orgId := GetOrgId{OrgId: orgId}

	orgData, _ := json.Marshal(orgId)

	//calling api to get list of org from db
	apiResp, err := ApiCall("POST", config.ORG_LIST, orgData, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}

	log.Println("response from chaincode/status from admin", apiResp)
	respData, _ := ioutil.ReadAll(apiResp.Body)
	if err != nil {
		fmt.Println("Can't readAll resp body", err)
		response := utils.Response{
			Status:  500,
			Message: "error converting response data to bytes array",
			// Data: ,
			Err: err,
		}
		return response
	}

	type OrgSignData struct {
		OrgId       int         `json:"org_id"`
		OrgName     string      `json:"org_name"`
		OrgMsp      string      `json:"org_msp"`
		SignedbyOrg interface{} `json:"signedby_org"`
		Join_Status int         `json:"join_status"`
		CreatedAt   string      `json:"created_at"`
	}

	//fetching called api response into
	type Data struct {
		Status  int           `json:"status"`
		Message string        `json:"message"`
		Data    []OrgSignData `json:"data"`
		Err     error         `json:"err,omitempty"`
	}

	//unmarshalling resp data
	var orgList Data
	err = json.Unmarshal(respData, &orgList)
	if err != nil {
		log.Println("error unmarshalling ", err)
		response := utils.Response{Status: 500, Message: "error unmarshalling organizations data", Err: err}
		log.Println("response", response)
		return response
	}

	response := utils.Response{Status: orgList.Status, Message: orgList.Message, Data: orgList.Data, Err: orgList.Err}
	log.Println("response", response)
	return response
}

// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@    utility function  @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

//zip download file function //table name chaincode_upgrade
func FileUpload(body *gin.Context) error {
	type CcUpdate struct {
		// File    byte   `json:"file"`
		Name    string `json:"name"`
		Label   string `json:"label"`
		Version string `json:"version"`
		Url     string `json:"url"`
		// File    form   `json:"file"`
		// CreatedAt     string `json:"createdat"`
		// // UpdatedAt   string `json:"updatedat"`
	}

	name := body.Request.PostFormValue("name")
	log.Println(reflect.TypeOf(name))
	label := body.Request.PostFormValue("label")
	version := body.Request.PostFormValue("version")
	url := body.Request.PostFormValue("url")
	// if err := body.ParseForm(); err != nil {
	// 	// handle error
	// }

	file, err := body.FormFile("file")
	if err != nil {
		log.Println("err 2", err)
		// body.IndentedJSON(http.StatusNotFound, gin.H{"message": "File uploading fail"})
		return err
	}
	jsonFilePath := filepath.Join("public/jsonFile/")
	newPath := filepath.Join(jsonFilePath, file.Filename)
	fmt.Printf(" File: %s\n", newPath)

	chaincodeInfo := CcUpdate{Name: name, Label: label, Version: version, Url: url}
	log.Println("data is ", chaincodeInfo)
	log.Println("file is ", file.Filename)

	return nil
}

func GetZipFile(fileName string, url string, zipExtractPath string) error {

	err := DownloadFile(url, fileName)
	if err != nil {
		return err
		// panic(err)

	} else {
		log.Println("Downloading file")
		log.Println("zip path is", zipExtractPath)
	}

	unzipSource(fileName, zipExtractPath)
	// unzip(CcPath1)
	return nil
}

func GetFile(fileName string, body *gin.Context, zipExtractPath string) error {
	type File struct {
		Url string `json:"url"`
	}

	var file File
	log.Println("entering Install chaincode controller", body)
	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := body.BindJSON(&file); err != nil {
		log.Println("err post route @@@@@@@@@@ ", err)
		return err

	} else {
		log.Println("payload from backend  @@@@@@@@@@@@@@@", file)
		fmt.Println("req.body = ", reflect.TypeOf(file))

	}
	// filename := urlparse(file.Url)
	url := file.Url
	// fileName := ""
	err := DownloadFile(url, fileName)
	if err != nil {
		return err
		// panic(err)

	} else {
		log.Println("Downloading file")
		log.Println("zip path is", zipExtractPath)
	}

	unzipSource(fileName, zipExtractPath)
	// unzip(CcPath1)
	return nil
}

func DownloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func unzipSource(source, zipExtractPath string) error {
	//source is filename
	// 1. Open the zip file
	//
	log.Println("enetered unzip source 1")
	reader, err := zip.OpenReader(source)
	if err != nil {
		log.Println("err in destination path", err)
		return err
	}

	defer reader.Close()
	log.Println("enetered unzip source 2", reader)
	// 2. Get the absolute destination path

	log.Println("zip extract path", zipExtractPath)
	destination, err := filepath.Abs(zipExtractPath)
	if err != nil {
		log.Println("enetered unzip source 3")
		log.Println("err in destination path", err)
		return err

	} else {
		log.Println("entered unzip source 4")
		log.Println("destination path is ", destination)
	}

	// // 3. Iterate over zip files inside the archive and unzip each of them
	for _, f := range reader.File {
		err := unzipFile(f, destination)
		if err != nil {
			log.Println("err unzipping in iteration", err)
			return err
		}
	}

	return nil
}

func unzipFile(f *zip.File, destination string) error {
	// 4. Check if file paths are not vulnerable to Zip Slip
	// log.Println("f.name", f.Name)
	filePath := filepath.Join(destination, f.Name)
	log.Println("filepath", filePath)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// 5. Create directory tree
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// 6. Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// 7. Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}

// @@@@@@@@@@@@@@@@@@@@@@ org utility
func signConfigUpdate(ctx *resmgmt.Client, clCtx contextApi.ClientProvider, channelID string, proposedConfigJSON string) (*common.ConfigSignature, error) {
	log.Println("clctx", clCtx, "ctx", ctx, "prposed")

	configUpdate, err := getConfigUpdate(ctx, config.CHANNEL_ID, proposedConfigJSON, config.ORDERER_ENDPOINT)
	if err != nil {
		log.Fatalf("signConfigUpdate getConfigUpdate returned error: %s", err)
	}
	configUpdate.ChannelId = config.CHANNEL_ID

	log.Println("configUpdate in signConfigUpdate fx :", configUpdate)
	configUpdateBytes, err := proto.Marshal(configUpdate)
	if err != nil {
		log.Fatalf("ConfigUpdate marshal returned error: %s", err)

	}
	orgClient, err := clCtx()
	if err != nil {
		log.Fatalf("Client provider returned error: %s", err)
	}
	log.Println("orgClient is ########", orgClient)
	return resource.CreateConfigSignature(orgClient, configUpdateBytes)

}

// getConfigUpdate Get the config update from two configs
func getConfigUpdate(resmgmtClient *resmgmt.Client, channelID string, proposedConfigJSON string, ordererEndPoint string) (*common.ConfigUpdate, error) {
	log.Println("channel id inside get configupdate is ", channelID)

	proposedConfig := &common.Config{}

	err := protolator.DeepUnmarshalJSON(bytes.NewReader([]byte(proposedConfigJSON)), proposedConfig)
	if err != nil {
		return nil, err
	}
	channelConfig, err := getCurrentChannelConfig(resmgmtClient, config.CHANNEL_ID, config.ORDERER_ENDPOINT)
	if err != nil {
		return nil, err
	}
	configUpdate, err := resmgmt.CalculateConfigUpdate(config.CHANNEL_ID, channelConfig, proposedConfig)
	if err != nil {
		return nil, err
	}
	configUpdate.ChannelId = config.CHANNEL_ID

	return configUpdate, nil
}

// getCurrentChannelConfig Get the current channel config
func getCurrentChannelConfig(resmgmtClient *resmgmt.Client, channelID, ordererEndPoint string) (*common.Config, error) {
	block, err := resmgmtClient.QueryConfigBlockFromOrderer(config.CHANNEL_ID, resmgmt.WithOrdererEndpoint(config.ORDERER_ENDPOINT))
	if err != nil {
		log.Println(" getCurrentChannelConfig error", err.Error())
		return nil, err
	}

	return resource.ExtractConfigFromBlock(block)
}

//struct for login credentails
type LoginInput struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	Id        uint      `json:"id" gorm:"unique;primaryKey;autoIncrement"`
	UserName  string    `json:"user_name" gorm:"unique"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"autoUpdateTime"`
}

func ApiCall(method string, url string, data []byte, token string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {

		return nil, err
	}

	if token != "" {
		log.Println("setting token in header")
		req.Header.Set("Authorization", token)
	}

	// Send req using http Client
	client := &http.Client{}
	apiResp, err := client.Do(req)
	if err != nil {

		return nil, err
	}

	return apiResp, nil

}

// create signature
func CreateSignature(resmgmtClient *resmgmt.Client, chConfigPath string, mspClient *msp.Client, user string, id string) utils.Response {
	//need logged in org id from logged in details later for dynamic logged in user id
	loggedInId := orgId
	log.Println("loggedin id :", loggedInId)
	//org_Id of new org to be signed
	var org_Id int
	if i, err := strconv.Atoi(id); err == nil {
		fmt.Printf("i=%d, type: %T\n", i, i)
		org_Id = i
	} else {
		response := utils.Response{Status: 500, Message: "error converting org id to int", Err: err}
		log.Println("response", response)
		return response
	}

	///////////////////////////////////download envelope pb file and use it ////////////////
	type OrgId struct {
		OrgId int `json:"org_id"`
	}

	orgId := OrgId{OrgId: org_Id}
	log.Println("org id in starting :", orgId)

	orgData, _ := json.Marshal(orgId)
	log.Println("orgData in starting :", orgData)

	//calling api to get list of org from db
	apiResp, err := ApiCall("POST", config.ORG, orgData, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}

	respData, _ := ioutil.ReadAll(apiResp.Body)
	if err != nil {
		fmt.Println("Can't readAll resp body", err)
		response := utils.Response{
			Status:  500,
			Message: "error converting response data to bytes array",
			// Data: ,
			Err: err,
		}
		return response
	}

	type OrgData struct {
		Id             int    `gorm:"primaryKey;autoIncrement"`
		Name           string `json:"name"`
		MspId          string `json:"msp_id"`
		PeersCount     int    `json:"peers_count"`
		Config         string `json:"file" gorm:"type:text"`
		ModifiedConfig string `json:"modified_config" gorm:"type:text"`
		Join_Status    int    `json:"join_status"`
		EnvelopeUrl    string `json:"envelope_url"`
		CreatedAt      string `json:"created_at"`
		UpdatedAt      string `json:"updated_at" gorm:"autoUpdateTime"`
	}

	//fetching called api response into
	type Data struct {
		Status  int     `json:"status"`
		Message string  `json:"message"`
		Data    OrgData `json:"data"`
		Err     error   `json:"err,omitempty"`
	}

	//unmarshalling resp data
	var orgList Data
	err = json.Unmarshal(respData, &orgList)
	if err != nil {
		log.Println("error unmarshalling ", err)
		response := utils.Response{Status: 500, Message: "error unmarshalling organizations data", Err: err}
		log.Println("response", response)
		return response
	}

	log.Println("org name :", orgList.Data.Name)

	fileUrl := orgList.Data.EnvelopeUrl
	log.Println("envelope url :", fileUrl)

	filename := filepath.Join("downloadFiles/" + orgList.Data.Name + "_" + "envelope")
	//chaincode zipfile extraction path
	forDest := "/home/riyaz/projects/integra/integra-client-backend/integra-nock-sdk/downloadFiles"
	zipExtractPath := filepath.Join(forDest)
	log.Println("zipeextractpath", zipExtractPath)
	GetZipFile(filename, fileUrl, zipExtractPath)

	log.Println("enetered unzip source 1")
	reader, err := zip.OpenReader(filename)
	if err != nil {
		log.Println("err in destination path", err)
	}

	defer reader.Close()
	log.Println("enetered unzip source 2", reader)
	// 2. Get the absolute destination path
	// getting new chaincode extracted path
	// 3. Iterate over zip files inside the archive and unzip each of them
	var newCCPath string
	for _, f := range reader.File {
		extractedCcPath, err := unzipTestFile(f, forDest)
		if err != nil {
			log.Println("err unzipping in iteration", err)
			// return err
		}
		log.Println("newPath is :", extractedCcPath)
		newCCPath = extractedCcPath
	}

	log.Println("newCCPath", newCCPath)

	// create signature procedure
	usr, err := mspClient.GetSigningIdentity(user)
	if err != nil {
		log.Println("error in finidng user ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in finidng user",
			Err:     err,
		}
		return response
	}
	log.Println("@@@@@@@@@@@@@@@@@@@@@@@@ 4", usr)

	chConfigReader, err := os.Open(newCCPath)
	if err != nil {
		log.Println("failed to create reader for the config  for org1 ", err)
		response := utils.Response{
			Status:  500,
			Message: "failed to create reader for the config  for org1",
			Err:     err,
		}
		return response
	}

	log.Println("@@@@@@@@@@@@@@@@@@@@@@@@ 5", chConfigReader)
	log.Println("@@@@@@@@@@@@@@@@@@@@@@@@")
	signature, err := resmgmtClient.CreateConfigSignatureFromReader(usr, chConfigReader)
	if err != nil {
		log.Println("err getting signing identity ", err)
		response := utils.Response{
			Status:  500,
			Message: "err getting signing identity",
			Err:     err,
		}
		return response
	}
	log.Println("signature in create signature fx :", signature)

	//////////////////////////////////////////////////////////////

	var buffOrg bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buffOrg, signature); err != nil {
		log.Println("error while deep marshaliing", err)
		response := utils.Response{Status: 500, Message: "error unmarshalling api", Err: err}
		return response
	}
	signatureString := buffOrg.String()
	log.Println("signatureString :", signatureString)

	//callng api to store signatures in db
	type SignatureData struct {
		OrgId      int    `json:"org_id"`
		SigningOrg int    `json:"signingorg_id"`
		Signatures string `json:"signatures"`
	}

	signData := SignatureData{OrgId: orgId.OrgId, SigningOrg: loggedInId, Signatures: signatureString}
	signDataByte, _ := json.Marshal(signData)

	log.Println("org to be signed id is ", orgId.OrgId)

	//calling api to fetch data from db
	newApiResp, err := ApiCall("POST", config.SAVE_ORG_SIGN, signDataByte, userToken)
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "error in calling api",
			Err:     err,
		}
		return response

	}

	_, err = ApiCall("POST", config.SIGN_STATUS_UPDATE, orgData, "")
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "Internal server error",
			Err:     err,
		}
		return response

	}

	newRespData, _ := ioutil.ReadAll(newApiResp.Body)
	type OrgSignature struct {
		Id int `gorm:"primaryKey;autoIncrement"`
		// ChaincodeId int       `json:"chaincode_id"`
		OrgId     int    `json:"org_id"`
		OrgMsp    string `json:"org_msp"`
		SignbyId  int    `json:"signby_id"`
		Signature string `json:"signature" gorm:"type:text"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	}

	fmt.Println("response Body:", string(newRespData))

	type ApiResponse struct {
		Status  int          `json:"status,omitempty"`
		Message string       `json:"message,omitempty"`
		Data    OrgSignature `json:"data,omitempty"`
		Err     error        `json:"err,omitempty"`
	}

	var apiResponse ApiResponse
	err = json.Unmarshal(newRespData, &apiResponse)
	if err != nil {
		log.Println("error unmarshalling ", err)
		response := utils.Response{Status: 500, Message: "error unmarshalling api", Err: err}
		log.Println("response", response)
		return response
	}

	orgIdString := strconv.Itoa(orgId.OrgId)
	fmt.Println(orgIdString)

	loggedInIdString := strconv.Itoa(loggedInId)
	fmt.Println(loggedInIdString)

	// add signorg details to db
	err = ioutil.WriteFile("signed_org"+"_"+orgIdString+"_"+loggedInIdString+".json", buffOrg.Bytes(), 0777)
	if err != nil {
		panic(err)
	}

	response := utils.Response{
		Status:  apiResponse.Status,
		Message: apiResponse.Message,
		Data:    apiResponse.Data,
		Err:     apiResponse.Err,
	}
	return response
}

func unzipTestFile(f *zip.File, destination string) (string, error) {
	filePath := filepath.Join(destination, f.Name)
	log.Println("filepath", filePath)

	return filePath, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// , resmgmtClient *resmgmt.Client, clCtx context.ClientProvider,mspClient *msp.Client, userName string

// join channel if save channel config successful
func JoinChannel(id int, resmgmtClient *resmgmt.Client, clCtx contextApi.ClientProvider, mspClient *msp.Client, userName string) utils.Response {
	log.Println("In join channel controller at client side")

	log.Println("id in join channel fx at client side :", id)
	orgAdminIdentity, err := mspClient.GetSigningIdentity(userName)
	if err != nil {
		log.Println("failed to get org3AdminIdentity", err)
		response := utils.Response{
			Status:  500,
			Message: "Inside join channel",

			Err: err,
		}
		return response
	}
	log.Println("org3 admin identity", orgAdminIdentity)

	orgResMgmt, err := resmgmt.New(clCtx)
	if err != nil {
		log.Println("inside error", err)
		response := utils.Response{
			Status:  500,
			Message: "error getting new org res. mangmnt client",

			Err: err,
		}
		return response
	}

	orgPeers, err := DiscoverLocalPeers(clCtx, 1)
	if err != nil {
		log.Println("inside error", err)
		response := utils.Response{
			Status:  500,
			Message: "error getting new org peers",
			// Data:    lastConfigBlock,
			Err: err,
		}
		return response
	}
	log.Println("org3 peers", orgPeers)

	joined, err := IsJoinedChannel(config.CHANNEL_ID, orgResMgmt, orgPeers[0])

	if err != nil {
		log.Println("err in is joined", err)
		response := utils.Response{
			Status:  500,
			Message: "error checking whether peer has joined or not",
			Err:     err,
		}
		return response
	}

	if joined {
		log.Println("already joined")

		response := utils.Response{
			Status:  500,
			Message: "peer already joined",
			Data:    "",
			// Err: err,
		}
		return response

	}

	var lastConfigBlock uint64

	//will return latest block number before joining new org peers
	lastConfigBlock, err = WaitForOrdererConfigUpdate(orgResMgmt, config.CHANNEL_ID, false, lastConfigBlock)
	if err != nil {
		response := utils.Response{
			Status:  500,
			Message: "getting err in lastCOnfig block",
			// Data:    lastConfigBlock,
			Err: err,
		}
		return response
	}

	log.Println("before joining peers config block no. is ", lastConfigBlock)

	// Org3 peers join channel
	err = orgResMgmt.JoinChannel(config.CHANNEL_ID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(config.ORDERER_ENDPOINT))

	if err != nil {
		log.Println("failed to join channel", err)
		response := utils.Response{
			Status:  500,
			Message: "error joining org peer",
			// Data:    lastConfigBlock,
			Err: err,
		}
		return response
	}

	time.Sleep(time.Second * 5)

	channelConfig, _ := getCurrentChannelConfig(resmgmtClient, config.CHANNEL_ID, config.ORDERER_ENDPOINT)

	var buf bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buf, channelConfig); err != nil {
		log.Fatalf("DeepMarshalJSON returned error: %s", err)
	}

	originalChConfigJSON := buf.String()

	// write the whole body at once
	err = ioutil.WriteFile("join-channel-files/config_org.json", []byte(originalChConfigJSON), 0777)
	if err != nil {
		response := utils.Response{
			Status:  500,
			Message: "failed writing config json file",
			Err:     err,
		}
		return response
	}

	peersAppend := "./append-anchor-peers.sh"

	var currentConfigPath = filepath.Join("join-channel-files/config_org.json")
	var newModifiedJsonPath = filepath.Join("join-channel-files/modified_config.json")

	var newConfigPBPath = filepath.Join("join-channel-files/config_org.pb")

	var newModifiedPBPath = filepath.Join("join-channel-files/modified.pb")

	var newConfigUpdatePBPath = filepath.Join("join-channel-files/config_update.pb")

	var newConfigUpdateJSONPath = filepath.Join("join-channel-files/config_update.json")

	var envelopFileJson = filepath.Join("join-channel-files/config-envelope.json")

	// var envelopFilePB = filepath.Join("join-channel-files/config-envelope.pb")
	var envelopFilePBAndAnchorTx = filepath.Join("join-channel-files/orgAnchor")

	channel := config.CHANNEL_ID

	cmd := exec.Command("/bin/sh", peersAppend, config.CORE_PEER_LOCALMSPID, config.HOST, config.PORT, currentConfigPath, newModifiedJsonPath)

	stdout, err := cmd.Output()
	if err != nil {
		log.Println("inside error", err)
		response := utils.Response{
			Status:  422,
			Message: "Inside join channel",
			// Data:    lastConfigBlock,
			Err: err,
		}
		return response
	}

	result := string(stdout)

	log.Println("string in block file", result)
	// // cmd := exec.Command("/bin/sh", app, arg0)

	envelopePbPeers := "./join-channel.sh"

	cmd2 := exec.Command("/bin/sh", envelopePbPeers, currentConfigPath, newModifiedJsonPath, newConfigPBPath, newModifiedPBPath, channel, newConfigUpdatePBPath, newConfigUpdateJSONPath, envelopFileJson, envelopFilePBAndAnchorTx)

	stdout, err = cmd2.Output()
	if err != nil {
		log.Println("inside error", err)
		response := utils.Response{
			Status:  500,
			Message: "Inside join channel",
			// Data:    lastConfigBlock,
			Err: err,
		}
		return response
	}

	result = string(stdout)

	reqNewOrg := resmgmt.SaveChannelRequest{
		ChannelID:         config.CHANNEL_ID,
		ChannelConfigPath: "join-channel-files/orgAnchor.tx",
		// SigningIdentities: []mspProvider.SigningIdentity{org3AdminIdentity},
	}
	txtId, err := orgResMgmt.SaveChannel(reqNewOrg, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(config.ORDERER_ENDPOINT))

	if err != nil {
		log.Println("error updatinf anchor peer", err)
		response := utils.Response{
			Status:  500,
			Message: "Inside join channel txtid",
			// Data:    lastConfigBlock,
			Err: err,
		}
		return response
	}

	log.Println("txtid 2", txtId)

	lastConfigBlock, err = WaitForOrdererConfigUpdate(orgResMgmt, config.CHANNEL_ID, false, lastConfigBlock)
	if err != nil {
		response := utils.Response{
			Status:  500,
			Message: "getting err in lastCOnfig block",
			// Data:    lastConfigBlock,
			Err: err,
		}
		return response
	}

	log.Println("after joining peers config block no. is ", lastConfigBlock)

	type OrgId struct {
		OrgId int `json:"org_id"`
	}

	orgId := OrgId{OrgId: id}
	log.Println("org id in starting :", orgId)

	orgData, _ := json.Marshal(orgId)
	log.Println("orgData in starting :", orgData)

	_, err = ApiCall("POST", config.JOIN_STATUS_UPDATE, orgData, "")
	if err != nil {
		log.Println("error in api calling ", err)
		response := utils.Response{
			Status:  500,
			Message: "Internal server error",
			Err:     err,
		}
		return response

	}

	response := utils.Response{
		Status:  200,
		Message: "Anchor peer successfully update",
		Data:    txtId,
		// Err: err,
	}
	return response
}

// DiscoverLocalPeers queries the local peers for the given MSP context and returns all of the peers. If
// the number of peers does not match the expected number then an error is returned.
func DiscoverLocalPeers(ctxProvider contextApi.ClientProvider, expectedPeers int) ([]fabAPI.Peer, error) {
	ctx, err := contextImpl.NewLocal(ctxProvider)
	if err != nil {
		return nil, errors.Wrap(err, "error creating local context")
	}

	discoveredPeers, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			peers, serviceErr := ctx.LocalDiscoveryService().GetPeers()
			if serviceErr != nil {
				return nil, errors.Wrapf(serviceErr, "error getting peers for MSP [%s]", ctx.Identifier().MSPID)
			}
			if len(peers) < expectedPeers {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Expecting %d peers but got %d", expectedPeers, len(peers)), nil)
			}
			return peers, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return discoveredPeers.([]fabAPI.Peer), nil
}

// IsJoinedChannel returns true if the given peer has joined the given channel
func IsJoinedChannel(channelID string, resMgmtClient *resmgmt.Client, peer fabAPI.Peer) (bool, error) {
	resp, err := resMgmtClient.QueryChannels(resmgmt.WithTargets(peer))
	// log.Println("response in join check fx", resp, err)
	if err != nil {
		return false, err
	}
	for _, chInfo := range resp.Channels {
		if chInfo.ChannelId == channelID {
			return true, nil
		}
	}
	return false, nil
}

func WaitForOrdererConfigUpdate(client *resmgmt.Client, channelID string, genesis bool, lastConfigBlock uint64) (uint64, error) {

	blockNum, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			chConfig, err := client.QueryConfigFromOrderer(channelID, resmgmt.WithOrdererEndpoint(config.ORDERER_ENDPOINT))
			if err != nil {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), err.Error(), nil)
			}

			currentBlock := chConfig.BlockNumber()
			if currentBlock <= lastConfigBlock && !genesis {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Block number was not incremented [%d, %d]", currentBlock, lastConfigBlock), nil)
			}

			block, err := client.QueryConfigBlockFromOrderer(channelID, resmgmt.WithOrdererEndpoint(config.ORDERER_ENDPOINT))
			if err != nil {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), err.Error(), nil)
			}
			if block.Header.Number != currentBlock {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Invalid block number [%d, %d]", block.Header.Number, currentBlock), nil)
			}

			return &currentBlock, nil
		},
	)

	if err == nil {
		return *blockNum.(*uint64), nil
	} else {
		log.Println("err inside block ", err)
		return 0, err
	}

}
