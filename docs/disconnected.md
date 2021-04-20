
### Disconnected clusters

If the system is in a disconnected environment, without access to the public image repository, edit the yaml examples to use internally provided container images.

``` bash
# Edit the operator image in operator-controller-manager.yaml
curl https://raw.githubusercontent.com/yaacov/virt-gateway-operator/main/deploy/virt-gateway-operator.yaml \
    -o virt-gateway-operator.yaml

vim virt-gateway-operator.yaml
kubectl create -f virt-gateway-operator.yaml
```
