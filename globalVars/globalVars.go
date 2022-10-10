package globalVars

import (
	"log"
)

type StorageTarget_HCPVault_ConfigStruct struct {
	APIURL           string   `json:"apiURL"`
	AccessToken      string   `json:"accessToken"`
	Namespace        string   `json:"namespace"`
	SecretEngineName string   `json:"secretEngineName"`
	SecretPathPrefix string   `json:"secretPathPrefix"`
	RegisterPolicies []string `json:"registerPolicies"`
}

type StorageTarget_AWS_ConfigStruct struct {
	APIEndpoint string `json:"apiEndPoint"`
	Bucket      string `json:"bucket"`
	AccessKey   string `json:"accessKey"`
	Secret      string `json:"secret"`
	Region      string `json:"region"`
	RootFolder  string `json:"rootFolder"`
}

type StorageTarget_Local_ConfigStruct struct {
	Location string `json:"location"`
}

type ConfigDataStruct struct {
	TargetStorage      string                              `json:"targetStorage"`
	HCPStorageConfig   StorageTarget_HCPVault_ConfigStruct `json:"hcpVaultStorageConfig"`
	AWSStorageConfig   StorageTarget_AWS_ConfigStruct      `json:"awsStorageConfig"`
	LocalStorageConfig StorageTarget_Local_ConfigStruct    `json:"localStorageConfig"`
	// Parameters from Script:
	WalletRegisteredToStorage bool `json:"walletRegisterToStorage"`
	TargetStorageName         string
}

type DIDInfoStruct struct {
	PeerId     string
	DidHash    string
	WalletHash string
}

type APPHTTPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Error   string `json:"error"`
}

type WalletDataInStorage struct {
	DIDHash         string `json:"didHash"`
	PeerId          string `json:"peerId"`
	PrivateSharePng string `json:"PrivateShare.png"`
	DIDPng          string `json:"DID.png"`
	PrivateKeyPem   string `json:"privatekey.pem"`
}

var AppConfig ConfigDataStruct

type AppLoggerStruct struct {
	Info    *log.Logger
	Debug   *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

var AppLogger AppLoggerStruct

type AppRegisterMethodReqDataStruct struct {
	Password string
}
