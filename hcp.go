package main

import (
	"path/filepath"
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"encoding/json"
	"errors"
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
		// Convert response data to Struct
		json.Unmarshal([]byte(apiResponse), &responseData)
	}

	return responseData, nil
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

	// Prepare PrivateShare.png absolute path
	privateSharePngPath := filepath.Join(homeDir, "Rubix/DATA/", didInfo.DidHash, "PrivateShare.png")

	// Read PrivateShare.png content
	privateSharePngContent, privateSharePngError := readFile(privateSharePngPath)

	if privateSharePngError != nil {
		response.Message = "Error while trying to get PrivateShare.png"
		response.Error = privateSharePngError.Error()
		return
	} else if len(privateSharePngContent) == 0 {
		response.Message = "Unable to get PrivateShare.png"
		return
	}

	// Prepare DID.png absolute path
	didPngPath := filepath.Join(homeDir, "Rubix/DATA/", didInfo.DidHash, "DID.png")

	// Read DID.png content
	didPngContent, didPngError := readFile(didPngPath)

	if didPngError != nil {
		response.Message = "Error while trying to get DID.png"
		response.Error = didPngError.Error()
		return
	} else if len(didPngContent) == 0 {
		response.Message = "Unable to get DID.png"
		return
	}

	// Prepare privatekey.pem absolute path
	privateKeyPemPath := filepath.Join(homeDir, "Rubix/DATA/", "privatekey.pem")

	// Read privatekey.pem content
	privateKeyPemContent, privateKeyPemError := readFile(privateKeyPemPath)

	if privateKeyPemError != nil {
		response.Message = "Error while trying to get privatekey.pem"
		response.Error = privateKeyPemError.Error()
		return
	} else if len(privateKeyPemContent) == 0 {
		response.Message = "Unable to get privatekey.pem"
		return
	}

	var walletData globalVars.WalletDataInHCPVault

	walletData.DIDHash = didInfo.DidHash
	walletData.PeerId = didInfo.PeerId
	walletData.PrivateSharePng = base64Encode(privateSharePngContent)
	walletData.DIDPng = base64Encode(didPngContent)
	walletData.PrivateKeyPem = base64Encode(privateKeyPemContent)

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

	// Remove PrivateShare.png
	privateSharePngRmError := os.Remove(privateSharePngPath)

	if privateSharePngRmError != nil {
		response.Message = "Error when Unlinking PrivateShare.png"
		response.Error = privateSharePngRmError.Error()
		return
	}

	// Remove DID.png
	didPngRmError := os.Remove(didPngPath)

	if didPngRmError != nil {
		response.Message = "Error when Unlinking DID.png"
		response.Error = didPngRmError.Error()
		return
	}

	// Remove privatekey.pem
	privateKeyPemRmError := os.Remove(privateKeyPemPath)

	if privateKeyPemRmError != nil {
		response.Message = "Error when Unlinking privatekey.pem"
		response.Error = privateKeyPemRmError.Error()
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
