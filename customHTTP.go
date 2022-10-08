package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func sendHTTPRequest(url string, method string, data map[string]any, headers map[string]string) (string, error) {
	// Convert Request Data to JSON
	jsonData, jsonConversionError := json.Marshal(data)

	if jsonConversionError != nil {
		return "", errors.New("Error while converting Request Data into JSON: " + jsonConversionError.Error())
	}

	httpBodyReader := bytes.NewReader([]byte(jsonData))

	// Prepare request
	request, requestError := http.NewRequest(strings.ToUpper(method), url, httpBodyReader)

	if requestError != nil {
		return "", errors.New("Error while Preparing Request: " + requestError.Error())
	}

	if len(headers) > 0 {
		// Set request headers
		for key, value := range headers {
			request.Header.Add(key, value)
		}
	}

	// Send request
	response, responseError := http.DefaultClient.Do(request)

	if responseError != nil {
		return "", errors.New("Error in Response: " + responseError.Error())
	}

	// Read response
	responseBody, responseBodyError := io.ReadAll(response.Body)

	if responseBodyError != nil {
		return "", errors.New("Error while Reading Response: " + responseBodyError.Error())
	}

	responseStr := string(responseBody)

	return responseStr, nil
}
