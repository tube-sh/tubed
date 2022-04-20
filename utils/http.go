package utils

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func InitHttpClient() *http.Client {
	// variables
	proxy := os.Getenv("http_proxy")
	skipTLSVerify := os.Getenv("skiptlsverify")
	tr := &http.Transport{}

	// configure proxy if requested
	if proxy != "" {
		proxyUrl, _ := url.Parse(proxy)
		tr.Proxy = http.ProxyURL(proxyUrl)
	}

	// configure skip TLS if requested
	if skipTLSVerify != "" {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{Transport: tr}
	return client
}

func PostJSON(url string, jsonBody []byte) []byte {

	// Create a new request using http
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Send req using http Client
	client := InitHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response:", err)
	}

	// read response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}

	return body
}
