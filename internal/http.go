package internal

import (
	"crypto/tls"
	"net/http"
	"net/url"
)

func initHttpClientNew(proxy string) *http.Client {
	var tr *http.Transport
	if proxy != "" {
		proxyUrl, _ := url.Parse(proxy)
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxyUrl),
		}
	}

	client := &http.Client{Transport: tr}
	return client
}
