# cloudflared-dns-controller
Kubernetes controller that automatically manages Cloudflare DNS records based on cloudflared tunnel ConfigMap.

This Kubernetes controller watches cloudflared ConfigMaps and automatically updates Cloudflare DNS records when the ingress field is modified.

For example, given the following ConfigMap, the controller creates a DNS record pointing api.example.com to the Cloudflare Tunnel specified in the ConfigMap. DNS records are managed via the Cloudflare API.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudflared
  namespace: cloudflared
data:
  config.yaml: |
    tunnel: xxxxx-xxxxxxxxxx

    ingress:
    - hostname: api.example.com
      service: http://traefik.traefik.svc.cluster.local:80
    - service: http_status:404
```

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

## Usage

First, create a ConfigMap to configure cloudflared.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudflared
  namespace: cloudflared
data:
  config.yaml: |
    tunnel: xxxxx-xxxxxxxxxx

    ingress:
    - hostname: api.example.com
      service: http://traefik.traefik.svc.cluster.local:80
    - service: http_status:404
```

Next, create an API token with Zone DNS Edit permission from the Cloudflare dashboard.

Then, deploy cloudflared-dns-controller by passing the token via a Kubernetes Secret or Helm values. The controller watches the ConfigMap, calculates the diff against existing DNS records, and automatically creates or deletes records accordingly.


See [values.yaml](charts/cloudflared-dns-controller/values.yaml) for the full list of configurable parameters.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
