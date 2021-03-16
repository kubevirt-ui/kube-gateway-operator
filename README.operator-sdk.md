# Build

```bash
# Operator SDK tool flow:
perator-sdk init --domain yaacov.com --repo github.com/yaacov/oc-gate
operator-sdk create api --group oc-gate --version v1alpha1 --kind Token --resource --controller

... create the app

... compile

... build bundle and push it

operator-sdk run bundle quay.io/yaacov/oc-gate-operator-bundle:v0.0.1 --index-image quay.io/operator-framework/upstream-opm-builder:latest

... run tests

operator-sdk cleanup oc-gate-operator
```

## Build bundle

```bash
export IMG=quay.io/yaacov/oc-gate-operator@sha256:98de8105eefb10263a52bd2730b3c5fee0b9a21960db34089f56a6dba8eec289
make bundle
make bundle-build
```

## Push Bundle

```bash
podman push quay.io/yaacov/oc-gate-operator-bundle:v0.0.1
```
