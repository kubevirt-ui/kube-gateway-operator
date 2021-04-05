# Build

```bash
make && make podman-build && make podman-push

# Get the image digests hase using podamn CLI:
# podman image inspect quay.io/yaacov/virt-gateway-operator | jq .[0].RepoDigests[0]
IMG=quay.io/yaacov/virt-gateway-operator@sha256:98de8105eefb10263a52bd2730b3c5fee0b9a21960db34089f56a6dba8eec289 make deploy-dir
```

## operator-sdk

```basg
perator-sdk init --domain yaacov.com --repo github.com/yaacov/oc-gate
operator-sdk create api --group oc-gate --version v1alpha1 \
  --kind Token --resource --controller

... create the app

... compile

... build bundle and push it

operator-sdk run bundle quay.io/yaacov/oc-gate-operator-bundle:v0.0.1 \
  --index-image quay.io/operator-framework/upstream-opm-builder:latest

... run tests

operator-sdk cleanup oc-gate-operator
```
