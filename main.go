package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"log"
	"os"
)

func initApp() {
	// Initialize Logging
	globalVars.AppLogger = globalVars.AppLoggerStruct{
		Info:    log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime),
		Debug:   log.New(os.Stdout, "DEBUG ", log.Ldate|log.Ltime),
		Warning: log.New(os.Stdout, "WARNING ", log.Ldate|log.Ltime),
		Error:   log.New(os.Stdout, "ERROR ", log.Ldate|log.Ltime),
	}

	globalVars.AppLogger.Info.Println("Fetching Config Data...")

	// Read Config Data
	tempAppConfig, configError := getConfigData()

	if configError != nil {
		globalVars.AppLogger.Error.Println("Error while trying to get Config Data: " + configError.Error())
		os.Exit(1)
	}

	if len(tempAppConfig.TargetStorage) == 0 {
		globalVars.AppLogger.Error.Println("Unable to get the Config Data!")
		os.Exit(1)
	}

	switch tempAppConfig.TargetStorage {
	case "hcp-vault":
		if len(tempAppConfig.HCPStorageConfig.APIURL) == 0 || len(tempAppConfig.HCPStorageConfig.Namespace) == 0 || len(tempAppConfig.HCPStorageConfig.SecretEngineName) == 0 || len(tempAppConfig.HCPStorageConfig.RegisterPolicies) == 0 {
			globalVars.AppLogger.Error.Println("Invalid HCP Vault storage config data in config.json")
			os.Exit(1)
		}

		tempAppConfig.TargetStorageName = "HCP Vault"
	case "aws":
		if len(tempAppConfig.AWSStorageConfig.APIEndpoint) == 0 || len(tempAppConfig.AWSStorageConfig.Bucket) == 0 || len(tempAppConfig.AWSStorageConfig.AccessKey) == 0 || len(tempAppConfig.AWSStorageConfig.Secret) == 0 || len(tempAppConfig.AWSStorageConfig.Region) == 0 {
			globalVars.AppLogger.Error.Println("Invalid AWS storage config data in config.json")
			os.Exit(1)
		}

		tempAppConfig.TargetStorageName = "AWS"
	case "other":
		tempAppConfig.TargetStorageName = "Other"
	default:
		globalVars.AppLogger.Error.Println("Invalid target storage: " + tempAppConfig.TargetStorage)
		os.Exit(1)
	}

	globalVars.AppLogger.Info.Println("Storage target: " + tempAppConfig.TargetStorageName)

	// Set Global config variable
	globalVars.AppConfig = globalVars.ConfigDataStruct(tempAppConfig)
}

func readAppReqData(r *http.Request) (data []byte, err error) {
	data, reqDataError := ioutil.ReadAll(r.Body)

	if reqDataError != nil {
		err = reqDataError
		return
	}

	return
}

func respondJson(w http.ResponseWriter, statusCode int, response any) {
	w.Header().Set("Content-Type", "application/json")

	bytesJson, jsonEncodeError := json.Marshal(response)

	if jsonEncodeError != nil {
		// Fail when Unable to convert the Response to JSON
		globalVars.AppLogger.Error.Println("Error when encoding Response Data: " + jsonEncodeError.Error())

		w.WriteHeader(500)
		w.Write([]byte("{\"success\": false, \"message\": \"Error occurred while trying to Respond.\"}"))
		return
	}

	w.WriteHeader(statusCode)
	w.Write(bytesJson)
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	respondJson(w, statusCode, globalVars.APPHTTPResponse{
		Success: false,
		Message: message,
	})
}

func handleRequests() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		respondJson(w, 200, globalVars.APPHTTPResponse{
			Success: true,
			Message: "Welcome to Rubix-HCP Vault middleware",
		})
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, 405, "405 Method Not Allowed")
			return
		}

		if globalVars.AppConfig.WalletRegisteredToStorage {
			respondError(w, 406, "Wallet has already been registered to "+globalVars.AppConfig.TargetStorageName+"!")
			return
		}

		var reqData globalVars.AppRegisterMethodReqDataStruct

		// Get Request Data
		byteReqData, reqDataError := readAppReqData(r)

		if reqDataError != nil {
			respondError(w, 500, "Error while trying to read Request Data")
			return
		}

		// Convert Request Data to Pre-defined format
		reqJsonError := json.Unmarshal(byteReqData, &reqData)

		reqData.Password = strings.TrimSpace(reqData.Password)

		if reqJsonError != nil {
			respondError(w, 400, "Invalid Request! Request data must be a valid JSON.")
			return
		} else if len(reqData.Password) == 0 {
			respondError(w, 400, "Password must not be empty!")
			return
		}

		response := globalVars.APPHTTPResponse{
			Success: false,
		}

		switch globalVars.AppConfig.TargetStorage {
		case "hcp-vault":
			// Register Wallet to HCP Vault
			response = hcpRegisterWallet(reqData)

			if response.Success {
				// Remove RegisterToken when Successfully Registered to HCP Vault
				globalVars.AppConfig.HCPStorageConfig.AccessToken = ""
			}
		case "aws":
			response = awsRegisterWallet(reqData)
		case "other":
			response = localStorageRegisterWallet(reqData)
		}

		if response.Success {
			// Update config.json with new Data
			updateConfigData(globalVars.AppConfig)
		}

		respondJson(w, 200, response)
	})

	http.HandleFunc("/wallet-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, 405, "405 Method Not Allowed")
			return
		}

		var reqData globalVars.AppRegisterMethodReqDataStruct

		// Get Request Data
		byteReqData, reqDataError := readAppReqData(r)

		if reqDataError != nil {
			respondError(w, 500, "Error while trying to read Request Data")
			return
		}

		// Convert Request Data to Pre-defined format
		reqJsonError := json.Unmarshal(byteReqData, &reqData)

		reqData.Password = strings.TrimSpace(reqData.Password)

		if reqJsonError != nil {
			respondError(w, 400, "Invalid Request! Request data must be a valid JSON.")
			return
		} else if len(reqData.Password) == 0 {
			respondError(w, 400, "Password must not be empty!")
			return
		}

		response := globalVars.APPHTTPResponse{
			Success: false,
		}

		switch globalVars.AppConfig.TargetStorage {
		case "hcp-vault":
			// Get Wallet Data stored in HCP Vault
			response = hcpGetWalletData(reqData.Password)
		case "aws":
			response = awsGetWalletData(reqData)
		case "other":
			response = localStorageGetWalletData(reqData)
		}

		respondJson(w, 200, response)
	})

	http.ListenAndServe(":3333", nil)
}

func main() {
	initApp()

	handleRequests()
}
