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
kubectl apply -k config/
kubectl apply -f config/samples.yaml
```

## Install with Helm (recommended)

This repo includes a Helm chart at `charts/backrest-volsync-operator`.

Install into any namespace (the chart installs into `.Release.Namespace`):

```sh
NAMESPACE=backups
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

helm install backrest-volsync-operator \
  oci://ghcr.io/<owner>/charts/backrest-volsync-operator \
  --namespace "$NAMESPACE" \
  --create-namespace
```

Override the operator image if needed:

```sh
helm upgrade --install backrest-volsync-operator \
  oci://ghcr.io/<owner>/charts/backrest-volsync-operator \
  --namespace "$NAMESPACE" \
  --set image.repository=ghcr.io/<owner>/<repo> \
  --set image.tag=latest
```

Examples are not installed by default. See `charts/backrest-volsync-operator/examples/`.

## Backrest auth (optional)

Create a Secret referenced by `spec.backrest.authRef.name` in the same namespace as the Binding:

- Bearer token:
  - key: `token`
- Basic auth:
  - keys: `username`, `password`

## Build + deploy

This operator is standalone (no local `replace ../backrest`).

### Local build

```sh
make fmt
make test
```

Build and push an image to a registry reachable by your cluster nodes (Talos uses containerd):

```sh
make docker-build IMAGE=backrest-volsync-operator:dev
```

### GitHub Actions (GHCR)

The workflow builds/pushes `ghcr.io/<owner>/<repo>` on pushes to `main` and tags.
If your repo name differs, update the default image in `config/deployment.yaml` (kustomize install) or override `image.*` via Helm.

#### Helm chart publishing (GHCR OCI)

The workflow in `.github/workflows/helm-chart-publish-oci.yaml` publishes the Helm chart to:

`oci://ghcr.io/<owner>/charts/backrest-volsync-operator`

To publish a new chart version, create a tag in the form `chart-vX.Y.Z` and push it:

```sh
git tag chart-v0.1.0
git push origin chart-v0.1.0
```

### Deploy (kustomize)

Apply the operator resources:

```sh
kubectl apply -k config/
```

If pulling from a private GHCR image, create an image pull secret and uncomment `imagePullSecrets` in `config/deployment.yaml`:

```sh
kubectl -n backups create secret docker-registry ghcr-pull \
  --docker-server=ghcr.io \
  --docker-username=<github-username> \
  --docker-password=<github-pat-with-read:packages> \
  --docker-email=unused@example.com
```

```sh
kubectl apply -f config/deployment.yaml
```

## Create a Binding

```yaml
apiVersion: backrest.garethgeorge.com/v1alpha1
kind: BackrestVolSyncBinding
metadata:
  name: my-app
  namespace: backups
spec:
  backrest:
    url: http://backrest.backups.svc:9898
  source:
    kind: ReplicationSource
    name: my-app-data
  repo:
    idOverride: my-app-data
    autoUnlock: true

```

## OperatorConfig (auto-create Bindings)

This operator can optionally auto-create `BackrestVolSyncBinding` objects from VolSync `ReplicationSource` / `ReplicationDestination` objects.

Create a `BackrestVolSyncOperatorConfig` in the same namespace as the operator:

```yaml
apiVersion: backrest.garethgeorge.com/v1alpha1
kind: BackrestVolSyncOperatorConfig
metadata:
  name: backrest-volsync-operator
  namespace: backups
spec:
  # Global kill-switch (disables Backrest API calls and auto-binding).
  paused: false

  # Defaults applied to generated Bindings.
  defaultBackrest:
    url: http://backrest.backups.svc:9898
    # authRef:
    #   name: backrest-auth

  bindingGeneration:
    # Disabled | Annotated | All
    policy: Annotated
    # Optional restrict which VolSync kinds are eligible for auto-binding.
    # Allowed values: ReplicationSource | ReplicationDestination
    # If omitted/empty, both kinds are allowed.
    # kinds: [ReplicationSource, ReplicationDestination]
    defaultRepo:
      autoUnlock: true
```

### Policy behavior

- `Disabled`: do not auto-create bindings.
- `Annotated`: only auto-create when the VolSync object has `backrest.garethgeorge.com/binding: "true"`.
- `All`: auto-create for all VolSync objects (unless opted out).

### VolSync opt-in / opt-out annotation

Set this annotation on a VolSync `ReplicationSource` / `ReplicationDestination`:

```yaml
metadata:
  annotations:
    backrest.garethgeorge.com/binding: "true"  # opt-in (for Annotated policy)
```

To force opt-out (even when policy is `All`):

```yaml
metadata:
  annotations:
    backrest.garethgeorge.com/binding: "false"
```
```

## Notes / security

Backrest stores `Repo.password` as plaintext in its config file. The operator avoids calling `GetConfig` and never writes credentials into status or logs, but Backrest’s own at-rest format remains unchanged.

Notes:

- The operator intentionally does **not** log or store secret values. If a reconcile fails, status will contain only non-sensitive metadata and an error hash.
- On success, the operator logs `Backrest repo applied` with only the repo ID + VolSync reference.
