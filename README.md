# oc-gate-operator

creates tokens for the oc-gate service

## Deploy

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

## Usage

Requires a secret with private key on 'oc-gate' namespace:

```bash
# Use the oc-gate namespace
oc priject oc-gate

# create a secret
oc create -n oc-gate-operator-system secret generic oc-gate-jwt-secret --from-file=test/cert.pem --from-file=test/key.pem

# create a token request
oc create -f config/samples/ocgate_v1beta1_gatetoken.yaml

# check the token
oc get gatetoken gatetoken-sample -o yaml
```

## Build

```bash
# Compile operator
make

# Install CRD on cluser for running loaclly
make install
# make uninstall

# Run locally
make run
```

## Build images

```bash
export USERNAME=yaacov
make podman-build podman-push IMG=quay.io/$USERNAME/oc-gate-operator:v0.0.1
```
