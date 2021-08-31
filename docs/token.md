# Token

Generating a signed token requires access to the secret holding the private key.

## Getting the name of the secret

When spinning up a gateway server it creates a secret containing the private
and public keys used to sign and authenticate the JWT tokens.

The secret will end with `jwt-secret`:

```bash
oc get secrets -n <namespace running the gateway proxy> | grep jwt-secret
```

## Generating a virtual machine for this demo

```bash
cat <<EOF | oc create -f -
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: testvm
spec:
  running: true
  template:
    spec:
      domain:
        devices:
          disks:
          - disk:
              bus: virtio
            name: rootfs
        resources:
          requests:
            memory: 64M
      volumes:
        - name: rootfs
          containerDisk:
            image: quay.io/kubevirt/cirros-registry-disk-demo
EOF
```

## Generating a token

The folowing example describe how to create a token resource using a curl commend to access kubevirt vnc server.

```bash
# Set the vm name
vm=testvm
# Set the vm namespace (the virtual machine must be in the same namespace as the proxy)
ns=gateway-example

# Get the admin user k8s bearer token, and the k8s API path,
# We will use the k8s API and credentials to create the gatetoken resource:
token=$(oc whoami -t)
apipath=$(oc whoami --show-server)/apis/kubegateway.kubevirt.io/v1beta1/namespaces/$ns/gatetokens

# Generate a uniqe gatetoken name
date=$(date "+%y%m%d%H%M")
name=$vm-$date
# Get the name of the secret holding the private key for signing the gatetoken
secret=$(oc get secrets -n $ns -o name | grep jwt-secret | cut -d "/" -f2)

# Generate the vnc subresource path
path=/apis/subresources.kubevirt.io/v1/namespaces/$ns/virtualmachineinstances/$vm/vnc

# Create the gatetoken resource
data="{\"apiVersion\":\"kubegateway.kubevirt.io/v1beta1\",\"kind\":\"GateToken\",\"metadata\":{\"name\":\"$name\",\"namespace\":\"$ns\"},\"spec\":{\"secret-name\":\"$secret\",\"urls\":[\"$path\"]}}"

# Call k8s API using admin credentials to create a new gatetoken
curl -k -H 'Accept: application/json' -H "Authorization: Bearer $token" -H "Content-Type: application/json" --request POST --data $data $apipath

# Another way to create the gatetoken is using the oc command
# cat <<EOF | oc create -f -
# apiVersion: kubegateway.kubevirt.io/v1beta1
# kind: GateToken\
# metadata:
#   name: $name
#   namespace: $ns
# spec:
#   secret-name: $secret
#   urls:
#   - $path
# EOF
```

## Get the JWT singed token from the token resource

Once a token resource is registered it will try to sign the token, get the sign token:

```bash
# Get the token resource
curl -k -H 'Accept: application/json' -H "Authorization: Bearer $token" $apipath/$name
```

## Use the token to access the resource

```bash
# Get the JWT from the gatetoken resource using admin credentials
jwt=$(curl -k -H 'Accept: application/json' -H "Authorization: Bearer $token" $apipath/$name | jq .status.token)

# Another way to get the gatetoken is using the oc command
# oc get gatetoken $name -o json | jq .status.token

# The proxy URL is set in the gateserver spec
proxyurl=https://$(oc get gateserver -o json | jq -r .items[0].spec.route)

# The link is signed using ${jwt} and will access the k8s API at ${path}.
signed_link="${proxyurl}/auth/jwt/set?token=${jwt}&then=/noVNC/vnc_lite.html?path=k8s${path}"

# Users holding the signed link will be able to use it for 1h:

# Open the link in a browser
google-chrome "${signed_link}"
```
