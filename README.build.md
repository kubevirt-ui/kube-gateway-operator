# Build

```bash
make
```

## Build image

```bash
export IMG=quay.io/yaacov/oc-gate-operator:v0.0.1
make podman-build
```

## Push image

```bash
make podman-push
```

## Build deploy yaml files

```bash
export IMG=quay.io/yaacov/oc-gate-operator@sha256:98de8105eefb10263a52bd2730b3c5fee0b9a21960db34089f56a6dba8eec289
make deploy-dir
```

## Build bundle

```bash
make bundle
make bundle-build
```

## Push Bundle

```bash
podman push quay.io/yaacov/oc-gate-operator-bundle:v0.0.1
```