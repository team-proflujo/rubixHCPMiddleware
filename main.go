package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"flag"
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

	tempAppConfig, configError := getConfigData()

	if configError != nil {
		globalVars.AppLogger.Error.Println("Error while trying to get Config Data: " + configError.Error())
		os.Exit(1)
	}

	if len(tempAppConfig.HcpAPIURL) == 0 {
		globalVars.AppLogger.Error.Println("Unable to get the Config Data!")
		os.Exit(1)
	}

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

		if len(globalVars.AppConfig.HcpAccessToken) == 0 {
			respondError(w, 406, "Wallet has already been registered to HCP Vault!")
			return
		}

		byteReqData, reqDataError := readAppReqData(r)

		if reqDataError != nil {
			respondError(w, 500, "Error while trying to read Request Data")
			return
		}

		type RegisterWalletReqData struct {
			Password string
		}

		var reqData RegisterWalletReqData

		reqJsonError := json.Unmarshal(byteReqData, &reqData)

		reqData.Password = strings.TrimSpace(reqData.Password)

		if reqJsonError != nil {
			respondError(w, 400, "Invalid Request! Request data must be a valid JSON.")
			return
		} else if len(reqData.Password) == 0 {
			respondError(w, 400, "Password must not be empty!")
			return
		}

		response := hcpRegisterWallet(reqData.Password)

		if response.Success {
			globalVars.AppConfig.HcpAccessToken = ""
			updateConfigData(globalVars.AppConfig)
		}

		respondJson(w, 200, response)
	})

	http.HandleFunc("/wallet-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, 405, "405 Method Not Allowed")
			return
		}

		byteReqData, reqDataError := readAppReqData(r)

		if reqDataError != nil {
			respondError(w, 500, "Error while trying to read Request Data")
			return
		}

		type WalletReqData struct {
			Password string
		}

		var reqData WalletReqData

		reqJsonError := json.Unmarshal(byteReqData, &reqData)

		reqData.Password = strings.TrimSpace(reqData.Password)

		if reqJsonError != nil {
			respondError(w, 400, "Invalid Request! Request data must be a valid JSON.")
			return
		} else if len(reqData.Password) == 0 {
			respondError(w, 400, "Password must not be empty!")
			return
		}

		response := hcpGetWalletData(reqData.Password)

		respondJson(w, 200, response)
	})

	http.ListenAndServe(":3333", nil)
}

func main() {
	initApp()

	var operation string

	flag.StringVar(&operation, "o", "", "Operation")

	flag.Parse()

	handleRequests()
}
