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
