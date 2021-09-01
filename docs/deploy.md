# Deploy

## Deploy using the example deployment file

The example deployment file use images from `quay.io/kubevirt-ui` and `gcr.io/kubebuilder`.
If your installation is connected to the internet and you do not intend to customize the images,
deployment using the example deployment file can be a good option.

```bash
oc create -f https://raw.githubusercontent.com/kubevirt-ui/kube-gateway-operator/main/deploy/kube-gateway-operator.yaml
```

## Deploy using customized/local images 

The gateway operator deployment uses three container images that require customization.
A user may use the default images or customized ones.

| Image | Description
|---|---
| gcr.io/kubebuilder/kube-rbac-proxy:v0.5.0 | RBAC proxy used by the operator manager
| quay.io/kubevirt-ui/kube-gateway:latest | The kube-gateway proxy server
| quay.io/kubevirt-ui/kube-gateway-operator:v0.0.1 | The kube-gateway operator image

To customize the deployment, replace these images in the example file:

```bash
curl https://raw.githubusercontent.com/kubevirt-ui/kube-gateway-operator/main/deploy/kube-gateway-operator.yaml > operator.yaml

# Check the current images for the rbac-proxy, kube-gateway, and kube-gateway-operator
# files and replace them with your customized/local images.

# Here is an example script for replacing the images:
RBAC_IMAGE=gcr.io/kubebuilder/kube-rbac-proxy:v0.5.0
GATEWAY_IMAGE=quay.io/kubevirt-ui/kube-gateway:latest
GATEWAY_OPERATOR_IMAGE=quay.io/kubevirt-ui/kube-gateway-operator:v0.0.1

RBAC_IMAGE_CI=ci.org/gateway/kube-rbac-proxy@sha256~1234
sed -i "s|${RBAC_IMAGE}|${RBAC_IMAGE_CI}|g;" operator.yaml

GATEWAY_IMAG_CI=ci.org/gateway/kube-gateway@sha256~1234
sed -i "s|${GATEWAY_IMAGE}|${GATEWAY_IMAGE_CI}|g;" operator.yaml

GATEWAY_OPERATOR_IMAGE_CI=ci.org/gateway/kube-gateway-operator@sha256~1234
sed -i "s|${GATEWAY_OPERATOR_IMAGE}|${GATEWAY_OPERATOR_IMAGE_CI}|g;" operator.yaml
```

```bash
# Once the operator.yaml file is ready, deploy the customized operator
oc create -f operator.yaml
```

## Starting a gateway

Now that the operator is installed, we can start running a kube-gateway server.
Create the proxy in the namespace that contains the k8s resources you wish to expose.

For this example, we will create a namespace called "gateway-example" and spin up a gateway server.

```bash
oc new-project gateway-example
```

Set the `namespace`, `route` and `image`.
`namespace` - the namespace containing the k8s resources to be exposed
`route` - the host of the proxy server
`image` - the `kube-gateway` container image

```bash
cat <<EOF | oc create -f -
apiVersion: kubegateway.kubevirt.io/v1beta1
kind: GateServer
metadata:
  name: gateserver-sample
  namespace: gateway-example
spec:
  route: 'kube-gateway-proxy.apps.ostest.test.metalkube.org'
  img: 'quay.io/kubevirt-ui/kube-gateway:latest'
EOF
```

The gateway manager pod should start running in the namespace.

### Important note

When creating signed tokens for this gateway proxy, a user must know the name of the secret holding the private key for signing the token.

```bash
oc get secrets -n gateway-example | grep jwt-secret
```



