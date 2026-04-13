# Project Guidelines

## Architecture

- This repository is a Go Kubernetes operator that binds VolSync restic repositories to Backrest.
- API and CRD types live in `api/v1alpha1`. Keep new API behavior consistent with the existing hand-written types and status fields.
- The main binding reconciliation logic is in `controllers/backrestvolsyncbinding_controller.go`.
- Auto-generation of managed bindings is handled separately in `controllers/volsync_autobinding_controller.go` and is gated by `BackrestVolSyncOperatorConfig` loaded via `controllers/operatorconfig.go`.
- Backrest API integration lives in `pkg/backrest/client.go`. VolSync object parsing helpers live in `pkg/volsync/extract.go` and operate on unstructured objects.
- Deployment and packaging assets are committed directly under `config/` and `charts/backrest-volsync-operator/`.

## Build And Test

- Run `make test` after Go code changes. This is the primary project test command.
- Run `make fmt` after editing Go files.
- Run `make lint` when `golangci-lint` is installed. On Windows or other environments without it, the target intentionally skips.
- Use `make docker-build` only when validating image build behavior.
- There is no code generation workflow configured. Do not assume `controller-gen` or generated CRDs will be refreshed automatically.

## Conventions

- Prefer small, focused reconciler changes that preserve existing status and condition semantics.
- Keep VolSync kind handling explicit. The code expects exact kinds such as `ReplicationSource` and `ReplicationDestination`.
- When editing API types in `api/v1alpha1`, update the manual `DeepCopyInto` implementations as needed instead of introducing generated deepcopy code.
- Preserve the separation between user-managed bindings and operator-managed bindings. Auto-binding logic must not mutate bindings unless they are marked managed.
- Controller tests use `controller-runtime`'s fake client and targeted fixtures. Follow the style in `controllers/backrestvolsyncbinding_controller_test.go` and related tests.
- Favor repo-specific tests when changing controller behavior, status transitions, or VolSync extraction logic.

## Operational Pitfalls

- Backrest repo task calls can be slow and effectively synchronous. The Backrest client uses a 2 minute HTTP timeout; avoid reintroducing short timeouts.
- Snapshot-triggered repo tasks are sensitive to duplicate reconcile paths. Preserve the existing marker and in-flight guard behavior when changing snapshot task logic.
- Missing `BackrestVolSyncOperatorConfig` is treated as a safe default with auto-binding disabled.
- CRDs and Helm chart assets are versioned in-repo. When changing API surface or install behavior, update both operator manifests and chart assets deliberately.

## Commit Hygiene

- Use semantic commits for any commit messages and PR titles, following Conventional Commit style such as `feat:`, `fix:`, `docs:`, `test:`, or `refactor:`.