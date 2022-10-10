package main

import (
	"fmt"
	"io"
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/EnsurityTechnologies/enscrypt"
	"golang.org/x/crypto/pbkdf2"
)

func readFile(filePath string) ([]byte, error) {
	var fileContent []byte = nil

	fileContent, fileReadError := os.ReadFile(filePath)

	if fileReadError != nil {
		return nil, errors.New("Error when Reading file: " + fileReadError.Error())
	}

	return fileContent, nil
}

func writeFile(filePath string, data []byte) (bool, error) {
	fileWriteSuccess := false

	filePointer, fileCreateError := os.Create(filePath)

	if fileCreateError != nil {
		return false, errors.New("Error when Creating file: " + fileCreateError.Error())
	}

	if filePointer != nil {
		noOfBytesWritten, fileWriteError := filePointer.Write(data)

		if fileWriteError != nil {
			return false, errors.New("Error when Writing file: " + fileWriteError.Error())
		}

		if noOfBytesWritten == len(data) {
			fileWriteSuccess = true
		} else {
			return false, errors.New("Writing file failed! Some data may not be written.")
		}
	}

	return fileWriteSuccess, nil
}

func base64Encode(rawData []byte) string {
	encodedData := base64.StdEncoding.EncodeToString(rawData)

	return encodedData
}

func base64Decode(encodedData string) (string, error) {
	rawData, decodeError := base64.StdEncoding.DecodeString(encodedData)

	if decodeError != nil {
		return "", errors.New("Error while decoding base64 data: " + decodeError.Error())
	}

	return string(rawData), nil
}

func getDIDInfo() (globalVars.DIDInfoStruct, string, error) {
	var didInfo globalVars.DIDInfoStruct
	didFilePath := "Rubix/DATA/DID.json"

	homeDir, homeDirError := os.UserHomeDir()

	if homeDirError != nil {
		return didInfo, "", errors.New("Error while trying to get Home Directory path: " + homeDirError.Error())
	}

	// Prepare DID.json file absolute path
	didFilePath = filepath.Join(homeDir, didFilePath)

	// Read DID.json content
	didFileContent, fileReadError := readFile(didFilePath)

	if fileReadError != nil {
		return didInfo, "", errors.New("Error while trying to get DID.json file content: " + fileReadError.Error())
	} else if didFileContent == nil {
		return didInfo, "", errors.New("DID.json file is empty!")
	}

	var didInfoList []globalVars.DIDInfoStruct

	// Convert DID.json to Struct
	decodeJsonError := json.Unmarshal(didFileContent, &didInfoList)

	if decodeJsonError != nil {
		return didInfo, "", errors.New("Error while trying to parse DID.json: " + decodeJsonError.Error())
	}

	if len(didInfoList) > 0 {
		didInfo = didInfoList[0]
	}

	return didInfo, string(didFileContent), nil
}

func getScriptPath() (string, error) {
	/*
		path, pathError := filepath.Abs(filepath.Dir(os.Args[0]))

		if pathError != nil {
			return "", errors.New("Error: " + pathError.Error())
		}
	*/

	/*
		path, execPathError := os.Executable()

		if execPathError != nil {
			return "", errors.New("Error: " + execPathError.Error())
		}

		path = filepath.Dir(path)
	*/

	path, pathError := os.Getwd()

	if pathError != nil {
		return "", errors.New("Error: " + pathError.Error())
	}

	return path, nil
}

func getConfigData() (globalVars.ConfigDataStruct, error) {
	var configData globalVars.ConfigDataStruct
	scriptPath, scriptPathError := getScriptPath()

	if scriptPathError != nil {
		return configData, errors.New("Error while trying to get Script Path: " + scriptPathError.Error())
	}

	// Prepare config.json file absolute path
	configFilePath := filepath.Join(scriptPath, "config.json")

	// Read config.json content
	configFileContent, fileReadError := readFile(configFilePath)

	if fileReadError != nil {
		return configData, errors.New("Error while trying to get config.json file content: " + fileReadError.Error())
	} else if configFileContent == nil {
		return configData, errors.New("config.json file is empty!")
	}

	// Convert config.json to Struct
	decodeJsonError := json.Unmarshal(configFileContent, &configData)

	if decodeJsonError != nil {
		return configData, errors.New("Error while trying to parse config.json: " + decodeJsonError.Error())
	}

	return configData, nil
}

