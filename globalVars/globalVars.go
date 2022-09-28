package globalVars

type ConfigDataStruct struct {
	HcpAPIURL                 string
	HcpAccessToken            string
	HcpNamespace              string
	HcpBaseSecretsPath        string
	IsHcpAccessTokenEncrypted bool
}

type DIDInfoStruct struct {
	Peerid     string
	DidHash    string
	WalletHash string
}

var AppConfig ConfigDataStruct
