package main

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func awsNewClient() (s3Client *s3.S3) {
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(globalVars.AppConfig.AWSStorageConfig.AccessKey, globalVars.AppConfig.AWSStorageConfig.Secret, ""),
		Endpoint:    aws.String(globalVars.AppConfig.AWSStorageConfig.APIEndpoint),
		Region:      aws.String(globalVars.AppConfig.AWSStorageConfig.Region),
	}

	newSession := session.New(s3Config)
	s3Client = s3.New(newSession)

	return
}

func awsRegisterWallet(reqData globalVars.AppRegisterMethodReqDataStruct) (response globalVars.APPHTTPResponse) {
	// Initialize Map before using it (otherwise, it would be nil)
	response.Data = map[string]any{}

	if len(reqData.Password) == 0 {
		response.Message = "Password must not be empty!"
		return
	}

	didInfo, didInfoError := getDIDInfo()

	if didInfoError != nil {
		response.Message = "Error while trying to get DID Info"
		response.Error = didInfoError.Error()
		return
	}

	walletData, prepWalletDataErr := prepareWalletDataToRegister(didInfo, true, reqData.Password)

	if prepWalletDataErr != nil {
		response.Message = "Error occurred while preparing Wallet Data to Register"
		response.Error = prepWalletDataErr.Error()
		return
	}

	s3Client := awsNewClient()

	privateSharePngObj := s3.PutObjectInput{
		Bucket: aws.String(globalVars.AppConfig.AWSStorageConfig.Bucket),
		Key:    aws.String(filepath.Join(didInfo.DidHash, "PrivateShare.png.encrypted")),
		Body:   strings.NewReader(walletData.PrivateSharePng),
	}

	_, privateSharePngUploadErr := s3Client.PutObject(&privateSharePngObj)

	if privateSharePngUploadErr != nil {
		response.Message = "Error when Uploading PrivateShare.png to Storage API"
		response.Error = privateSharePngUploadErr.Error()
		return
	}

	didPngObj := s3.PutObjectInput{
		Bucket: aws.String(globalVars.AppConfig.AWSStorageConfig.Bucket),
		Key:    aws.String(filepath.Join(didInfo.DidHash, "DID.png.encrypted")),
		Body:   strings.NewReader(walletData.DIDPng),
	}

	_, didPngUploadErr := s3Client.PutObject(&didPngObj)

	if didPngUploadErr != nil {
		response.Message = "Error when Uploading DID.png to Storage API"
		response.Error = didPngUploadErr.Error()
		return
	}

	privateKeyPemObj := s3.PutObjectInput{
		Bucket: aws.String(globalVars.AppConfig.AWSStorageConfig.Bucket),
		Key:    aws.String(filepath.Join(didInfo.DidHash, "privatekey.pem.encrypted")),
		Body:   strings.NewReader(walletData.DIDPng),
	}

	_, privateKeyPemUploadErr := s3Client.PutObject(&privateKeyPemObj)

	if privateKeyPemUploadErr != nil {
		response.Message = "Error when Uploading DID.png to Storage API"
		response.Error = privateKeyPemUploadErr.Error()
		return
	}

	postRegisterErr := postRegisterAction(didInfo)

	if postRegisterErr != nil {
		response.Message = "Error occurred after Registration"
		response.Error = postRegisterErr.Error()
		return
	}

	response.Success = true
	response.Message = "Wallet has been successfully registered with Storage API."

	return
}

func awsDownloadFile(s3Client *s3.S3, didInfo globalVars.DIDInfoStruct, filePath string, password string) (data string, err error) {
	object := &s3.GetObjectInput{
		Bucket: aws.String(globalVars.AppConfig.AWSStorageConfig.Bucket),
		Key:    aws.String(filepath.Join(didInfo.DidHash, filePath)),
	}

	downloadResult, downloadError := s3Client.GetObject(object)

	if downloadError != nil {
		err = downloadError
		return
	}

	bufferReader := new(strings.Builder)

	_, readFileContentError := io.Copy(bufferReader, downloadResult.Body)

	if readFileContentError != nil {
		err = readFileContentError
		return
	}

	encryptedData := bufferReader.String()

	if len(encryptedData) == 0 {
		err = errors.New("File download failed")
		return
	}

	decryptedData, aesDecryptError := aesDecrypt(encryptedData, password)

	if aesDecryptError != nil {
		err = aesDecryptError
		return
	}

	if len(decryptedData) == 0 {
		err = errors.New("Decryption of file failed")
		return
	}

	data = decryptedData

	return
}

func awsGetWalletData(reqData globalVars.AppRegisterMethodReqDataStruct) (response globalVars.APPHTTPResponse) {
	var walletData globalVars.WalletDataInStorage

	if len(reqData.Password) == 0 {
		response.Message = "Password must not be empty!"
		return
	}

	didInfo, didInfoError := getDIDInfo()

	if didInfoError != nil {
		response.Message = "Error while trying to get DID Info"
		response.Error = didInfoError.Error()
		return
	}

	s3Client := awsNewClient()

	privateSharePngContent, privateSharePngError := awsDownloadFile(s3Client, didInfo, "PrivateShare.png.encrypted", reqData.Password)

	if privateSharePngError != nil {
		response.Message = "Error when downloading PrivateShare.png from Storage API"
		response.Error = privateSharePngError.Error()
		return
	}

	didPngContent, didPngError := awsDownloadFile(s3Client, didInfo, "DID.png.encrypted", reqData.Password)

	if didPngError != nil {
		response.Message = "Error when downloading DID.png from Storage API"
		response.Error = didPngError.Error()
		return
	}

	privateKeyPemContent, privateKeyPemError := awsDownloadFile(s3Client, didInfo, "privatekey.pem.encrypted", reqData.Password)

	if privateKeyPemError != nil {
		response.Message = "Error when downloading privatekey.pem from Storage API"
		response.Error = privateKeyPemError.Error()
		return
	}

	walletData.DIDHash = didInfo.DidHash
	walletData.PeerId = didInfo.PeerId
	walletData.PrivateSharePng = privateSharePngContent
	walletData.DIDPng = didPngContent
	walletData.PrivateKeyPem = privateKeyPemContent

	response.Data = walletData
	response.Success = true

	return
}
