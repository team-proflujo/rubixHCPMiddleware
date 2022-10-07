package main

import (
	"encoding/json"
	"net/http"
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

		response := hcpRegisterWallet("letmein1!")
		jsonResponse, jsonError := json.Marshal(response)

		if jsonError != nil {
			fmt.Println("Error when converting Response to JSON string: " + jsonError.Error())
			os.Exit(1)
		}

		fmt.Println("Response: " + string(jsonResponse))

	})

	http.HandleFunc("/wallet-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			fmt.Fprintf(w, "405 Method Not Allowed")
			return
		}

		response := hcpGetWalletData("123")

		jsonResponse, jsonError := json.Marshal(response)

		if jsonError != nil {
			fmt.Println("Error when converting Response to JSON string: " + jsonError.Error())
			os.Exit(1)
		}

		fmt.Println("Response: " + string(jsonResponse))
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