func ecdsaP256KeyGen(password string) (*rsa.PublicKey, *rsa.PrivateKey, error) {
	var publicKey *rsa.PublicKey
	var privateKey *rsa.PrivateKey

	privateKeyBytes, publicKeyBytes, keyPairError := enscrypt.GenerateKeyPair(&enscrypt.CryptoConfig{Alg: enscrypt.ECDSAP256, Pwd: password})

	if keyPairError != nil {
		return publicKey, privateKey, keyPairError
	}

	fmt.Println("Public key: " + string(publicKeyBytes))
	fmt.Println("Private key: " + string(privateKeyBytes))

	publicKey, publicKeyConvertError := x509.ParsePKCS1PublicKey(publicKeyBytes)

	if publicKeyConvertError != nil {
		return publicKey, privateKey, publicKeyConvertError
	}

	privateKey, privateKeyConvertError := x509.ParsePKCS1PrivateKey(privateKeyBytes)

	if privateKeyConvertError != nil {
		return publicKey, privateKey, privateKeyConvertError
	}

	return publicKey, privateKey, nil
}

func ecdsaP256Encrypt(password string, rawData string) ([]byte, error) {
	var encryptedData []byte

	publicKey, _, keyError := ecdsaP256KeyGen(password)

	if keyError != nil {
		return nil, keyError
	}

	cryptoReader := rand.Reader

	encryptedData, encryptError := rsa.EncryptPKCS1v15(cryptoReader, publicKey, []byte(rawData))

	if encryptError != nil {
		return nil, encryptError
	}

	return encryptedData, nil
}

func ecdsaP256Decrypt(password string, encryptedData string) ([]byte, error) {
	var decryptedData []byte

	_, privateKey, keyError := ecdsaP256KeyGen(password)

	if keyError != nil {
		return nil, keyError
	}

	cryptoReader := rand.Reader

	decryptedData, decryptError := rsa.DecryptPKCS1v15(cryptoReader, privateKey, []byte(encryptedData))

	if decryptError != nil {
		return nil, decryptError
	}

	return decryptedData, nil
}

func updateConfigData(newAppConfigData globalVars.ConfigDataStruct) (success bool, err error) {
	// Convert Struct data to JSON
	newContent, jsonEncodeError := json.MarshalIndent(newAppConfigData, "", "\t")

	if jsonEncodeError != nil {
		err = errors.New("Error while Converting to JSON: " + jsonEncodeError.Error())
		return
	}

	scriptPath, scriptPathError := getScriptPath()

	if scriptPathError != nil {
		err = errors.New("Error while trying to get Script Path: " + scriptPathError.Error())
		return
	}

	// Prepare config.json absolute path
	configFilePath := filepath.Join(scriptPath, "config.json")

	// Update config.json with new data
	success, err = writeFile(configFilePath, newContent)

	return
}

func prepareWalletDataToRegister(didInfo globalVars.DIDInfoStruct, encryptContent bool, passwordToEncrypt string) (walletData globalVars.WalletDataInStorage, err error) {
	homeDir, homeDirError := os.UserHomeDir()

	if homeDirError != nil {
		err = homeDirError
		return
	}

	// Prepare PrivateShare.png absolute path
	privateSharePngPath := filepath.Join(homeDir, "Rubix/DATA/", didInfo.DidHash, "PrivateShare.png")

	// Read PrivateShare.png content
	bytesPrivateSharePngContent, privateSharePngError := readFile(privateSharePngPath)
	strPrivateSharePngContent := ""

	if privateSharePngError != nil {
		err = privateSharePngError
		return
	} else if len(bytesPrivateSharePngContent) == 0 {
		err = errors.New("Unable to get PrivateShare.png")
		return
	}

	encodedPrivateSharePngContent := base64Encode(bytesPrivateSharePngContent)

	if encryptContent {
		encryptedPrivateSharePngContent, aesEncryptError := aesEncrypt([]byte(encodedPrivateSharePngContent), passwordToEncrypt)

		if aesEncryptError != nil {
			err = aesEncryptError
			return
		}

		strPrivateSharePngContent = encryptedPrivateSharePngContent
	} else {
		strPrivateSharePngContent = encodedPrivateSharePngContent
	}

	// Prepare DID.png absolute path
	didPngPath := filepath.Join(homeDir, "Rubix/DATA/", didInfo.DidHash, "DID.png")

	// Read DID.png content
	bytesDIDPngContent, didPngError := readFile(didPngPath)
	strDIDPngContent := ""

	if didPngError != nil {
		err = didPngError
		return
	} else if len(bytesDIDPngContent) == 0 {
		err = errors.New("Unable to get DID.png")
		return
	}

	encodedDIDPngContent := base64Encode(bytesDIDPngContent)

	if encryptContent {
		encryptedDIDPngContent, aesEncryptError := aesEncrypt([]byte(encodedDIDPngContent), passwordToEncrypt)

		if aesEncryptError != nil {
			err = aesEncryptError
			return
		}

		strDIDPngContent = encryptedDIDPngContent
	} else {
		strDIDPngContent = encodedDIDPngContent
	}

	// Prepare privatekey.pem absolute path
	privateKeyPemPath := filepath.Join(homeDir, "Rubix/DATA/", "privatekey.pem")

	// Read privatekey.pem content
	bytesPrivateKeyPemContent, privateKeyPemError := readFile(privateKeyPemPath)
	strPrivateKeyPemContent := ""

	if privateKeyPemError != nil {
		err = privateKeyPemError
		return
	} else if len(bytesPrivateKeyPemContent) == 0 {
		err = errors.New("Unable to get privatekey.pem")
		return
	}

	encodedPrivateKeyPemContent := base64Encode(bytesPrivateKeyPemContent)

	if encryptContent {
		encryptedPrivateKeyPemContent, aesEncryptError := aesEncrypt([]byte(encodedPrivateKeyPemContent), passwordToEncrypt)

		if aesEncryptError != nil {
			err = aesEncryptError
			return
		}

		strPrivateKeyPemContent = encryptedPrivateKeyPemContent
	} else {
		strPrivateKeyPemContent = encodedPrivateKeyPemContent
	}

	walletData.DIDInfo = didInfo
	walletData.PrivateSharePng = strPrivateSharePngContent
	walletData.DIDPng = strDIDPngContent
	walletData.PrivateKeyPem = strPrivateKeyPemContent

	return
}

