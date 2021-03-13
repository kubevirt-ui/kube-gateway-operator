# Build

```bash
make

IMG=quay.io/yaacov/oc-gate-operator make && make podman-build && make podman-push
```

## Build image

```bash
export IMG=quay.io/yaacov/oc-gate-operator:latest
make podman-build
```

## Push image

```bash
make podman-push
```

## Build deploy yaml files

```bash
# Get the image digests hase using podamn CLI:
# podman image inspect quay.io/yaacov/oc-gate-operator | jq .[0].RepoDigests[0]
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