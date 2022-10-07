package main

import (
	"path/filepath"
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

func hcpSecretDataURL(didInfo globalVars.DIDInfoStruct) (url string) {
	url = "/v1/" + globalVars.AppConfig.HcpSecretEngineName + "/data"

	if len(globalVars.AppConfig.HcpSecretPathPrefix) > 0 {
		url += "/" + globalVars.AppConfig.HcpSecretPathPrefix
	}

	url += "/" + didInfo.DidHash

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

	homeDir, homeDirError := os.UserHomeDir()

	if homeDirError != nil {
		response.Message = "Error while trying to get PrivateShare.png"
		response.Error = homeDirError.Error()
		return
	}

	privateSharePngPath := filepath.Join(homeDir, "Rubix/DATA/", didInfo.DidHash, "PrivateShare.png")

	privateSharePngContent, privateSharePngError := readFile(privateSharePngPath)

	if privateSharePngError != nil {
		response.Message = "Error while trying to get PrivateShare.png"
		response.Error = privateSharePngError.Error()
		return
	} else if len(privateSharePngContent) == 0 {
		response.Message = "Unable to get PrivateShare.png"
		return
	}

	var walletData globalVars.WalletDataInHCPVault

	walletData.DIDHash = didInfo.DidHash
	walletData.PeerId = didInfo.PeerId
	walletData.PrivateSharePng = base64Encode(privateSharePngContent)

	walletInfoStoreApiResponse, walletInfoStoreApiError := sendHCPAPIRequest(hcpSecretDataURL(didInfo), "post", map[string]any{
		"data": walletData,
	})

	if walletInfoStoreApiError != nil {
		response.Message = "Error when Storing Data to HCP Vault"
		response.Error = walletInfoStoreApiError.Error()
		return
	} else if len(walletInfoStoreApiResponse.Errors) > 0 {
		response.Message = "Error when Storing Data to HCP Vault"
		response.Error = walletInfoStoreApiResponse.Errors[0]
		return
	}

	registerUserApiResponse, registerUserApiError := sendHCPAPIRequest("/v1/auth/userpass/users/"+didInfo.DidHash, "post", map[string]any{
		"password": password,
		"policies": globalVars.AppConfig.RegisterPolicies,
	})

	if registerUserApiError != nil {
		response.Message = "Error when Registering to HCP Vault"
		response.Error = registerUserApiError.Error()
		return
	} else if len(registerUserApiResponse.Errors) > 0 {
		response.Message = "Error when Registering to HCP Vault"
		response.Error = registerUserApiResponse.Errors[0]
		return
	}

	response.Success = true
	response.Message = "Successfully registered Wallet to HCP Vault."

	return
}

func hcpLoginWallet(didInfo globalVars.DIDInfoStruct, password string) (clientToken string, err error) {
	if len(didInfo.DidHash) == 0 {
		err = errors.New("Wallet has not been registered with HCP Vault!")
		return
	}

	apiResponse, apiReqError := sendHCPAPIRequest("/v1/auth/userpass/login/"+didInfo.DidHash, "post", map[string]any{
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

	clientToken, loginError := hcpLoginWallet(didInfo, password)

	if loginError != nil {
		response.Message = "Error while Logging in to HCP Vault"
		response.Error = loginError.Error()
		return
	} else if len(clientToken) == 0 {
		response.Message = "HCP Vault Login failed"
		return
	}

	globalVars.AppConfig.HcpAccessToken = clientToken

	apiResponse, apiReqError := sendHCPAPIRequest(hcpSecretDataURL(didInfo), "get", nil)

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
