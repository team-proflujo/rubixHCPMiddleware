package main

import (
	"encoding/json"
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"flag"
	"fmt"
	"os"
)

func initApp() {
	// fmt.Println("Fetching Config Data...")

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

func main() {
	initApp()

	var operation string

	flag.StringVar(&operation, "o", "", "Operation")

	flag.Parse()

	switch operation {
	case "register":
		response := hcpRegisterWallet("letmein1!")
		jsonResponse, jsonError := json.Marshal(response)

		if jsonError != nil {
			fmt.Println("Error when converting Response to JSON string: " + jsonError.Error())
			os.Exit(1)
		}

		fmt.Println("Response: " + string(jsonResponse))
		// hcpCheckToken()
	case "wallet-data":
		response := hcpGetWalletData("123")

		jsonResponse, jsonError := json.Marshal(response)

		if jsonError != nil {
			fmt.Println("Error when converting Response to JSON string: " + jsonError.Error())
			os.Exit(1)
		}

		fmt.Println("Response: " + string(jsonResponse))
	default:
		fmt.Println("Invalid Argument value! Operation argument value does not supported.")
		os.Exit(1)
	}
}
