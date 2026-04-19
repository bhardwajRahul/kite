# 资源操作

Kite 在 `/api/v1/<resource>` 下为内置 Kubernetes 资源提供通用的 CRUD 风格接口。

## 调用前提

资源接口需要：

- 一个已经认证的用户或 API 密钥
- 一个目标集群，通常通过 `x-cluster-name` 传入

示例：

```bash
-H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
-H "x-cluster-name: demo-cluster"
```

## 路径规则

内置资源会使用下面两种路径形式：

- namespace 级资源：`/api/v1/<resource>/<namespace>`
- 集群级资源：`/api/v1/<resource>/_all`

示例：

- ConfigMap：`/api/v1/configmaps/default`
- Deployment：`/api/v1/deployments/default`
- Namespace：`/api/v1/namespaces/_all`
- Node：`/api/v1/nodes/_all`

## 支持的操作

对于内置资源，Kite 暴露了这些通用路由：

### namespace 级资源

```text
GET    /api/v1/<resource>/<namespace>
GET    /api/v1/<resource>/<namespace>/<name>
POST   /api/v1/<resource>/<namespace>
PUT    /api/v1/<resource>/<namespace>/<name>
PATCH  /api/v1/<resource>/<namespace>/<name>
DELETE /api/v1/<resource>/<namespace>/<name>
```

### 集群级资源

```text
GET    /api/v1/<resource>/_all
GET    /api/v1/<resource>/_all/<name>
POST   /api/v1/<resource>/_all
PUT    /api/v1/<resource>/_all/<name>
PATCH  /api/v1/<resource>/_all/<name>
DELETE /api/v1/<resource>/_all/<name>
```

## Create 示例

下面的例子会在 `default` namespace 创建一个 ConfigMap。

```bash
curl \
  -X POST \
  -H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "x-cluster-name: demo-cluster" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "name": "example-config",
      "namespace": "default"
    },
    "data": {
      "APP_MODE": "prod",
      "LOG_LEVEL": "info"
    }
  }' \
  https://kite.example.com/api/v1/configmaps/default
```

## Update 示例

`PUT` 是完整更新，不是局部更新。请求体里需要带上当前对象的 `metadata.resourceVersion`，否则 Kubernetes 会拒绝这次更新。

下面的例子会整体更新这个 ConfigMap。

```bash
curl \
  -X PUT \
  -H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "x-cluster-name: demo-cluster" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "name": "example-config",
      "namespace": "default",
      "resourceVersion": "123456"
    },
    "data": {
      "APP_MODE": "staging",
      "LOG_LEVEL": "debug"
    }
  }' \
  https://kite.example.com/api/v1/configmaps/default/example-config
```

## Patch 示例

`PATCH` 接收原始 patch 内容。当前支持的 patch 类型有：

- 默认：strategic merge patch
- `?patchType=merge`：JSON merge patch
- `?patchType=json`：JSON patch

下面的例子使用 JSON merge patch，只更新一个字段，不发送整个对象。

```bash
curl \
  -X PATCH \
  -H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "x-cluster-name: demo-cluster" \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "LOG_LEVEL": "warn"
    }
  }' \
  "https://kite.example.com/api/v1/configmaps/default/example-config?patchType=merge"
```

如果要通过 `PATCH` 重启一个 Deployment，可以更新 `spec.template.metadata.annotations` 下的一个时间戳字段。这样会修改 Pod template，从而触发一次新的 rollout。

```bash
curl \
  -X PATCH \
  -H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "x-cluster-name: demo-cluster" \
  -H "Content-Type: application/json" \
  -d '{
    "spec": {
      "template": {
        "metadata": {
          "annotations": {
            "kite.kubernetes.io/restartedAt": "2026-04-20T12:00:00Z"
          }
        }
      }
    }
  }' \
  "https://kite.example.com/api/v1/deployments/default/example-app?patchType=merge"
```

## Delete 示例

下面的例子会删除这个 ConfigMap。

```bash
curl \
  -X DELETE \
  -H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "x-cluster-name: demo-cluster" \
  https://kite.example.com/api/v1/configmaps/default/example-config
```

可选删除参数：

- `force=true`：使用 0 秒 grace period 强制删除
- `wait=false`：立即返回，不等待删除完成
- `cascade=false`：孤儿化依赖对象，而不是前台级联删除

示例：

```bash
curl \
  -X DELETE \
  -H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "x-cluster-name: demo-cluster" \
  "https://kite.example.com/api/v1/deployments/default/example-app?force=true&wait=false"
```

## 集群级资源示例

集群级资源路径里使用 `/_all`。

下面的例子会给一个 Namespace 打标签：

```bash
curl \
  -X PATCH \
  -H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "x-cluster-name: demo-cluster" \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "labels": {
        "team": "platform"
      }
    }
  }' \
  https://kite.example.com/api/v1/namespaces/_all/demo
```

## 自定义资源

自定义资源使用 `/:crd/...` 形式的路由，其中 `:crd` 是 CRD 名称，例如 `rollouts.argoproj.io`。

示例：

- namespace 级自定义资源查询：`/api/v1/rollouts.argoproj.io/default/example`
- 集群级自定义资源查询：`/api/v1/<crd>/_all/example`

按当前服务端已注册的路由来看：

- 自定义资源支持 list 和 get
- 自定义资源支持 update 和 delete
- 这里没有把自定义资源的 create 和 patch 作为当前稳定文档面来写，当前通用的 create / update / patch / delete 示例以已注册的内置资源路由为准
