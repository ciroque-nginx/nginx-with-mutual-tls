package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	nginxClient "github.com/nginxinc/nginx-plus-go-client/client"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	caCert, err := ioutil.ReadFile("tls/ca.crt")
	if err != nil {
		log.Fatalf("could not open certificate file: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair("tls/client.crt", "tls/client.key")
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}

	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = tlsConfig

	httpClient := &http.Client{
		Transport: transport,
	}

	apiEndpoint := os.Getenv("NGINX_PLUS_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://localhost/api"
	}

	ngxClient, err := nginxClient.NewNginxClient(httpClient, apiEndpoint)
	if err != nil {
		panic(err)
	}

	nginxInfo, err := ngxClient.GetNginxInfo()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", nginxInfo)
}
