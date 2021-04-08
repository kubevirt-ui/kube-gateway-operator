# Test automation resources

## Setup Minikube with ingress and kubevirt

```bash
minikube start --driver=podman --addons=kubevirt,ingress

# wait for minikube to finish installation
kubectl wait --timeout=180s --for=condition=Available -n kubevirt deployments virt-controller
kubectl wait --timeout=180s --for=condition=Available -n kubevirt kv/kubevirt
```

## Deploy operator

```bash
IMG=quay.io/yaacov/virt-gateway-operator make deploy

kubectl wait --timeout=180s --for=condition=Available -n kube-gateway deployments/kube-gateway-server

# check service and ingress
kubectl get pods,ingress -n kube-gateway

```

## Start two virtual machines

```bash
# create the test namespace
kubectl create -f test_namespace.yaml

kubectl create -f test_vm_01.yaml
kubectl create -f test_vm_02.yaml

kubectl wait --timeout=180s --for=condition=Ready -n kube-gateway-test vmi/testvm01
kubectl wait --timeout=180s --for=condition=Ready -n kube-gateway-test vmi/testvm02
```

## Create two tokens to access the noVNC web application

```bash
# setup some helper variables
export vmnamespace=kube-gateway-test
export vmname1=testvm01
export vmname2=testvm02

# create tokens
kubectl create -f test_token_vm_01.yaml
kubectl create -f test_token_vm_02.yaml

kubectl wait --timeout=180s --for=condition=Ready -n kube-gateway gatetoken/token-testvm01
kubectl wait --timeout=180s --for=condition=Ready -n kube-gateway gatetoken/token-testvm02

export jwt1=$(kubectl get gatetoken/token-testvm01 -n kube-gateway -o jsonpath="{.status.token}"); echo $jwt1
export jwt2=$(kubectl get gatetoken/token-testvm02 -n kube-gateway -o jsonpath="{.status.token}"); echo $jwt2 

# test token1 with vm1
google-chrome "https://kube-gateway-proxy.apps-crc.testing/auth/token?token=${jwt1}&then=/noVNC/vnc_lite.html?path=k8s/apis/subresources.kubevirt.io/v1alpha3/namespaces/${vmnamespace}/virtualmachineinstances/${vmname1}/vnc"

# test token1 with vm2
google-chrome "https://kube-gateway-proxy.apps-crc.testing/auth/token?token=${jwt1}&then=/noVNC/vnc_lite.html?path=k8s/apis/subresources.kubevirt.io/v1alpha3/namespaces/${vmnamespace}/virtualmachineinstances/${vmname2}/vnc"

# test token2 with vm2
google-chrome "https://kube-gateway-proxy.apps-crc.testing/auth/token?token=${jwt2}&then=/noVNC/vnc_lite.html?path=k8s/apis/subresources.kubevirt.io/v1alpha3/namespaces/${vmnamespace}/virtualmachineinstances/${vmname2}/vnc"

# wait for token to expire
kubectl wait --timeout=180s --for=condition=Completed -n kube-gateway gatetoken/token-testvm01

# test token1 with vm1
google-chrome "https://kube-gateway-proxy.apps-crc.testing/auth/token?token=${jwt1}&then=/noVNC/vnc_lite.html?path=k8s/apis/subresources.kubevirt.io/v1alpha3/namespaces/${vmnamespace}/virtualmachineinstances/${vmname1}/vnc"
```

## Cleanup

``` bash
# delete VMs test namespace
kubectl delete -f test_vm_01.yaml
kubectl delete -f test_vm_02.yaml
kubectl delete -f test_namespace.yaml

# delete tokens
kubectl delete -f test_token_vm_01.yaml
kubectl delete -f test_token_vm_02.yaml
```
