package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	maxIdleConnsPerHost = 1024
)


func buildHttpClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	if !UseHttp() {
		// Set up our own certificate pool
		tlsConfig := &tls.Config{RootCAs: x509.NewCertPool(), InsecureSkipVerify: true}
		transport.TLSClientConfig = tlsConfig
		transport.TLSHandshakeTimeout = 30 * time.Second
	}
	client := &http.Client{Transport: transport}
	return client
}

func authenticateAgentWithConsole(httpClient *http.Client, scopeApiUrl, authKey string) (bool, error) {
	req, err := http.NewRequest("GET", scopeApiUrl, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("deepfence-key", authKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	body, err := io.ReadAll(resp.Body)
	if err == nil {
		fmt.Printf("agent authentication: got status code %d with message %s\n", resp.StatusCode,
			string(body)[:int(math.Min(160.0, float64(len(body))))])
	}
	return false, nil
}

const UseHttpENV = "USE_HTTP"

func GetProtocol() string {
	if UseHttp() {
		return "http"
	} else {
		return "https"
	}
}

func UseHttp() bool {
	useInsecure := os.Getenv(UseHttpENV)
	if useInsecure != "" {
		return false
	} else {
		return true
	}
}

func main() {
	authKey := os.Getenv("DEEPFENCE_KEY")
	mgmtConsoleUrl := os.Getenv("MGMT_CONSOLE_URL")
	consolePort := os.Getenv("MGMT_CONSOLE_PORT")
	if consolePort != "" && consolePort != "443" {
		mgmtConsoleUrl += ":" + consolePort
	}
	scopeApiUrl := fmt.Sprintf("%s://%s/topology-api", GetProtocol(), mgmtConsoleUrl)
	var httpClient *http.Client
	for {
		if httpClient == nil {
			httpClient = buildHttpClient()
		}
		authenticated, err := authenticateAgentWithConsole(httpClient, scopeApiUrl, authKey)
		if err != nil {
			fmt.Println("Could not connect to Deepfence Management Console. Retrying...")
			time.Sleep(10 * time.Second)
			continue
		}
		if authenticated == true {
			return
		} else {
			os.Exit(1)
		}
	}
}
