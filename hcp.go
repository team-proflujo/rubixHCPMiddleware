package main

import (
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type HCPAPILoginResponse struct {
	ClientToken string `json:"client_token"`
}

type HCPAPIResponse struct {
	Errors []string
	Data   map[string]any
	Auth   HCPAPILoginResponse
}

func hcpRequestHeader() (reqHeader map[string]string) {
	reqHeader = map[string]string{
		"X-Vault-Namespace": globalVars.AppConfig.HcpNamespace,
	}

	if len(globalVars.AppConfig.HcpAccessToken) > 0 {
		reqHeader["X-Vault-Token"] = globalVars.AppConfig.HcpAccessToken
	}

	return
}

func sendHCPAPIRequest(url string, method string, data map[string]any) (HCPAPIResponse, error) {
	var responseData HCPAPIResponse

	apiResponse, apiReqError := sendHTTPRequest(globalVars.AppConfig.HcpAPIURL+url, method, data, hcpRequestHeader())

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

	apiResponse, apiReqError := sendHCPAPIRequest("/v1/auth/token/lookup-self", "get", nil)

	if apiReqError != nil {
		fmt.Println("Error from HCP Vault API: " + apiReqError.Error())
		os.Exit(1)
	}

	fmt.Println(apiResponse)

	return
}

func hcpRegisterWallet(password string) (response globalVars.APPHTTPResponse) {
	// Initialize Map before using it (otherwise, it would be nil)
	response.Data = map[string]any{}

	if len(password) == 0 {
		response.Message = "Password must not be empty!"
		return
	}

	didInfo, didInfoError := getDIDInfo()

	if didInfoError != nil {
		response.Message = "Error while trying to get DID Info"
		response.Error = didInfoError.Error()
		return
	}

	response.Data = map[string]any{"didInfo": didInfo}
	response.Success = true

	return
}

func hcpLoginWallet(password string) (clientToken string, err error) {
	if len(globalVars.AppConfig.HCPUserName) == 0 {
		err = errors.New("Wallet has not been registered with HCP Vault!")
		return
	}

	apiResponse, apiReqError := sendHCPAPIRequest("/v1/auth/userpass/login/"+globalVars.AppConfig.HCPUserName, "post", map[string]any{
		"password": password,
	})

	if apiReqError != nil {
		err = apiReqError
		return
	}

	if len(apiResponse.Errors) > 0 {
		err = errors.New(apiResponse.Errors[0])
		return
	} else {
		if (HCPAPILoginResponse{} != apiResponse.Auth) {
			clientToken = apiResponse.Auth.ClientToken
		} else {
			err = errors.New("Login to HCP Vault failed.")
			return
		}
	}

	return
}

func hcpGetWalletData(password string) (response globalVars.APPHTTPResponse) {
	var walletData globalVars.WalletDataInHCPVault

	if len(password) == 0 {
		response.Message = "Password must not be empty!"
		return
	}

	didInfo, didInfoError := getDIDInfo()

	if didInfoError != nil {
		response.Message = "Error while trying to get DID Info"
		response.Error = didInfoError.Error()
		return
	}

	clientToken, loginError := hcpLoginWallet(password)

	if loginError != nil {
		response.Message = "Error while Logging in to HCP Vault"
		response.Error = loginError.Error()
		return
	} else if len(clientToken) == 0 {
		response.Message = "HCP Vault Login failed"
		return
	}

	globalVars.AppConfig.HcpAccessToken = clientToken

	apiURL := "/v1/" + globalVars.AppConfig.HcpSecretEngineName + "/data"

	if len(globalVars.AppConfig.HcpSecretPathPrefix) > 0 {
		apiURL += "/" + globalVars.AppConfig.HcpSecretPathPrefix
	}

	apiResponse, apiReqError := sendHCPAPIRequest(apiURL+"/"+didInfo.DidHash, "get", nil)

	if apiReqError != nil {
		response.Message = "Error while trying to get DID Info"
		response.Error = didInfoError.Error()
		return
	}

	if len(apiResponse.Data) > 0 {
		if secretData, secretDataExists := apiResponse.Data["data"]; secretDataExists {
			strSecretData, jsonEncodeError := json.Marshal(secretData)

			if jsonEncodeError != nil {
				response.Message = "Error while parsing Vault Response"
				response.Error = jsonEncodeError.Error()
				return
			}

			jsonDecodeError := json.Unmarshal(strSecretData, &walletData)

			if jsonDecodeError != nil {
				response.Message = "Error while parsing Vault Response"
				response.Error = jsonDecodeError.Error()
				return
			}

			response.Data = walletData
			response.Success = true
			walletData.DIDHash = ""
		}
	} else if len(apiResponse.Errors) > 0 {
		response.Message = "Invalid Response from HCP Vault"
		response.Error = apiResponse.Errors[0]
	}

	return response
}
