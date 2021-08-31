# Deploy

## Deploy using the example deployment

The example deployment, use images from `quay.io/kubevirt-ui` and `gcr.io/kubebuilder`
If your installation is connected to the internet and you do not intend to customize the images,
Deployment using the example deployment file can be a good option.

```bash
oc create -f https://raw.githubusercontent.com/kubevirt-ui/kube-gateway-operator/main/deploy/kube-gateway-operator.yaml
```

## Deploy using customized / local images 

The example deployment use 3 container images that need customization

gcr.io/kubebuilder/kube-rbac-proxy:v0.5.0 - rbac proxy used by the operator manager
quay.io/kubevirt-ui/kube-gateway:latest - the kube gateway proxy server
quay.io/kubevirt-ui/kube-gateway-operator:v0.0.1 - the kube gateway operator image

To cusomize the deployment replace this images in the example file

```bash
curl https://raw.githubusercontent.com/kubevirt-ui/kube-gateway-operator/main/deploy/kube-gateway-operator.yaml > operator.yaml

# Check the current images for rbac-proxy, kube-gateway and kube-gateway-operator
# and replacy them with you customize / local images:
RBAC_IMAG=gcr.io/kubebuilder/kube-rbac-proxy:v0.5.0
GATEWAY_IMAG=quay.io/kubevirt-ui/kube-gateway:latest
GATEWAY_OPERATOR_IMAG=quay.io/kubevirt-ui/kube-gateway-operator:v0.0.1

RBAC_IMAG_CI=ci.org/gateway/kube-rbac-proxy@sha256~1234
sed -i "s|${RBAC_IMAG}|${RBAC_IMAG_CI}|g;" operator.yaml

GATEWAY_IMAG_CI=ci.org/gateway/kube-gateway@sha256~1234
sed -i "s|${GATEWAY_IMAG}|${GATEWAY_IMAG_CI}|g;" operator.yaml

GATEWAY_OPERATOR_IMAG_CI=ci.org/gateway/kube-gateway-operator@sha256~1234
sed -i "s|${GATEWAY_OPERATOR_IMAG}|${GATEWAY_OPERATOR_IMAG_CI}|g;" operator.yaml
```

```bash
# Deploy the customized operator
oc create -f operator.yaml
```

## Starting a gateway

Once operator it installed, we can start running a gateway.

Create the proxy in the namespace that contain the k8s resources you with to expose.

```bash
oc new-project gateway-example
```

Set the `namespace`, `route` and `image`
`namespace` - the namespace to expose k8s resources
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





