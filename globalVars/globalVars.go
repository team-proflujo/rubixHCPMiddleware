package globalVars

type ConfigDataStruct struct {
	HcpAPIURL                 string `json:"hcpAPIURL"`
	HcpAccessToken            string `json:"hcpAccessToken"`
	HcpNamespace              string `json:"hcpNamespace"`
	HcpSecretEngineName       string `json:"hcpSecretEngineName"`
	HcpSecretPathPrefix       string `json:"hcpSecretPathPrefix"`
	IsHcpAccessTokenEncrypted bool   `json:"isHcpAccessTokenEncrypted"`
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
	PrivateSharePng string `json:"privateSharePng"`
}

var AppConfig ConfigDataStruct
