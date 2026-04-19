# User Management

User management APIs are under `/api/v1/admin/users/`.

These endpoints require an authenticated admin user, or an API key with the `admin` role, unless noted otherwise.

## Bootstrap the first super user

This route is intended for first-time setup and is available before admin authentication is established.

```http
POST /api/v1/admin/users/create_super_user
Content-Type: application/json
```

Request body:

```json
{
  "username": "admin",
  "password": "change-me",
  "name": "Administrator"
}
```

It only succeeds when the system has no users yet.

## List users

```http
GET /api/v1/admin/users/
```

Supported query parameters:

- `page`
- `size`
- `search`
- `role`
- `sortBy`
- `sortOrder`

Example:

```bash
curl \
  -H "Authorization: kite1-adminsecret" \
  "https://kite.example.com/api/v1/admin/users/?page=1&size=20&search=alice&sortOrder=desc"
```

Response example:

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

## Create a password user

```http
POST /api/v1/admin/users/
Content-Type: application/json
```

Request body:

```json
{
  "username": "alice",
  "password": "change-me",
  "name": "Alice"
}
```

Example:

```bash
curl \
  -X POST \
  -H "Authorization: kite1-adminsecret" \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"change-me","name":"Alice"}' \
  https://kite.example.com/api/v1/admin/users/
```

## Update a user

```http
PUT /api/v1/admin/users/:id
Content-Type: application/json
```

Request body:

```json
{
  "name": "Alice Zhang",
  "avatar_url": "https://example.com/avatar.png"
}
```

Example:

```bash
curl \
  -X PUT \
  -H "Authorization: kite1-adminsecret" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Zhang"}' \
  https://kite.example.com/api/v1/admin/users/12
```

## Delete a user

```http
DELETE /api/v1/admin/users/:id
```

Example:

```bash
curl \
  -X DELETE \
  -H "Authorization: kite1-adminsecret" \
  https://kite.example.com/api/v1/admin/users/12
```

Response example:

```json
{
  "success": true
}
```

## Reset a user's password

```http
POST /api/v1/admin/users/:id/reset_password
Content-Type: application/json
```

Request body:

```json
{
  "password": "new-password"
}
```

## Enable or disable a user

```http
POST /api/v1/admin/users/:id/enable
Content-Type: application/json
```

Request body:

```json
{
  "enabled": false
}
```

Response example:

```json
{
  "success": true
}
```
