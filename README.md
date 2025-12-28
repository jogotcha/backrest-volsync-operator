# Backrest VolSync Operator (prototype)

Prototype Kubernetes operator that watches VolSync `ReplicationSource` / `ReplicationDestination` objects and registers the referenced Restic repository in Backrest via the Backrest API.

## What it does

- Reads `spec.restic.repository` from a VolSync object.
- Reads the referenced Secret keys:
  - `RESTIC_REPOSITORY` -> Backrest repo URI
  - `RESTIC_PASSWORD` -> Backrest repo password
  - All other non-`RESTIC_*` keys -> Backrest `repo.env` as `KEY=value` (or restrict via `spec.repo.envAllowlist`).
- Calls Backrest `AddRepo` (upsert) using the generated Connect client.
- Stores only a hash + non-sensitive status fields in the CR status.

## Install CRD + RBAC + sample

```sh
kubectl apply -f config/crd.yaml
kubectl apply -f config/rbac.yaml
kubectl apply -f config/samples.yaml
```

## Backrest auth (optional)

Create a Secret referenced by `spec.backrest.authRef.name` in the same namespace as the Binding:

- Bearer token:
  - key: `token`
- Basic auth:
  - keys: `username`, `password`

## Build + deploy

This repo is structured as a submodule under `operator/` so it can be moved to its own repository later.

Build and push an image to a registry reachable by your cluster nodes (Talos uses containerd):

```sh
docker build -t <registry>/backrest-volsync-operator:dev -f Dockerfile .
docker push <registry>/backrest-volsync-operator:dev
```

Then update `config/deployment.yaml` `spec.template.spec.containers[0].image` to match and apply:

```sh
kubectl apply -f config/deployment.yaml
```

## Create a Binding

```yaml
apiVersion: backrest.garethgeorge.com/v1alpha1
kind: BackrestVolSyncBinding
metadata:
  name: my-app
  namespace: monitoring
spec:
  backrest:
    url: http://backrest.monitoring.svc:9898
  source:
    kind: ReplicationSource
    name: my-app-data
  repo:
    idOverride: my-app-data
    autoUnlock: true
```

## Notes / security

Backrest stores `Repo.password` as plaintext in its config file. The operator avoids calling `GetConfig` and never writes credentials into status or logs, but Backrest’s own at-rest format remains unchanged.
