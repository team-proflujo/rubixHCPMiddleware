package main

import (
	"team-proflujo/rubixHCPMiddleware/globalVars"

	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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

func base64Decode(encodedData string) ([]byte, error) {
	rawData, decodeError := base64.StdEncoding.DecodeString(encodedData)

	if decodeError != nil {
		return nil, errors.New("Error while decoding base64 data: " + decodeError.Error())
	}

	return rawData, nil
}

func toJSON(data any) (string, error) {
	var jsonStr string

	jsonBytes, jsonEncodeError := json.Marshal(data)

	if jsonEncodeError != nil {
		return "", errors.New("Error while encoding to JSON: " + jsonEncodeError.Error())
	}

	jsonStr = string(jsonBytes)

	return jsonStr, nil
}

func getDIDInfo() (globalVars.DIDInfoStruct, error) {
	var didInfo globalVars.DIDInfoStruct
	didFilePath := "Rubix/DATA/DID.json"

	homeDir, homeDirError := os.UserHomeDir()

	if homeDirError != nil {
		return didInfo, errors.New("Error while trying to get Home Directory path: " + homeDirError.Error())
	}

	didFilePath = filepath.Join(homeDir, didFilePath)

	didFileContent, fileReadError := readFile(didFilePath)

	if fileReadError != nil {
		return didInfo, errors.New("Error while trying to get DID.json file content: " + fileReadError.Error())
	} else if didFileContent == nil {
		return didInfo, errors.New("DID.json file is empty!")
	}

	var didInfoList []globalVars.DIDInfoStruct

	decodeJsonError := json.Unmarshal(didFileContent, &didInfoList)

	if decodeJsonError != nil {
		return didInfo, errors.New("Error while trying to parse DID.json: " + decodeJsonError.Error())
	}

	if len(didInfoList) > 0 {
		didInfo = didInfoList[0]
	}

	return didInfo, nil
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

	configFilePath := filepath.Join(scriptPath, "config.json")

	configFileContent, fileReadError := readFile(configFilePath)

	if fileReadError != nil {
		return configData, errors.New("Error while trying to get config.json file content: " + fileReadError.Error())
	} else if configFileContent == nil {
		return configData, errors.New("config.json file is empty!")
	}

	decodeJsonError := json.Unmarshal(configFileContent, &configData)

	if decodeJsonError != nil {
		return configData, errors.New("Error while trying to parse config.json: " + decodeJsonError.Error())
	}

	return configData, nil
}
