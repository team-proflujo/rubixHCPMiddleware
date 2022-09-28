package main

import (
	"encoding/json"
	"errors"
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"fmt"
	"os"
)

type HCPAPIResponse struct {
	Errors []string
	Data   map[string]any
}

func hcpRequestHeader() (reqHeader map[string]string) {
	reqHeader = map[string]string{
		"X-Vault-Token":     globalVars.AppConfig.HcpAccessToken,
		"X-Vault-Namespace": globalVars.AppConfig.HcpNamespace,
	}

	return
}

func sendHCPAPIRequest(url string, method string, data map[string]any) (HCPAPIResponse, error) {
	var responseData HCPAPIResponse

	apiResponse, apiReqError := sendHTTPRequest(url, method, data, hcpRequestHeader())

	if apiReqError != nil {
		return responseData, errors.New(apiReqError.Error())
	}

	if len(apiResponse) > 0 {
		json.Unmarshal([]byte(apiResponse), &responseData)
	}

	return responseData, nil
}

func hcpStoreData() {
	apiResponse, apiReqError := sendHCPAPIRequest("", "get", nil)

	if apiReqError != nil {
		fmt.Println("Error while sending Request to HCP Vault API: " + apiReqError.Error())
		os.Exit(1)
	}

	fmt.Println(apiResponse)
}

func hcpCheckToken() (isValid bool) {
	fmt.Println("Requesting HCP Vault...")

	apiResponse, apiReqError := sendHCPAPIRequest(globalVars.AppConfig.HcpAPIURL+"/v1/auth/token/lookup-self", "get", nil)

	if apiReqError != nil {
		fmt.Println("Error from HCP Vault API: " + apiReqError.Error())
		os.Exit(1)
	}

	fmt.Println(apiResponse)

	return
}

func hcpRegisterWallet() {
	didInfo, didInfoError := getDIDInfo()

	if didInfoError != nil {
		fmt.Printf("Error while trying to get DID Info: " + didInfoError.Error())
		os.Exit(1)
	}

	fmt.Println(didInfo)
}
