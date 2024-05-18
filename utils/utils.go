package utils

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"moul.io/http2curl/v2"
)

// DoHTTPRequest Sends generic http request
func DoHTTPRequest(method string, url string, headers map[string]string, body []byte) (responseBody string, err error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	req.Close = true
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	command, _ := http2curl.GetCurlCommand(req)

	response, err := httpClient.Do(req)
	if err != nil {
		log.Printf("ERROR error requesting with http: %s, error: %v\n", command, err)
		err = errors.New("failed_do_request")
		return
	}
	bodyBytes, err := ioutil.ReadAll(response.Body)
	response.Body.Close()

	if err != nil {
		log.Printf("ERROR error requesting with http: %s, error: %v\n", command, err)
		err = errors.New("failed_read_body")
		return
	}

	responseBody = string(bodyBytes)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		log.Printf("ERROR error requesting with http: %s, status code: %v, response: %s\n", command, response.StatusCode, responseBody)
		err = errors.New("invalid_status")
		return
	}

	return
}

// FileExists checks if a file exists and is not a directory before we
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// GetFileSize gets the fie size
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) || info.IsDir() {
		return 0, errors.New("file_not_found")
	}
	return info.Size(), nil
}
