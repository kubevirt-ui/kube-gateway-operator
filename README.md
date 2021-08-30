[![Go Report Card](https://goreportcard.com/badge/github.com/kubevirt-ui/kube-gateway-operator)](https://goreportcard.com/report/github.com/kubevirt-ui/kube-gateway-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# kube-gateway-operator

![alt gopher network](https://raw.githubusercontent.com/kubevirt-ui/kube-gateway/main/docs/network-side.png)

kube-gateway-operator installs and operate [kube-gateway](https://github.com/kubevirt-ui/kube-gateway)

## Build and push images

```bash
export USERNAME=yaacov
IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1 make podman-build
IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1 make podman-push
```

## Deploy

```bash
# Deploy the operator, RBAC roles and CRDs
export USERNAME=yaacov
IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1 make deploy
```

```bash
# Remove deployment of the operator, RBAC roles and CRDs
export USERNAME=yaacov
IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1 make undeploy
```

## Create GateToken CR

Requires a secret with private key on 'kube-gateway' namespace:

```bash
# Use the kube-gateway namespace
oc create namespace kube-gateway

# create a sample gateway server
oc create -f config/samples/kubegateway_v1beta1_gateserver.yaml

# create a sample token request
oc create -f config/samples/kubegateway_v1beta1_gatetoken.yaml

# check the token
oc get gatetoken gatetoken-sample -o yaml
```

## Local dev build

```bash
# Compile operator
make

# Install CRD on cluser for running loaclly
make install
# make uninstall

# Run locally
make run
```

(gopher network image - [egonelbre/gophers](https://github.com/egonelbre/gophers))
