# Resources

Kite exposes generic CRUD-style APIs for built-in Kubernetes resources under `/api/v1/<resource>`.

## Prerequisites

Resource endpoints require:

- an authenticated user or API key
- a target cluster, usually passed through `x-cluster-name`

Example:

```bash
-H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
-H "x-cluster-name: demo-cluster"
```

## Path patterns

Built-in resources use one of these two patterns:

- namespaced resources: `/api/v1/<resource>/<namespace>`
- cluster-scoped resources: `/api/v1/<resource>/_all`

Examples:

- ConfigMap: `/api/v1/configmaps/default`
- Deployment: `/api/v1/deployments/default`
- Namespace: `/api/v1/namespaces/_all`
- Node: `/api/v1/nodes/_all`

## Supported operations

For built-in resources, Kite exposes these generic routes:

### Namespaced resources

```text
GET    /api/v1/<resource>/<namespace>
GET    /api/v1/<resource>/<namespace>/<name>
POST   /api/v1/<resource>/<namespace>
PUT    /api/v1/<resource>/<namespace>/<name>
PATCH  /api/v1/<resource>/<namespace>/<name>
DELETE /api/v1/<resource>/<namespace>/<name>
```

### Cluster-scoped resources

```text
GET    /api/v1/<resource>/_all
GET    /api/v1/<resource>/_all/<name>
POST   /api/v1/<resource>/_all
PUT    /api/v1/<resource>/_all/<name>
PATCH  /api/v1/<resource>/_all/<name>
DELETE /api/v1/<resource>/_all/<name>
```

## Create example

This example creates a ConfigMap in the `default` namespace.

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

## Update example

`PUT` is a full update. Include `metadata.resourceVersion` from the current object, otherwise Kubernetes will reject the request.

This example replaces the ConfigMap content.

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

## Patch example

`PATCH` accepts raw patch bodies. Supported patch types are:

- default: strategic merge patch
- `?patchType=merge`: JSON merge patch
- `?patchType=json`: JSON patch

This example uses JSON merge patch to update one field without sending the full object.

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

To restart a Deployment with `PATCH`, update an annotation under `spec.template.metadata.annotations`. This changes the Pod template and triggers a new rollout.

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

## Delete example

This example deletes the ConfigMap.

```bash
curl \
  -X DELETE \
  -H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "x-cluster-name: demo-cluster" \
  https://kite.example.com/api/v1/configmaps/default/example-config
```

Optional delete query parameters:

- `force=true`: delete with zero grace period
- `wait=false`: return immediately instead of waiting for deletion
- `cascade=false`: orphan dependents instead of foreground deletion

Example:

```bash
curl \
  -X DELETE \
  -H "Authorization: kite12-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "x-cluster-name: demo-cluster" \
  "https://kite.example.com/api/v1/deployments/default/example-app?force=true&wait=false"
```

## Cluster-scoped example

Cluster-scoped resources use `/_all` in the path.

This example patches a Namespace label:

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

## Custom resources

Custom resources use `/:crd/...` routes, where `:crd` is the CRD name such as `rollouts.argoproj.io`.

Examples:

- namespaced custom resource get: `/api/v1/rollouts.argoproj.io/default/example`
- cluster-scoped custom resource get: `/api/v1/<crd>/_all/example`

As of the current server routes:

- custom resources expose list and get routes
- custom resources expose update and delete routes
- custom resource create and patch are not documented here because the generic built-in resource routes are the stable CRUD surface currently registered for normal resource operations
