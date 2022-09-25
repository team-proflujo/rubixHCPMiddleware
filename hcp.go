package main

import (
	"fmt"
	"os"
)

func hcpStoreData() {
	apiResponse, apiReqError := sendRequest("", "get", nil, nil, true)

	if apiReqError != nil {
		fmt.Println("Error while sending Request to HCP Vault API: " + apiReqError.Error())
		os.Exit(1)
	}

	if apiResponse != nil {
		fmt.Println(apiResponse)
	} else {
		fmt.Println("Invalid Response from HCP Vault API: " + apiReqError.Error())
		os.Exit(1)
	}
}

func hcpCheckToken(token string) (isValid bool) {
	//

	return
}

func hcpRegisterWallet() {
	didInfo, didInfoError := getDIDInfo()

	// TODO: Unable to convert any type to Array
	if didInfoError != nil {
		fmt.Printf("Error while trying to get DID Info: " + didInfoError.Error())
		os.Exit(1)
	}

	fmt.Println(didInfo)
}
