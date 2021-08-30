[![Go Report Card](https://goreportcard.com/badge/github.com/kubevirt-ui/kube-gateway-operator)](https://goreportcard.com/report/github.com/kubevirt-ui/kube-gateway-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# kube-gateway-operator

![alt gopher network](https://raw.githubusercontent.com/kubevirt-ui/kube-gateway/main/docs/network-side.png)

kube-gateway-operator installs and operate [kube-gateway](https://github.com/kubevirt-ui/kube-gateway)
## Build

```bash

IMG=quay.io/kubevirt-ui/kube-gateway-operator make podman-build
```

## Usage

Requesting a token for [kube-gateway](https://github.com/kubevirt-ui/kube-gateway) service is done using GateToken CRD,

Available fields are:

- user-id: string (required), user-id is the user id of the user requesting this token.
- match-path: string (required), match-path is a regular expresion used to validate API request path, API requests matching this pattern will be validated by the token. This field may not be empty.
- match-method: string, a comma separeted list of allowed http methods, defoult is "GET,OPTIONS"
- duration-sec: int, duration-sec is the duration in sec the token will be validated since it's invocation. Defalut value is 3600s (1h).
- from: string, from is time of token invocation, the token will not validate before this time, the token duration will start from this time. Defalut to token object creation time.

Creating a token requires a secret holding a RSA private-key for sighing the token in the namespace of the token (secret name: kube-gateway-jwt-secret), nce token is ready it will be available in the GateToken status.

Get a token:

[![asciicast](https://asciinema.org/a/397136.svg)](https://asciinema.org/a/397136)

Deploy:

[![asciicast](https://asciinema.org/a/397137.svg)](https://asciinema.org/a/397137)

(gopher network image - [egonelbre/gophers](https://github.com/egonelbre/gophers))

## Deploy

Requires

- requires GOLANG ver >= v1.15 dev env.
- user with admin permisions logged into the cluster.

```bash
# Deploy the operator, RBAC roles and CRDs
export USERNAME=yaacov
make deploy IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1
```

```bash
# Remove deployment of the operator, RBAC roles and CRDs
export USERNAME=yaacov
make undeploy IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1
```

## Create GateToken CR

Requires a secret with private key on 'kube-gateway' namespace:

```bash
# Use the kube-gateway namespace
oc project kube-gateway

# create a secret
oc create -n kube-gateway-operator-system secret generic kube-gateway-jwt-secret --from-file=test/cert.pem --from-file=test/key.pem

# create a token request
oc create -f config/samples/kubegateway_v1beta1_gatetoken.yaml

# check the token
oc get gatetoken gatetoken-sample -o yaml
```

## Build

```bash
# Compile operator
make

# Install CRD on cluser for running loaclly
make install
# make uninstall

# Run locally
make run
```

## Build images

```bash
export USERNAME=yaacov
make podman-build podman-push IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1
```
