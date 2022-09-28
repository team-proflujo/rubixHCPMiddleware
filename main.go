package main

import (
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"flag"
	"fmt"
	"os"
)

func initApp() {
	fmt.Println("Fetching Config Data...")

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
		// hcpRegisterWallet()
		hcpCheckToken()
	default:
		fmt.Println("Invalid Argument value! Operation argument value does not supported.")
		os.Exit(1)
	}
}