func postRegisterAction(didInfo globalVars.DIDInfoStruct) (err error) {
	homeDir, homeDirError := os.UserHomeDir()

	if homeDirError != nil {
		err = homeDirError
		return
	}

	// Prepare PrivateShare.png absolute path
	privateSharePngPath := filepath.Join(homeDir, "Rubix/DATA/", didInfo.DidHash, "PrivateShare.png")

	// Prepare DID.png absolute path
	didPngPath := filepath.Join(homeDir, "Rubix/DATA/", didInfo.DidHash, "DID.png")

	// Prepare privatekey.pem absolute path
	privateKeyPemPath := filepath.Join(homeDir, "Rubix/DATA/", "privatekey.pem")

	// Remove PrivateShare.png
	privateSharePngRmError := os.Remove(privateSharePngPath)

	if privateSharePngRmError != nil {
		err = privateSharePngRmError
		return
	}

	// Remove DID.png
	didPngRmError := os.Remove(didPngPath)

	if didPngRmError != nil {
		err = didPngRmError
		return
	}

	// Remove privatekey.pem
	privateKeyPemRmError := os.Remove(privateKeyPemPath)

	if privateKeyPemRmError != nil {
		err = privateKeyPemRmError
		return
	}

	return
}

func aesKey(password string) (key string) {
	salt := make([]byte, 8)

	byteKey := pbkdf2.Key([]byte(password), salt, 1000, 32, sha256.New)

	key = string(byteKey[:])

	return
}

func aesEncrypt(rawData []byte, password string) (encryptedData string, err error) {
	key := aesKey(password)

	newCipher, aesNewCipherError := aes.NewCipher([]byte(key))

	if aesNewCipherError != nil {
		err = aesNewCipherError
		return
	}

	gcm, gcmError := cipher.NewGCM(newCipher)

	if gcmError != nil {
		err = gcmError
		return
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, nonceReadError := io.ReadFull(rand.Reader, nonce); nonceReadError != nil {
		err = nonceReadError
		return
	}

	// byteEncryptedData := make([]byte, len(rawData))

	// newCipher.Encrypt(byteEncryptedData, []byte(rawData))

	byteEncryptedData := gcm.Seal(nonce, nonce, rawData, nil)

	encryptedData = string(byteEncryptedData[:])

	return
}

func aesDecrypt(encryptedData string, password string) (decryptedData string, err error) {
	key := aesKey(password)

	newCipher, aesNewCipherError := aes.NewCipher([]byte(key))

	if aesNewCipherError != nil {
		err = aesNewCipherError
		return
	}

	gcm, gcmError := cipher.NewGCM(newCipher)

	if gcmError != nil {
		err = gcmError
		return
	}

	byteEncryptedData := []byte(encryptedData)

	nonceSize := gcm.NonceSize()

	if len(byteEncryptedData) < nonceSize {
		err = errors.New("Invalid Encrypted Data!")
		return
	}

	nonce, byteEncryptedData := byteEncryptedData[:nonceSize], byteEncryptedData[nonceSize:]

	// byteDecryptedData := make([]byte, len(encryptedData))

	// newCipher.Decrypt(byteDecryptedData, []byte(encryptedData))

	byteDecryptedData, decryptError := gcm.Open(nil, nonce, []byte(byteEncryptedData), nil)

	if decryptError != nil {
		err = decryptError
		return
	}

	decryptedData = string(byteDecryptedData[:])

	return
}
