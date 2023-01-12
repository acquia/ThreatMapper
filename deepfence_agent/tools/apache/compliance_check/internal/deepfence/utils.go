package deepfence

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"time"
)

const UseHttpENV = "USE_HTTP"

func buildHttpClient() (*http.Client, error) {
	transport := &http.Transport{
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 1024,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Minute,
			KeepAlive: 15 * time.Minute,
		}).DialContext,

		ResponseHeaderTimeout: 5 * time.Minute,
	}
	if !useHttp() {
		// Set up our own certificate pool
		tlsConfig := &tls.Config{RootCAs: x509.NewCertPool(), InsecureSkipVerify: true}
		transport.TLSClientConfig = tlsConfig
		transport.TLSHandshakeTimeout = 30 * time.Second
	}

	client := &http.Client{
		Transport: transport,
		Timeout: 15 * time.Minute,
	}

	return client, nil
}

type dfApiAuthResponse struct {
	Data struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	} `json:"data"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
	Success bool `json:"success"`
}

func getProtocol() string {
	if useHttp() {
		return "http"
	} else {
		return "https"
	}
}

func useHttp() bool {
	useInsecure := os.Getenv(UseHttpENV)
	if useInsecure != "" {
		return false
	} else {
		return true
	}
}