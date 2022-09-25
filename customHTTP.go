package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func sendRequest(url string, method string, data map[string]any, headers map[string]string, isJSON bool) (any, error) {
	jsonData, jsonConversionError := json.Marshal(data)

	if jsonConversionError != nil {
		return "", errors.New("Error while converting Request Data into JSON: " + jsonConversionError.Error())
	}

	httpBodyReader := bytes.NewReader([]byte(jsonData))

	request, requestError := http.NewRequest(strings.ToUpper(method), url, httpBodyReader)

	if requestError != nil {
		return "", errors.New("Error while Preparing Request: " + requestError.Error())
	}

	if headers != nil {
		for key, value := range headers {
			request.Header.Add(key, value)
		}
	}

	response, responseError := http.DefaultClient.Do(request)

	if responseError != nil {
		return "", errors.New("Error in Response: " + responseError.Error())
	}

	responseBody, responseBodyError := io.ReadAll(response.Body)

	if responseBodyError != nil {
		return "", errors.New("Error while Reading Response: " + responseBodyError.Error())
	}

	if isJSON {
		var jsonData = map[string]any{}
		jsonDecodeError := json.Unmarshal(responseBody, &jsonData)

		if jsonDecodeError != nil {
			return "", errors.New("Error while Decoding Response to JSON: " + jsonDecodeError.Error())
		}

		return jsonData, nil
	}

	responseStr := string(responseBody)

	return responseStr, nil
}
