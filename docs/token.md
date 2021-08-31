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
# Set the vm namespace (must be same as proxy)
ns=gateway-example

# Get a bearer token of a k8s user who can create and read gatetoken resources in the example namespace:
token=$(oc whoami -t)
apipath=$(oc whoami --show-server)/apis/kubegateway.kubevirt.io/v1beta1/namespaces/$ns/gatetokens

# Generate the path for the demo:
date=$(date "+%y%m%d%H%M")
name=$vm-$date
# Get the secret, 
secret=$(oc get secrets -n $ns -o name | grep jwt-secret | cut -d "/" -f2)
# Generate the vnc subresource
path=/apis/subresources.kubevirt.io/v1/namespaces/$ns/virtualmachineinstances/$vm/vnc

# Create the gatetoken resource:
data="{\"apiVersion\":\"kubegateway.kubevirt.io/v1beta1\",\"kind\":\"GateToken\",\"metadata\":{\"name\":\"$name\",\"namespace\":\"$ns\"},\"spec\":{\"secret-name\":\"$secret\",\"urls\":[\"$path\"]}}"

curl -k -H 'Accept: application/json' -H "Authorization: Bearer $token" -H "Content-Type: application/json" --request POST --data $data $apipath
```

## Get the JWT singed token from the token resource

Once a token resource is registered it will try to sign the token, get the sign token:

```bash
curl -k -H 'Accept: application/json' -H "Authorization: Bearer $token" $apipath/$name | jq .status.token
```

## Use the token to access the resource

```bash
path=/apis/subresources.kubevirt.io/v1/namespaces/$ns/virtualmachineinstances/$vm/vnc
jwt=$(curl -k -H 'Accept: application/json' -H "Authorization: Bearer $token" $apipath/$name | jq .status.token)

# The proxy URL is set in the gateserver spec
proxyurl=https://$(oc get gateserver -o json | jq -r .items[0].spec.route)

# Open the link in a browser.
# The link is signed using ${jwt} and will access the k8s API at ${path}.
signed_link="${proxyurl}/auth/jwt/set?token=${jwt}&then=/noVNC/vnc_lite.html?path=k8s${path}"

google-chrome "${signed_link}"
```