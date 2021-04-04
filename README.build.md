# Build

```bash
make && make podman-build && make podman-push

# Get the image digests hase using podamn CLI:
# podman image inspect quay.io/yaacov/virt-gateway-operator | jq .[0].RepoDigests[0]
IMG=quay.io/yaacov/virt-gateway-operator@sha256:98de8105eefb10263a52bd2730b3c5fee0b9a21960db34089f56a6dba8eec289 make deploy-dir
```
