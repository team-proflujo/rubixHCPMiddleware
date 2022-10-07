package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"flag"
	"fmt"
	"os"
)

func initApp() {
	tempAppConfig, configError := getConfigData()

	if configError != nil {
		fmt.Println("Error while trying to get Config Data: " + configError.Error())
		os.Exit(1)
	}

	if (globalVars.ConfigDataStruct{} == tempAppConfig) {
		fmt.Println("Unable to get the Config Data!")
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

func handleRequests() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintf(w, "Welcome to rubix middleware")
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			fmt.Fprintf(w, "405 Method Not Allowed")
			return
		}

		byteReqData, reqDataError := readAppReqData(r)

		if reqDataError != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Error while trying to read Request Data")
			return
		}

		type RegisterWalletReqData struct {
			Password string
		}

		var reqData RegisterWalletReqData

		reqJsonError := json.Unmarshal(byteReqData, &reqData)

		reqData.Password = strings.TrimSpace(reqData.Password)

		if reqJsonError != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Invalid Request! Request data must be a valid JSON.")
			return
		} else if len(reqData.Password) == 0 {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Password must not be empty!")
			return
		}

		response := hcpRegisterWallet(reqData.Password)
		jsonResponse, jsonError := json.Marshal(response)

		if jsonError != nil {
			fmt.Fprintf(w, "Error when converting Response to JSON string: "+jsonError.Error())
			return
		}

		if response.Success {
			globalVars.AppConfig.HcpAccessToken = ""
			updateConfigData(globalVars.AppConfig)
		}

		fmt.Fprintf(w, string(jsonResponse))
	})

	http.HandleFunc("/wallet-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			fmt.Fprintf(w, "405 Method Not Allowed")
			return
		}

		byteReqData, reqDataError := readAppReqData(r)

		if reqDataError != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Error while trying to read Request Data")
			return
		}

		type WalletReqData struct {
			Password string
		}

		var reqData WalletReqData

		reqJsonError := json.Unmarshal(byteReqData, &reqData)

		reqData.Password = strings.TrimSpace(reqData.Password)

		if reqJsonError != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Invalid Request! Request data must be a valid JSON.")
			return
		} else if len(reqData.Password) == 0 {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Password must not be empty!")
			return
		}

		response := hcpGetWalletData(reqData.Password)

		jsonResponse, jsonError := json.Marshal(response)

		if jsonError != nil {
			fmt.Println("Error when converting Response to JSON string: " + jsonError.Error())
			os.Exit(1)
		}

		fmt.Fprintf(w, string(jsonResponse))
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
