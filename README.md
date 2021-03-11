# oc-gate-operator

![alt gopher network](https://raw.githubusercontent.com/yaacov/oc-gate/main/web/public/network-side.png)

creates tokens for the [oc-gate](https://github.com/yaacov/oc-gate) service

[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/oc-gate-operator)](https://goreportcard.com/report/github.com/yaacov/oc-gate-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

(gopher network image - [egonelbre/gophers](https://github.com/egonelbre/gophers))

## Install

``` bash
# Clone repository
git clone git@github.com:yaacov/oc-gate-operator.git
cd oc-gate-operator

# Add the private/public key secret used to generate tokens
oc create -n oc-gate-operator-system secret generic oc-gate-jwt-secret --from-file=test/cert.pem --from-file=test/key.pem

# Deoploy
oc new-project oc-gate-operator-system
oc create -f deploy

#oc delete -f deploy
```

## Usage

Requesting a token for [oc-gate](https://github.com/yaacov/oc-gate) service is done using GateToken CRD,

Available fields are:

- match-path: string (required), match-path is a regular expresion used to validate API request path, API requests matching this pattern will be validated by the token. This field may not be empty.
- match-method: string, a comma separeted list of allowed http methods, defoult is "GET,OPTIONS"
- duration-sec: int, duration-sec is the duration in sec the token will be validated since it's invocation. Defalut value is 3600s (1h).
- from: string, from is time of token invocation, the token will not validate before this time, the token duration will start from this time. Defalut to token object creation time.

Creating a token requires a secret holding a RSA private-key for sighing the token in the namespace of the token (secret name: oc-gate-jwt-secret), nce token is ready it will be available in the GateToken status.

Get a token:

[![asciicast](https://asciinema.org/a/397136.svg)](https://asciinema.org/a/397136)

## Example GateToken CR

```yaml
apiVersion: ocgate.yaacov.com/v1beta1
kind: GateToken
metadata:
  name: gatetoken-sample
  namespace: oc-gate
spec:
  match-path: ^/k8s/apis/subresources.kubevirt.io/v1alpha3/namespaces/default/virtualmachineinstances/my-vm/vnc
```

## Example GateServer CR

```yaml
apiVersion: ocgate.yaacov.com/v1beta1
kind: GateServer
metadata:
  name: gateserver-sample
  namespace: oc-gate
spec:
  route: test-proxy.apps.ostest.test.metalkube.org
```

## Customize deploy

Requires

- requires GOLANG ver >= v1.15 dev env.
- user with admin permisions logged into the cluster.

```bash
# Deploy the operator, RBAC roles and CRDs
export USERNAME=yaacov
make deploy IMG=quay.io/$USERNAME/oc-gate-operator:v0.0.1
```

```bash
# Remove deployment of the operator, RBAC roles and CRDs
export USERNAME=yaacov
make undeploy IMG=quay.io/$USERNAME/oc-gate-operator:v0.0.1
```

