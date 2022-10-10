package main

import (
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"encoding/json"
	"errors"
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
		"X-Vault-Namespace": globalVars.AppConfig.HCPStorageConfig.Namespace,
	}

	if len(globalVars.AppConfig.HCPStorageConfig.AccessToken) > 0 {
		reqHeader["X-Vault-Token"] = globalVars.AppConfig.HCPStorageConfig.AccessToken
	}

	return
}

func sendHCPAPIRequest(url string, method string, data map[string]any) (HCPAPIResponse, error) {
	var responseData HCPAPIResponse

	apiResponse, apiReqError := sendHTTPRequest(globalVars.AppConfig.HCPStorageConfig.APIURL+url, method, data, hcpRequestHeader())

	if apiReqError != nil {
		return responseData, errors.New(apiReqError.Error())
	}

	if len(apiResponse) > 0 {
		// Convert response data to Struct
		json.Unmarshal([]byte(apiResponse), &responseData)
	}

	return responseData, nil
}

func hcpSecretDataURL(didInfo globalVars.DIDInfoStruct) (url string) {
	url = "/v1/" + globalVars.AppConfig.HCPStorageConfig.SecretEngineName + "/data"

	if len(globalVars.AppConfig.HCPStorageConfig.SecretPathPrefix) > 0 {
		url += "/" + globalVars.AppConfig.HCPStorageConfig.SecretPathPrefix
	}

	url += "/" + didInfo.DidHash

	return
}

func hcpRegisterWallet(reqData globalVars.AppRegisterMethodReqDataStruct) (response globalVars.APPHTTPResponse) {
	// Initialize Map before using it (otherwise, it would be nil)
	response.Data = map[string]any{}

	if len(reqData.Password) == 0 {
		response.Message = "Password must not be empty!"
		return
	}

	didInfo, didInfoError := getDIDInfo()

	if didInfoError != nil {
		response.Message = "Error while trying to get DID Info"
		response.Error = didInfoError.Error()
		return
	}

	walletData, prepWalletDataErr := prepareWalletDataToRegister(didInfo, false, "")

	if prepWalletDataErr != nil {
		response.Message = "Error occurred while preparing Wallet Data to Register"
		response.Error = prepWalletDataErr.Error()
		return
	}

	// Store Wallet Data to HCP Vault
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

	// Create Wallet User in HCP Vault
	registerUserApiResponse, registerUserApiError := sendHCPAPIRequest("/v1/auth/userpass/users/"+didInfo.DidHash, "post", map[string]any{
		"password": reqData.Password,
		"policies": globalVars.AppConfig.HCPStorageConfig.RegisterPolicies,
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

	postRegisterErr := postRegisterAction(didInfo)

	if postRegisterErr != nil {
		response.Message = "Error occurred after Registration"
		response.Error = postRegisterErr.Error()
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

	// Login Wallet User to HCP Vault
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
			// Obtain access token
			clientToken = apiResponse.Auth.ClientToken
		} else {
			err = errors.New("Login to HCP Vault failed.")
			return
		}
	}

	return
}

func hcpGetWalletData(password string) (response globalVars.APPHTTPResponse) {
	var walletData globalVars.WalletDataInStorage

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

	globalVars.AppConfig.HCPStorageConfig.AccessToken = clientToken

	// Get Wallet Data from HCP Vault
	apiResponse, apiReqError := sendHCPAPIRequest(hcpSecretDataURL(didInfo), "get", nil)

	if apiReqError != nil {
		response.Message = "Error while trying to get DID Info"
		response.Error = didInfoError.Error()
		return
	}

	if len(apiResponse.Data) > 0 {
		// Retrieve Wallet Data from Response JSON
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
		}
	} else if len(apiResponse.Errors) > 0 {
		response.Message = "Invalid Response from HCP Vault"
		response.Error = apiResponse.Errors[0]
	}

	return response
}
