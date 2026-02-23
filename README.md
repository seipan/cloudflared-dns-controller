# cloudflared-dns-controller
Kubernetes controller that automatically manages Cloudflare DNS records based on cloudflared tunnel ConfigMap.

## Installation

### Helm

```bash
helm repo add cloudflared-dns-controller https://seipan.github.io/cloudflared-dns-controller
helm repo update
```

With inline credentials:

```bash
helm install cloudflared-dns-controller cloudflared-dns-controller/cloudflared-dns-controller \
  --set cloudflare.apiToken=<your-api-token> \
  --set cloudflare.zoneID=<your-zone-id>
```

With an existing Secret:

```bash
kubectl create secret generic cloudflare-credentials \
  --from-literal=api-token=<your-api-token> \
  --from-literal=zone-id=<your-zone-id>

helm install cloudflared-dns-controller cloudflared-dns-controller/cloudflared-dns-controller \
  --set cloudflare.existingSecret=cloudflare-credentials
```

### Kustomize

```bash
# Create the Cloudflare credentials Secret
kubectl create ns cloudflared-dns-controller-system
kubectl create secret generic cloudflare-credentials \
  --from-literal=api-token=<your-api-token> \
  --from-literal=zone-id=<your-zone-id> \
  -n cloudflared-dns-controller-system

# Deploy the controller
make deploy IMG=ghcr.io/seipan/cloudflared-dns-controller:<tag>
```

## Usage

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudflared
  namespace: cloudflared
data:
  config.yaml: |
    tunnel: xxxxx-xxxxxxxxxx
    credentials-file: /etc/cloudflared/creds/credentials.json
    metrics: 0.0.0.0:2000

    no-autoupdate: true
    ingress:
    # The first rule proxies traffic to the httpbin sample Service defined in app.yaml

    - hostname: api.example.com
      service: http://traefik.traefik.svc.cluster.local:80
    - service: http_status:404
```
