# RBAC 管理

RBAC 管理接口位于 `/api/v1/admin/roles/`。

这些接口要求调用方已经是管理员用户，或者使用一个拥有 `admin` 角色的 API 密钥。

## 获取角色列表

```http
GET /api/v1/admin/roles/
```

示例：

```bash
curl \
  -H "Authorization: kite1-adminsecret" \
  https://kite.example.com/api/v1/admin/roles/
```

响应示例：

```json
{
  "roles": [
    {
      "id": 2,
      "name": "viewer",
      "description": "Viewer role with read-only access",
      "isSystem": true,
      "clusters": ["*"],
      "namespaces": ["*"],
      "resources": ["*"],
      "verbs": ["get", "log"],
      "assignments": []
    }
  ]
}
```

## 获取单个角色

```http
GET /api/v1/admin/roles/:id
```

## 创建角色

```http
POST /api/v1/admin/roles/
Content-Type: application/json
```

请求体：

```json
{
  "name": "namespace-editor",
  "description": "Edit workloads in dev",
  "clusters": ["demo-cluster"],
  "namespaces": ["dev"],
  "resources": ["deployments", "pods", "services"],
  "verbs": ["get", "create", "update", "delete", "log"]
}
```

示例：

```bash
curl \
  -X POST \
  -H "Authorization: kite1-adminsecret" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"namespace-editor",
    "description":"Edit workloads in dev",
    "clusters":["demo-cluster"],
    "namespaces":["dev"],
    "resources":["deployments","pods","services"],
    "verbs":["get","create","update","delete","log"]
  }' \
  https://kite.example.com/api/v1/admin/roles/
```

## 更新角色

```http
PUT /api/v1/admin/roles/:id
Content-Type: application/json
```

请求体字段与创建角色时一致。

## 删除角色

```http
DELETE /api/v1/admin/roles/:id
```

示例：

```bash
curl \
  -X DELETE \
  -H "Authorization: kite1-adminsecret" \
  https://kite.example.com/api/v1/admin/roles/3
```

响应示例：

```json
{
  "success": true
}
```

## 分配角色

```http
POST /api/v1/admin/roles/:id/assign
Content-Type: application/json
```

请求体：

```json
{
  "subjectType": "user",
  "subject": "alice"
}
```

当前服务端实现支持的 `subjectType`：

- `user`
- `group`

示例：

```bash
curl \
  -X POST \
  -H "Authorization: kite1-adminsecret" \
  -H "Content-Type: application/json" \
  -d '{"subjectType":"group","subject":"platform-team"}' \
  https://kite.example.com/api/v1/admin/roles/3/assign
```

## 移除角色分配

```http
DELETE /api/v1/admin/roles/:id/assign?subjectType=user&subject=alice
```

示例：

```bash
curl \
  -X DELETE \
  -H "Authorization: kite1-adminsecret" \
  "https://kite.example.com/api/v1/admin/roles/3/assign?subjectType=group&subject=platform-team"
```
