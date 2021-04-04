# virt-gateway-operator

![alt gopher network](https://raw.githubusercontent.com/yaacov/kube-gateway/main/web/public/network-side.png)

Operate the [kube-gateway](https://github.com/yaacov/kube-gateway) service and access tokens on a cluster.

[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/virt-gateway-operator)](https://goreportcard.com/report/github.com/yaacov/virt-gateway-operator)
[![Go Reference](https://pkg.go.dev/badge/github.com/yaacov/virt-gateway-operator.svg)](https://pkg.go.dev/github.com/yaacov/virt-gateway-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

kube-gateway service allow none-k8s users access to single k8s resource for a limited time.
It uses signed, expiring JWTs to grant non k8s users access via a proxy server.

Once installed the operator manages two custom resources:

- [GateServer](#example-gateserver-cr): lounches the kube-gateway service that proxy k8s API calls to users outside the cluster.
- [GateToken](#example-gatetoken-cr): manages the creation of signed tokens used to authenticate with the kube-gateway service.

(gopher network image - [egonelbre/gophers](https://github.com/egonelbre/gophers))

## Deploy the operator

``` bash
# Deoploy the gate operator.
kubectl create -f \
    https://raw.githubusercontent.com/yaacov/virt-gateway-operator/main/deploy/virt-gateway-operator.yaml
```

### Deploy a gate server

``` bash
# Create a namespace to run the gate server.
kubectl create namespace kube-gateway

# Download and customize the kube-gateway-server example.
curl https://raw.githubusercontent.com/yaacov/virt-gateway-operator/main/deploy/virt-gateway-server.yaml \
    -o kube-gateway-server.yaml

vim kube-gateway-server.yaml
kubectl create -f kube-gateway-server.yaml
```

### Disconnected clusters

In disconnected enviorments without access to public image repository, edit the yaml examples to use internaly provided container images.

``` bash
# Edit the operator image in operator-controller-manager yaml file.
curl https://raw.githubusercontent.com/yaacov/virt-gateway-operator/main/deploy/virt-gateway-operator.yaml \
    -o virt-gateway-operator.yaml

vim virt-gateway-operator.yaml
kubectl create -f virt-gateway-operator.yaml
```

## Example GateToken CR

This example will generate a token that will give it's holder access to API calls matching the path "/k8s/apis/subresources.kubevirt.io/v1alpha3/namespaces/default/virtualmachineinstances/my-vm/vnc" for 1 hour. You can edit the route to match the route designated for the gate server on your cluster.

```yaml
apiVersion: ocgate.yaacov.com/v1beta1
kind: GateToken
metadata:
  name: gatetoken-sample
  namespace: kube-gateway
spec:
  namespace: "default"
  APIGroups:
    - "subresources.kubevirt.io"
  resources:
    - "virtualmachineinstances"
    - "virtualmachineinstances/vnc"
  resourceNames:
    - testvm
  generateServiceAccount: false
```

## Example GateServer CR

After the operator is set, users need to set up a gate server, this example will create an kube-gateway proxy server, wating for requests on URL "https://test-proxy.apps.ostest.test.metalkube.org". One gate server can handle requests for resources from different users and over different namespaces.


```yaml
apiVersion: ocgate.yaacov.com/v1beta1
kind: GateServer
metadata:
  name: gateserver-sample
  namespace: kube-gateway
spec:
  route: kube-gateway-proxy.apps-crc.testing
  # serviceAccount fields are used to create a service account for the oc gate proxy.
  # The proxy will run using this service account, it will be abale to
  # proxy only requests available to this service account. Make sure to allow the 
  # proxy to access all k8s resources that the web application will consume.
  serviceAccountVerbs:
    - "get"
  serviceAccountAPIGroups:
    - "subresources.kubevirt.io"
  serviceAccountResources:
    - "virtualmachineinstances"
  # gnerateSecret is used to automatically create a secret holding the asymetrical
  # keys needed to sign and authenticate the JWT tokens.
  gnerateSecret: true
  # passThrough is used to pass the request token directoy to k8s API server
  # without authenticating and replacing with the service account access token of the proxy.
  passThrough: false
  # the proxy server container image
  image: 'quay.io/yaacov/kube-gateway'
  # webAppImage is used to customize the static files of your web app.
  # this example will install the noVNC web application that consume
  # websockets streaming VNC data.
  webAppImage: 'quay.io/yaacov/kube-gateway-web-app-novnc'
```
