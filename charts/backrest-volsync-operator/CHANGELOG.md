# Changelog

## [0.3.0](https://github.com/jogotcha/backrest-volsync-operator/compare/chart-v0.2.1...chart-v0.3.0) (2026-02-28)


### Features

* trigger Backrest INDEX_SNAPSHOTS and STATS from VolSync snapshots ([#42](https://github.com/jogotcha/backrest-volsync-operator/issues/42)) ([a4ae8b2](https://github.com/jogotcha/backrest-volsync-operator/commit/a4ae8b2b124900e81fec45bc820226f7e41f4be9))


### Bug Fixes

* **chart:** default image tag to Chart appVersion ([#37](https://github.com/jogotcha/backrest-volsync-operator/issues/37)) ([8012332](https://github.com/jogotcha/backrest-volsync-operator/commit/80123327e87f3dce23949605d1d4a29521f3ae88))

## [0.2.1](https://github.com/jogotcha/backrest-volsync-operator/compare/chart-v0.2.0...chart-v0.2.1) (2026-02-17)


### Bug Fixes

* **rbac:** allow creating events in events.k8s.io ([#32](https://github.com/jogotcha/backrest-volsync-operator/issues/32)) ([694fb73](https://github.com/jogotcha/backrest-volsync-operator/commit/694fb731fa857dc0f656e8e6e47b7b1321db336c)), closes [#27](https://github.com/jogotcha/backrest-volsync-operator/issues/27)

## [0.2.0](https://github.com/jogotcha/backrest-volsync-operator/compare/chart-v0.1.0...chart-v0.2.0) (2025-12-30)


### Features

* add OperatorConfig and VolSync auto-binding ([4f740af](https://github.com/jogotcha/backrest-volsync-operator/commit/4f740afa30ed35be23060c39f68a781914795e83))


### Bug Fixes

* correct indentation in backrestvolsyncbinding.yaml and update CPU limits in deployment.yaml ([e41fc88](https://github.com/jogotcha/backrest-volsync-operator/commit/e41fc8866fd257d074dcf073134a7a7c1307e044))
* update autoUnlock behavior and documentation across configurations and examples ([bbcc208](https://github.com/jogotcha/backrest-volsync-operator/commit/bbcc2089e08d8dfddad1874b8ba27b239b922e5d))
* update release-please token handling and fix image tag format ([1f1247e](https://github.com/jogotcha/backrest-volsync-operator/commit/1f1247ec9ad1d5afc56c8e838cd46b71084b947c))
