[![Go Report Card](https://goreportcard.com/badge/github.com/kubevirt-ui/kube-gateway-operator)](https://goreportcard.com/report/github.com/kubevirt-ui/kube-gateway-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# kube-gateway-operator

![alt gopher network](https://raw.githubusercontent.com/kubevirt-ui/kube-gateway-operator/main/docs/network-side.png)

The kube-gateway-operator installs and operates the [kube-gateway](https://github.com/kubevirt-ui/kube-gateway) service, which allows access to the k8s API using time-limited access tokens and usage of one-time access tokens to access k8s resources.

The operator manages service accounts, permissions, and secrets needed for the operation of the [kube-gateway](https://github.com/kubevirt-ui/kube-gateway) service and JWT token generation used for one-time k8s API access.

## Build and push images

```bash
export USERNAME=yaacov
IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1 make podman-build
IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1 make podman-push
```

## Deploy

For more information about deployment options see the [deploy](/docs/deploy.md) doc.

```bash
# Deploy the operator, RBAC roles, and CRDs
export USERNAME=yaacov
IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1 make deploy

# Deploy from an example deployment yaml. Will use pre-defined images and permissions. 
# Users can also copy this file to a local directory and edit the container image used.
oc create -f https://raw.githubusercontent.com/kubevirt-ui/kube-gateway-operator/main/deploy/kube-gateway-operator.yaml
```

```bash
# Remove deployment of the operator, RBAC roles, and CRDs
export USERNAME=yaacov
IMG=quay.io/$USERNAME/kube-gateway-operator:v0.0.1 make undeploy
```

![alt install operator](https://raw.githubusercontent.com/kubevirt-ui/kube-gateway-operator/main/docs/install-operator.gif)

## Create GateServer and GateToken examples

For more information about running the gateway proxy and generating a token see the [token](/docs/token.md) and [deploy](/docs/deploy.md#starting-a-gateway) docs.

```bash
# Use the kube-gateway namespace
oc create namespace kube-gateway

# Create a sample gateway server
oc create -f config/samples/kubegateway_v1beta1_gateserver.yaml

# Create a sample token request
oc create -f config/samples/kubegateway_v1beta1_gatetoken.yaml

# Check the token
oc get gatetoken gatetoken-sample -o yaml
```
Example files:
[gateserver.yaml](/config/samples/kubegateway_v1beta1_gateserver.yaml),
[gatetoken.yaml](/config/samples/kubegateway_v1beta1_gatetoken.yaml)

![alt create signed link](https://raw.githubusercontent.com/kubevirt-ui/kube-gateway-operator/main/docs/create-signed-link.gif)

## Building for local development

```bash
# Compile the operator
make

# Install CRDs on the cluster for running locally
make install
# make uninstall

# Run locally
make run
```

(gopher network image - [egonelbre/gophers](https://github.com/egonelbre/gophers))
