package globalVars

import (
	"log"
)

type ConfigDataStruct struct {
	HcpAPIURL                 string   `json:"hcpAPIURL"`
	HcpAccessToken            string   `json:"hcpAccessToken"`
	HcpNamespace              string   `json:"hcpNamespace"`
	HcpSecretEngineName       string   `json:"hcpSecretEngineName"`
	HcpSecretPathPrefix       string   `json:"hcpSecretPathPrefix"`
	IsHcpAccessTokenEncrypted bool     `json:"isHcpAccessTokenEncrypted"`
	RegisterPolicies          []string `json:"registerPolicies"`
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

type WalletDataInHCPVault struct {
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
