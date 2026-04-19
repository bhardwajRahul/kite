# Cluster Management

Cluster management APIs are under `/api/v1/admin/clusters/`.

These endpoints require an authenticated admin user, or an API key with the `admin` role.

## List clusters

```http
GET /api/v1/admin/clusters/
```

Example:

```bash
curl \
  -H "Authorization: kite1-adminsecret" \
  https://kite.example.com/api/v1/admin/clusters/
```

Response example:

```json
[
  {
    "id": 1,
    "name": "demo-cluster",
    "description": "staging cluster",
    "enabled": true,
    "inCluster": false,
    "isDefault": true,
    "prometheusURL": "http://prometheus.monitoring.svc:9090",
    "config": "",
    "version": "v1.31.0"
  }
]
```

## Create a cluster

```http
POST /api/v1/admin/clusters/
Content-Type: application/json
```

Request body:

```json
{
  "name": "demo-cluster",
  "description": "staging cluster",
  "config": "<kubeconfig-yaml>",
  "prometheusURL": "http://prometheus.monitoring.svc:9090",
  "inCluster": false,
  "isDefault": true
}
```

Example:

```bash
curl \
  -X POST \
  -H "Authorization: kite1-adminsecret" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"demo-cluster",
    "description":"staging cluster",
    "config":"apiVersion: v1\nkind: Config\n...",
    "prometheusURL":"http://prometheus.monitoring.svc:9090",
    "inCluster":false,
    "isDefault":true
  }' \
  https://kite.example.com/api/v1/admin/clusters/
```

Response example:

```json
{
  "id": 1,
  "message": "cluster created successfully"
}
```

## Update a cluster

```http
PUT /api/v1/admin/clusters/:id
Content-Type: application/json
```

Request body:

```json
{
  "description": "production cluster",
  "prometheusURL": "http://prometheus.monitoring.svc:9090",
  "isDefault": false,
  "enabled": true
}
```

Example:

```bash
curl \
  -X PUT \
  -H "Authorization: kite1-adminsecret" \
  -H "Content-Type: application/json" \
  -d '{
    "description":"production cluster",
    "enabled":true
  }' \
  https://kite.example.com/api/v1/admin/clusters/1
```

Response example:

```json
{
  "message": "cluster updated successfully"
}
```

Notes:

- `config` is optional on update. If omitted, the existing kubeconfig is kept.
- `name` can be changed by passing a new non-empty value.

## Delete a cluster

```http
DELETE /api/v1/admin/clusters/:id
```

Example:

```bash
curl \
  -X DELETE \
  -H "Authorization: kite1-adminsecret" \
  https://kite.example.com/api/v1/admin/clusters/1
```

Response example:

```json
{
  "message": "cluster deleted successfully"
}
```

Note: the default cluster cannot be deleted.

## Import clusters from kubeconfig

This bootstrap endpoint is available at:

```http
POST /api/v1/admin/clusters/import
Content-Type: application/json
```

It is intended for the initial setup flow and only works when no clusters exist yet.

Request body:

```json
{
  "config": "<kubeconfig-yaml>",
  "inCluster": false
}
```

In-cluster import example:

```json
{
  "inCluster": true
}
```
