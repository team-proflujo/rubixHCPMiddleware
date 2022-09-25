package main

import (
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

func fromJSON(jsonStr string) (map[string]any, error) {
	var data map[string]any

	jsonDecodeError := json.Unmarshal([]byte(jsonStr), &data)

	if jsonDecodeError != nil {
		return nil, errors.New("Error while decoding to JSON: " + jsonDecodeError.Error())
	}

	return data, nil
}

func getDIDInfo() (map[string]any, error) {
	var didInfo = map[string]any{}
	didFilePath := "Rubix/DATA/DID.json"

	homeDir, homeDirError := os.UserHomeDir()

	if homeDirError != nil {
		return nil, errors.New("Error while trying to get Home Directory path: " + homeDirError.Error())
	}

	didFilePath = filepath.Join(homeDir, didFilePath)

	didFileContent, fileReadError := readFile(didFilePath)

	if fileReadError != nil {
		return nil, errors.New("Error while trying to get DID.json file content: " + fileReadError.Error())
	} else if didFileContent == nil {
		return nil, errors.New("DID.json file is empty!")
	}

	didInfo, decodeJsonError := fromJSON(string(didFileContent))

	if decodeJsonError != nil {
		return nil, errors.New("Error while trying to parse DID.json: " + decodeJsonError.Error())
	}

	/*if reflect.ValueOf(rawDidInfo).Kind() == reflect.Array {
		didInfo = rawDidInfo[0]
	}*/

	return didInfo, nil
}
