[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)
[![Community Support](https://badgen.net/badge/support/community/cyan?icon=awesome)](https://github.com/nginxinc/ciroque-nginx/nginx-with-mutual-tls/blob/main/SUPPORT.md)
[![Project Status: Concept – Minimal or no implementation has been done yet, or the repository is only intended to be a limited example, demo, or proof-of-concept.](https://www.repostatus.org/badges/latest/concept.svg)](https://www.repostatus.org/#concept)

# nginx_with_mutual_tls

## Overview

As part of a Jazz Exploration in D-minor, this repo shows how to configure mutual TLS (mTLS) to secure communication between NGINX Plus and a client written in Go.

The primary impetus for this exploration is to bolster the [NGINX LoadBalancer for Kubernetes](https://github.com/nginxinc/nginx-loadbalancer-kubernetes) project with mTLS support.

A primary objective of this exploration is to make use of self-signed certificates. While the provided steps may
work with certificates signed by a Certificate Authority (CA), it is not a goal of this project. It is not tested, nor is it a design goal. 
Therefore, you would need to test it for yourself.

## Requirements

This project requires Go 1.19.4 or later.

Use `asdf`? If so, you can run `asdf install` from the root of the project, and you'll be good to go.

Oh. You will also need a running NGINX Plus. You can [get a trial license](https://www.nginx.com/free-trial-request/); 
and follow the [installation guide](https://docs.nginx.com/nginx/admin-guide/installing-nginx/installing-nginx-plus/) to meet this requirement.

## Getting Started

Clone the repo:

```bash 
git clone git@github.com:ciroque-nginx/nginx-with-mutual-tls.git
```

Then:

1. Review and edit the [configuration files](#configuration-files),
1. Follow the [certificate generation instructions](#generate-certificates),
1. Follow the [NGINX Plus configuration instructions](#configure-nginx-plus),
1. Follow the [client configuration instructions](#configure-client). 

### Configuration Files

There are three configuration files that need to be edited to match your environment.
- [ca.cnf](#cacnf)
- [server.cnf](#servercnf)
- [client.cnf](#clientcnf)

These files are located in the `tls` directory. Each of the files will need to be updated with your desired
distinguished name (DN) information -- which is used to identify the certificate owner. The fields are:
- C: Country
- ST: State
- L: Locality
- O: Organization
- OU: Organizational Unit
- CN: Common Name

More information about the DN fields can be found [here](https://www.cryptosys.net/pki/manpki/pki_distnames.html).

It should be noted that the CN field is deprecated, and the Subject Alternative Name (SAN) should be used instead. 
To include the SAN, the `alt_names` section of the configuration file will need to be updated. The `alt_names` includes either IP Addresses or hostnames. 
IP Addresses use the format: `IP.i = n.n.n.n` where `i` is the desired index in the list. 
Hostnames use the format: `DNS.i = hostname` where `i` is the desired index in the list.
The `server.cnf` file has examples of both IP Addresses and hostnames.

If multiple NGINX Plus instances are being used, the DNS entries can use wildcards. For example, `*.example.com` would match `foo.example.com` and `bar.example.com`.

The required updates to each file are detailed below.

#### ca.cnf

This file only needs to have the DN information updated.

#### server.cnf

This file needs to have the DN information updated, and the SAN information (DNS / IP entries in the `alt_names` section) added / updated.

#### client.cnf

This file only needs to have the DN information updated.

### Generate Certificates

In order to use self-signed certificates, a Certificate Authority (CA) is required. The `openssl` utility will be used to generate the CA and the certificates.
This will then be used to sign the certificates for the client and the server. The CA certificate will then need to be 
provided to both the client and NGINX Plus to allow authentication.

#### Generate the CA

```bash
openssl req -newkey rsa:2048 -nodes -x509 -config ca.cnf -out ca.crt -keyout ca.key
```

#### Generate the Server Certificate

Ensure the `alt_names` section of the `server.cnf` file has been updated with the desired DNS entries, use wildcards where appropriate.

```bash
openssl genrsa -out server.key 2048
openssl req -new -key server.key -config server.cnf -out server.csr
```

#### Sign the Server Certificate

```bash
openssl x509  -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256 -extensions req_ext -extfile server.cnf
```

#### Verify the Server Certificate has the SAN

Look for the DNS / IP entries in the `Subject Alternative Name` section.

```bash
openssl x509 -in server.crt -text -noout | grep -A 1 "Subject Alternative Name"
```

#### Generate the Client Certificate

```bash 
openssl genrsa -out client.key 2048
openssl req -new -key client.key -config client.cnf -out client.csr
```

#### Sign the Client Certificate

```bash
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256 -extfile client.cnf -extensions client
```

#### Verify the Client Certificate has the correct extendedKeyUsage

Look for `SSL client : Yes` in the output.

```bash
openssl x509 -in client.crt -noout -purpose | grep 'SSL client :'
```

### Configure NGINX Plus

#### Configuring the Server certificate

Copy the necessary server files (`server.crt`, `server.key`, and `ca.crt`) to the NGINX host; place the files in the `/etc/ssl/certs/nginx` directory.

In the `/etc/nginx/nginx.conf` file, add the following to the `http` or `server` section (refer to the [documentation](https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_certificate) for details):

```bash
http {
  ssl_certificate       /etc/ssl/certs/nginx/server.crt;
  ssl_certificate_key   /etc/ssl/certs/nginx/server.key;
}
```

For a client to validate this certificate, it will need the CA certificate. See the [Configure Client](#configure-client) section for details.

#### Configuring NGINX Plus to require client certificates

In the `/etc/nginx/nginx.conf` file, add the following to the `http` or `server` section (refer to the [documentation](https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_client_certificate) for details):

```bash
http {
  ssl_client_certificate    /etc/ssl/certs/nginx/ca.crt;
  ssl_verify_client         on;
  ssl_verify_depth          3;
}
```

Restart NGINX Plus to apply the changes.

```bash
nginx -s reload
```

Test with curl:

```bash
curl --cert client.crt --key client.key --cacert ca.crt https://<your-host>/api
```

### Configure Client

A simple client to test the certs can be found in the `client/main.go` file. 
The client uses the `tls/ca.crt` file to validate the server certificate; 
the `tls/client.crt` and `tls/client.key` files are used to authenticate the client.

The client makes use of the CA certificate by loading the `tls/ca.crt` file and adding it to a Certificate Pool:

```go
caCert, err := os.ReadFile("tls/ca.crt")
if err != nil {
    log.Fatalf("could not open certificate file: %v", err)
}
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)
```

The client likewise uses the client certificate and key by loading the `tls/client.crt` and `tls/client.key` files:

```go
cert, err := tls.LoadX509KeyPair("tls/client.crt", "tls/client.key")
if err != nil {
    log.Fatalf("could not load client key pair: %v", err)
}
```

Once these files are loaded, the `tls.Config` object is initialized with the `RootCAs` and `Certificates`; 
additionally, `InsecureSkipVerify` is set to false, ensuring the server certificate is validated:

```go
tlsConfig := &tls.Config{
    RootCAs:            caCertPool,
    Certificates:       []tls.Certificate{cert},
    InsecureSkipVerify: false,
}
```

To run the client, execute the following command:

```bash
NGINX_PLUS_API_ENDPOINT=<your-host> go run client/main.go
```

If successful, the client will print the NGINX Info response. Otherwise, an error will be printed.

## Contributing

Please see the [contributing guide](https://github.com/ciroque-nginx/nginx-with-mutual-tls/blob/main/CONTRIBUTING.md) for guidelines on how to best contribute to this project.

## License

[Apache License, Version 2.0](https://github.com/ciroque-nginx/nginx-with-mutual-tls/blob/main/LICENSE)

## References
[What is mutual TLS (mTLS)?](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/)

[How to MTLS in golang](https://kofo.dev/how-to-mtls-in-golang)

[A step-by-step guide to mTLS in Go](https://venilnoronha.io/a-step-by-step-guide-to-mtls-in-go)

[TLS mutual authentication with golang and nginx](https://medium.com/rahasak/tls-mutual-authentication-with-golang-and-nginx-937f0da22a0e)
