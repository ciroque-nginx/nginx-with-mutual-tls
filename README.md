[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)
[![Community Support](https://badgen.net/badge/support/community/cyan?icon=awesome)](https://github.com/nginxinc/ciroque-nginx/nginx-with-mutual-tls/blob/main/SUPPORT.md)
[![Project Status: Concept – Minimal or no implementation has been done yet, or the repository is only intended to be a limited example, demo, or proof-of-concept.](https://www.repostatus.org/badges/latest/concept.svg)](https://www.repostatus.org/#concept)

# nginx_with_mutual_tls

## Overview

As part of a Jazz Exploration in D-minor, this repo shows how to configure mutual TLS (mTLS) to secure communication between NGINX Plus and a client written in Go.

The primary impetus for this exploration is to bolster the [NGINX LoadBalancer for Kubernetes](https://github.com/nginxinc/nginx-loadbalancer-kubernetes) project with mTLS support.

A primary objective of this exploration is to make use of self-signed certificates. While the provided steps may
work with certificates signed by a Certificate Authority (CA), it is not a goal of this project.

## Requirements

This project simply requires Go 1.19.4 or later.

Use `asdf`? If so, you can simply run `asdf install` from the root of the project, and you'll be good to go.

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

It should be noted that the CN field is deprecated, and the Subject Alternative Name (SAN) should be used instead. The
SAN will be described in more detail in the [Server Certificate generation section](#generate-the-server-certificate).

More information about the DN fields can be found [here](https://www.cryptosys.net/pki/manpki/pki_distnames.html).

The required updates to each file are detailed below.

#### ca.cnf

This file only needs to have the DN information updated.

#### server.cnf

This file needs to have the DN information updated, and the SAN information (DNS / IP entries in the `alt_names` section) added / updated.

#### client.cnf

This file only needs to have the DN information updated.

### Generate Certificates

In order to use self-signed certificates, we'll need a Certificate Authority (CA). We'll use OpenSSL to generate the CA and the certificates.
This will then be used to sign the certificates for the client and the server. The CA certificate will then need to be 
provided to both the client and NGINX Plus to allow authentication.

#### Generate the CA

```bash
openssl req -newkey rsa:2048 -nodes -x509 -config ca.cnf -out ca.crt -keyout ca.key
```

#### Generate the Server Certificate

```bash
openssl genrsa -out server.key 2048
openssl req -new -key server.key -config server.cnf -out server.csr
```

#### Sign the Server Certificate

```bash
openssl x509  -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256 -extensions req_ext -extfile server.cnf
```

#### Verify the Server Certificate has the SAN

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

```bash
openssl x509 -in client.crt -noout -purpose | grep 'SSL client :'
```

### Configure NGINX Plus

### Configure Client

## Contributing

Please see the [contributing guide](https://github.com/ciroque-nginx/nginx-with-mutual-tls/blob/main/CONTRIBUTING.md) for guidelines on how to best contribute to this project.

## License

[Apache License, Version 2.0](https://github.com/ciroque-nginx/nginx-with-mutual-tls/blob/main/LICENSE)

## References
[What is mutual TLS (mTLS)?](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/)

[How to MTLS in golang](https://kofo.dev/how-to-mtls-in-golang)

[A step-by-step guide to mTLS in Go](https://venilnoronha.io/a-step-by-step-guide-to-mtls-in-go)

[TLS mutual authentication with golang and nginx](https://medium.com/rahasak/tls-mutual-authentication-with-golang-and-nginx-937f0da22a0e)
