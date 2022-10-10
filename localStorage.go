package main

import (
	"errors"
	"os"
	"path/filepath"
	"team-proflujo/rubixHCPMiddleware/globalVars"
)

func localStorageIsRegistered(storageBasePath string) (registered bool) {
	registered = isDir(storageBasePath)

	return
}

func localStorageRegisterWallet(reqData globalVars.AppRegisterMethodReqDataStruct) (response globalVars.APPHTTPResponse) {
	// Initialize Map before using it (otherwise, it would be nil)
	response.Data = map[string]any{}

	if len(reqData.Password) == 0 {
		response.Message = "Password must not be empty!"
		return
	}

	if len(reqData.StorageLocation) == 0 {
		response.Message = "Storage Location must not be empty!"
		return
	} else if !isDir(reqData.StorageLocation) {
		response.Message = "The given Storage Location does not exists!"
		return
	}

	didInfo, didInfoJsonStr, didInfoError := getDIDInfo()

	if didInfoError != nil {
		response.Message = "Error while trying to get DID Info"
		response.Error = didInfoError.Error()
		return
	}

	storageBasePath := filepath.Join(reqData.StorageLocation, didInfo.DidHash)

	if localStorageIsRegistered(storageBasePath) {
		response.Message = "Already Registered"
		response.Success = true
		return
	}

	walletData, prepWalletDataErr := prepareWalletDataToRegister(didInfo, true, reqData.Password)

	if prepWalletDataErr != nil {
		response.Message = "Error occurred while preparing Wallet Data to Register"
		response.Error = prepWalletDataErr.Error()
		return
	}

	storageBasePathError := os.MkdirAll(storageBasePath, os.ModePerm)

	if storageBasePathError != nil {
		response.Message = "Error when creating directory at " + storageBasePath
		response.Error = storageBasePathError.Error()
		return
	}

	didInfoJsonMoved, didInfoJsonError := writeFile(filepath.Join(storageBasePath, "DID.json"), []byte(didInfoJsonStr))

	if didInfoJsonError != nil {
		response.Message = "Error while moving DID.json to Storage location"
		response.Error = didInfoJsonError.Error()
		return
	} else if !didInfoJsonMoved {
		response.Message = "Moving DID.json to Storage location failed"
		return
	}

	privateSharePngMoved, privateSharePngError := writeFile(filepath.Join(storageBasePath, "PrivateShare.png.encrypted"), []byte(walletData.PrivateSharePng))

	if privateSharePngError != nil {
		response.Message = "Error while moving PrivateShare.png to Storage location"
		response.Error = privateSharePngError.Error()
		return
	} else if !privateSharePngMoved {
		response.Message = "Moving PrivateShare.png to Storage location failed"
		return
	}

	didPngMoved, didPngError := writeFile(filepath.Join(storageBasePath, "DID.png.encrypted"), []byte(walletData.DIDPng))

	if didPngError != nil {
		response.Message = "Error while moving DID.png to Storage location"
		response.Error = didPngError.Error()
		return
	} else if !didPngMoved {
		response.Message = "Moving DID.png to Storage location failed"
		return
	}

	privateKeyPemMoved, privateKeyPemError := writeFile(filepath.Join(storageBasePath, "privatekey.pem.encrypted"), []byte(walletData.PrivateKeyPem))

	if privateKeyPemError != nil {
		response.Message = "Error while moving privatekey.pem to Storage location"
		response.Error = privateKeyPemError.Error()
		return
	} else if !privateKeyPemMoved {
		response.Message = "Moving privatekey.pem to Storage location failed"
		return
	}

	postRegisterErr := postRegisterAction(didInfo)

	if postRegisterErr != nil {
		response.Message = "Error occurred after Registration"
		response.Error = postRegisterErr.Error()
		return
	}

	globalVars.AppConfig.WalletRegisteredToStorage = true

	response.Success = true
	response.Message = "Wallet Data have successfully been moved to Storage Location."

	return
}

func localStorageGetFileContent(storageBasePath string, filePath string, password string) (fileContent string, err error) {
	encryptedFileContent, fileReadError := readFile(filepath.Join(storageBasePath, filePath))

	if fileReadError != nil {
		err = fileReadError
		return
	}

	decryptedFileContent, decryptError := aesDecrypt(string(encryptedFileContent[:]), password)

	if decryptError != nil {
		err = decryptError
		return
	} else if len(decryptedFileContent) == 0 {
		err = errors.New("Decryption of file failed")
		return
	}

	fileContent = decryptedFileContent

	return
}

func localStorageGetWalletData(reqData globalVars.AppRegisterMethodReqDataStruct) (response globalVars.APPHTTPResponse) {
	var walletData globalVars.WalletDataInStorage

	if len(reqData.Password) == 0 {
		response.Message = "Password must not be empty!"
		return
	}

	if len(reqData.StorageLocation) == 0 {
		response.Message = "Storage Location must not be empty!"
		return
	} else if !isDir(reqData.StorageLocation) {
		response.Message = "The given Storage Location does not exists!"
		return
	}

	didInfo, _, didInfoError := getDIDInfo()

	if didInfoError != nil {
		response.Message = "Error while trying to get DID Info"
		response.Error = didInfoError.Error()
		return
	}

	storageBasePath := filepath.Join(reqData.StorageLocation, didInfo.DidHash)

	storageBasePathError := os.MkdirAll(storageBasePath, os.ModePerm)

	if storageBasePathError != nil {
		response.Message = "Error when creating directory at " + storageBasePath
		response.Error = storageBasePathError.Error()
		return
	}

	privateSharePngContent, privateSharePngError := localStorageGetFileContent(storageBasePath, "PrivateShare.png.encrypted", reqData.Password)

	if privateSharePngError != nil {
		response.Message = "Error while retrieving PrivateShare.png from Storage location"
		response.Error = privateSharePngError.Error()
		return
	}

	didPngContent, didPngError := localStorageGetFileContent(storageBasePath, "DID.png.encrypted", reqData.Password)

	if didPngError != nil {
		response.Message = "Error while retrieving DID.png from Storage location"
		response.Error = didPngError.Error()
		return
	}

	privateKeyPemContent, privateKeyPemError := localStorageGetFileContent(storageBasePath, "privatekey.pem.encrypted", reqData.Password)

	if privateKeyPemError != nil {
		response.Message = "Error while retrieving privatekey.pem from Storage location"
		response.Error = privateKeyPemError.Error()
		return
	}

	walletData.DIDInfo = didInfo
	walletData.PrivateSharePng = privateSharePngContent
	walletData.DIDPng = didPngContent
	walletData.PrivateKeyPem = privateKeyPemContent

	response.Data = walletData
	response.Success = true

	return
}
