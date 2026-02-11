# cloudflare-dns-controller
kubernetes controller for cloudflare dns


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
