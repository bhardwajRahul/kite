# 集群管理

集群管理接口位于 `/api/v1/admin/clusters/`。

这些接口要求调用方已经是管理员用户，或者使用一个拥有 `admin` 角色的 API 密钥。

## 获取集群列表

```http
GET /api/v1/admin/clusters/
```

示例：

```bash
curl \
  -H "Authorization: kite1-adminsecret" \
  https://kite.example.com/api/v1/admin/clusters/
```

响应示例：

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

## 创建集群

```http
POST /api/v1/admin/clusters/
Content-Type: application/json
```

请求体：

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

示例：

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

响应示例：

```json
{
  "id": 1,
  "message": "cluster created successfully"
}
```

## 更新集群

```http
PUT /api/v1/admin/clusters/:id
Content-Type: application/json
```

请求体：

```json
{
  "description": "production cluster",
  "prometheusURL": "http://prometheus.monitoring.svc:9090",
  "isDefault": false,
  "enabled": true
}
```

示例：

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

响应示例：

```json
{
  "message": "cluster updated successfully"
}
```

说明：

- 更新时 `config` 是可选的，不传就会保留现有 kubeconfig。
- 如果传了新的非空 `name`，集群名称会被更新。

## 删除集群

```http
DELETE /api/v1/admin/clusters/:id
```

示例：

```bash
curl \
  -X DELETE \
  -H "Authorization: kite1-adminsecret" \
  https://kite.example.com/api/v1/admin/clusters/1
```

响应示例：

```json
{
  "message": "cluster deleted successfully"
}
```

说明：默认集群不能被删除。

## 从 kubeconfig 导入集群

这个初始化接口位于：

```http
POST /api/v1/admin/clusters/import
Content-Type: application/json
```

它主要用于首次初始化，并且只会在系统里还没有任何集群时生效。

请求体：

```json
{
  "config": "<kubeconfig-yaml>",
  "inCluster": false
}
```

使用 in-cluster 导入时：

```json
{
  "inCluster": true
}
```
