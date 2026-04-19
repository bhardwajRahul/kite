# 用户管理

用户管理接口位于 `/api/v1/admin/users/`。

除特别说明外，这些接口要求调用方已经是管理员用户，或者使用一个拥有 `admin` 角色的 API 密钥。

## 初始化第一个超级管理员

这个接口用于首次初始化，在还没有建立管理员认证之前即可调用。

```http
POST /api/v1/admin/users/create_super_user
Content-Type: application/json
```

请求体：

```json
{
  "username": "admin",
  "password": "change-me",
  "name": "Administrator"
}
```

它只会在系统里还没有任何用户时成功。

## 获取用户列表

```http
GET /api/v1/admin/users/
```

支持的查询参数：

- `page`
- `size`
- `search`
- `role`
- `sortBy`
- `sortOrder`

示例：

```bash
curl \
  -H "Authorization: kite1-adminsecret" \
  "https://kite.example.com/api/v1/admin/users/?page=1&size=20&search=alice&sortOrder=desc"
```

响应示例：

```json
{
  "users": [
    {
      "id": 12,
      "username": "alice",
      "provider": "password",
      "name": "Alice",
      "enabled": true,
      "lastLoginAt": "2026-04-20T08:30:00Z",
      "roles": [
        {
          "id": 2,
          "name": "viewer"
        }
      ]
    }
  ],
  "total": 1,
  "page": 1,
  "size": 20
}
```

## 创建密码用户

```http
POST /api/v1/admin/users/
Content-Type: application/json
```

请求体：

```json
{
  "username": "alice",
  "password": "change-me",
  "name": "Alice"
}
```

示例：

```bash
curl \
  -X POST \
  -H "Authorization: kite1-adminsecret" \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"change-me","name":"Alice"}' \
  https://kite.example.com/api/v1/admin/users/
```

## 更新用户

```http
PUT /api/v1/admin/users/:id
Content-Type: application/json
```

请求体：

```json
{
  "name": "Alice Zhang",
  "avatar_url": "https://example.com/avatar.png"
}
```

示例：

```bash
curl \
  -X PUT \
  -H "Authorization: kite1-adminsecret" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Zhang"}' \
  https://kite.example.com/api/v1/admin/users/12
```

## 删除用户

```http
DELETE /api/v1/admin/users/:id
```

示例：

```bash
curl \
  -X DELETE \
  -H "Authorization: kite1-adminsecret" \
  https://kite.example.com/api/v1/admin/users/12
```

响应示例：

```json
{
  "success": true
}
```

## 重置用户密码

```http
POST /api/v1/admin/users/:id/reset_password
Content-Type: application/json
```

请求体：

```json
{
  "password": "new-password"
}
```

## 启用或禁用用户

```http
POST /api/v1/admin/users/:id/enable
Content-Type: application/json
```

请求体：

```json
{
  "enabled": false
}
```

响应示例：

```json
{
  "success": true
}
```
