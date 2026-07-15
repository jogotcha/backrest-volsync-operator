# Changelog

## [0.3.1](https://github.com/jogotcha/backrest-volsync-operator/compare/v0.3.0...v0.3.1) (2026-07-15)


### Bug Fixes

* bump Go to 1.25.9 and teach Renovate to bump it ([#67](https://github.com/jogotcha/backrest-volsync-operator/issues/67)) ([e9201fe](https://github.com/jogotcha/backrest-volsync-operator/commit/e9201fe60b682acbf32601e7b9cd1507208b6ffe))
* dedupe same-snapshot stats triggers ([#82](https://github.com/jogotcha/backrest-volsync-operator/issues/82)) ([33f6468](https://github.com/jogotcha/backrest-volsync-operator/commit/33f6468b3cf7e45190a3a88de1f7afcd291bab94))
* **deps:** update k8s.io/utils digest to 28399d8 ([#60](https://github.com/jogotcha/backrest-volsync-operator/issues/60)) ([b7d23c6](https://github.com/jogotcha/backrest-volsync-operator/commit/b7d23c60abcf6509993b96b01d06367e1c224388))
* **deps:** update k8s.io/utils digest to a95e086 ([#88](https://github.com/jogotcha/backrest-volsync-operator/issues/88)) ([cc32336](https://github.com/jogotcha/backrest-volsync-operator/commit/cc32336f027e7bfb4cc80fc3ff6e39cd1672dcc7))
* **deps:** update k8s.io/utils digest to be93311 ([#91](https://github.com/jogotcha/backrest-volsync-operator/issues/91)) ([f686e9b](https://github.com/jogotcha/backrest-volsync-operator/commit/f686e9b9f5a293237024425774fb89a0ba1c861a))
* **deps:** update k8s.io/utils digest to cf1189d ([#92](https://github.com/jogotcha/backrest-volsync-operator/issues/92)) ([a507d95](https://github.com/jogotcha/backrest-volsync-operator/commit/a507d954ba5308e6aa8d5e9321a81c667c814af3))
* **deps:** update k8s.io/utils digest to ff6756f ([#76](https://github.com/jogotcha/backrest-volsync-operator/issues/76)) ([8540de0](https://github.com/jogotcha/backrest-volsync-operator/commit/8540de0bac0ba4bb70167c3e60129662e32b96e9))
* **deps:** update kubernetes monorepo to v0.35.3 ([#59](https://github.com/jogotcha/backrest-volsync-operator/issues/59)) ([8d80c3c](https://github.com/jogotcha/backrest-volsync-operator/commit/8d80c3cafceaf8959ad9d9996e4dc13c20016854))
* **deps:** update kubernetes monorepo to v0.36.0 ([#69](https://github.com/jogotcha/backrest-volsync-operator/issues/69)) ([f9e07b7](https://github.com/jogotcha/backrest-volsync-operator/commit/f9e07b7e199303d7442acbdc5bd115d2a7a82f65))
* **deps:** update kubernetes monorepo to v0.36.1 ([#78](https://github.com/jogotcha/backrest-volsync-operator/issues/78)) ([4b73f17](https://github.com/jogotcha/backrest-volsync-operator/commit/4b73f17ebbf81d2d51aede6fb9b852f6d345e849))
* **deps:** update kubernetes monorepo to v0.36.2 ([#87](https://github.com/jogotcha/backrest-volsync-operator/issues/87)) ([40449b9](https://github.com/jogotcha/backrest-volsync-operator/commit/40449b9b44a9789b7561cb4b13815f5a02782d40))
* **deps:** update module connectrpc.com/connect to v1.20.0 ([#70](https://github.com/jogotcha/backrest-volsync-operator/issues/70)) ([72e58df](https://github.com/jogotcha/backrest-volsync-operator/commit/72e58df177497f45f175695af21da4ee7570e23e))
* **deps:** update module github.com/garethgeorge/backrest to v1.12.1 ([#58](https://github.com/jogotcha/backrest-volsync-operator/issues/58)) ([18131ad](https://github.com/jogotcha/backrest-volsync-operator/commit/18131ad02f11ba737ecb693f90be98541f5a6556))
* **deps:** update module github.com/garethgeorge/backrest to v1.13.0 ([#75](https://github.com/jogotcha/backrest-volsync-operator/issues/75)) ([ea9a3a3](https://github.com/jogotcha/backrest-volsync-operator/commit/ea9a3a3e2360bf3558f9a9d152f50cf3363c4e59))
* **deps:** update module github.com/garethgeorge/backrest to v1.14.1 ([#94](https://github.com/jogotcha/backrest-volsync-operator/issues/94)) ([87fea46](https://github.com/jogotcha/backrest-volsync-operator/commit/87fea4626bf3025f2e17b57cf41db0911086f740))
* **deps:** update module go.uber.org/zap to v1.28.0 ([#73](https://github.com/jogotcha/backrest-volsync-operator/issues/73)) ([591a9ba](https://github.com/jogotcha/backrest-volsync-operator/commit/591a9ba09f3067277507340d05df2c1761bfec03))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.24.1 ([#81](https://github.com/jogotcha/backrest-volsync-operator/issues/81)) ([ed85eb7](https://github.com/jogotcha/backrest-volsync-operator/commit/ed85eb7f06bc8ce49d2d07721b2d385d8c444cad))

## [0.3.0](https://github.com/jogotcha/backrest-volsync-operator/compare/v0.2.2...v0.3.0) (2026-03-08)


### Features

* trigger Backrest INDEX_SNAPSHOTS and STATS from VolSync snapshots ([#42](https://github.com/jogotcha/backrest-volsync-operator/issues/42)) ([a4ae8b2](https://github.com/jogotcha/backrest-volsync-operator/commit/a4ae8b2b124900e81fec45bc820226f7e41f4be9))


### Bug Fixes

* **ci:** release please push to main should not trigger build-image-push ([31f1a01](https://github.com/jogotcha/backrest-volsync-operator/commit/31f1a01abc0afc7bb859abe97473a76cddb38dbc))
* **deps:** update kubernetes packages to v0.35.2 ([#43](https://github.com/jogotcha/backrest-volsync-operator/issues/43)) ([d2ec1fb](https://github.com/jogotcha/backrest-volsync-operator/commit/d2ec1fb9e12d0bafd9b5d960625b7d3e63ee1bac))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.23.3 ([#50](https://github.com/jogotcha/backrest-volsync-operator/issues/50)) ([b9ff8dc](https://github.com/jogotcha/backrest-volsync-operator/commit/b9ff8dc1fd09b74edb19509c7a620ba9f1b31384))
* **makefile:** make lint cross-platform ([#45](https://github.com/jogotcha/backrest-volsync-operator/issues/45)) ([491f0b9](https://github.com/jogotcha/backrest-volsync-operator/commit/491f0b900bd1678f14b08a73d906575859db1638))
* Stabilize VolSync marker to dedupe repo task triggers ([#44](https://github.com/jogotcha/backrest-volsync-operator/issues/44)) ([378e881](https://github.com/jogotcha/backrest-volsync-operator/commit/378e8814f63ee504b210c96edcb25a2c4cf7a9a9))

## [0.2.2](https://github.com/jogotcha/backrest-volsync-operator/compare/v0.2.1...v0.2.2) (2026-02-23)


### Bug Fixes

* bump Go toolchain to 1.25.7 ([#35](https://github.com/jogotcha/backrest-volsync-operator/issues/35)) ([6ca2130](https://github.com/jogotcha/backrest-volsync-operator/commit/6ca2130e7c9e35e20eb1a4ee3433cfcfe24e7449))
* **deps:** update module github.com/garethgeorge/backrest to v1.12.0 ([#38](https://github.com/jogotcha/backrest-volsync-operator/issues/38)) ([e7834b5](https://github.com/jogotcha/backrest-volsync-operator/commit/e7834b5a8288b9ce0a433c0c193441ea4ac0b0ba))

## [0.2.1](https://github.com/jogotcha/backrest-volsync-operator/compare/v0.2.0...v0.2.1) (2026-02-17)


### Bug Fixes

* **deps:** update k8s.io/utils digest to 914a6e7 ([#15](https://github.com/jogotcha/backrest-volsync-operator/issues/15)) ([1bb793d](https://github.com/jogotcha/backrest-volsync-operator/commit/1bb793dc2cbab1f3e51f983a5412f64496306464))
* **deps:** update k8s.io/utils digest to b8788ab ([#28](https://github.com/jogotcha/backrest-volsync-operator/issues/28)) ([a285cf6](https://github.com/jogotcha/backrest-volsync-operator/commit/a285cf6cb8825c55eeb49f4915f7cfd527cce242))
* **deps:** update kubernetes packages to v0.35.1 ([#29](https://github.com/jogotcha/backrest-volsync-operator/issues/29)) ([b589642](https://github.com/jogotcha/backrest-volsync-operator/commit/b5896427fffc1f2dad4564d878611917b12f1fe5))
* **deps:** update module github.com/garethgeorge/backrest to v1.11.1 ([#16](https://github.com/jogotcha/backrest-volsync-operator/issues/16)) ([c47f785](https://github.com/jogotcha/backrest-volsync-operator/commit/c47f7856b88bb82f46d134e90ddcc18c04c7ecca))
* **deps:** update module github.com/garethgeorge/backrest to v1.11.2 ([#24](https://github.com/jogotcha/backrest-volsync-operator/issues/24)) ([d22a048](https://github.com/jogotcha/backrest-volsync-operator/commit/d22a048789addd9652c7806b014e0f52045ef665))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.23.0 ([#17](https://github.com/jogotcha/backrest-volsync-operator/issues/17)) ([e796a72](https://github.com/jogotcha/backrest-volsync-operator/commit/e796a723f0e1cb6d2004c11f2eacbb7031ad5b31))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.23.1 ([#23](https://github.com/jogotcha/backrest-volsync-operator/issues/23)) ([cb29557](https://github.com/jogotcha/backrest-volsync-operator/commit/cb29557cad2ab547e781f935a773a275d188f021))
* **rbac:** allow creating events in events.k8s.io ([#32](https://github.com/jogotcha/backrest-volsync-operator/issues/32)) ([694fb73](https://github.com/jogotcha/backrest-volsync-operator/commit/694fb731fa857dc0f656e8e6e47b7b1321db336c)), closes [#27](https://github.com/jogotcha/backrest-volsync-operator/issues/27)

## [0.2.0](https://github.com/jogotcha/backrest-volsync-operator/compare/v0.1.1...v0.2.0) (2025-12-30)


### Features

* add end-to-end testing workflow and comprehensive unit tests for BackrestVolSyncBinding and VolSyncAutoBinding controllers ([2157b8f](https://github.com/jogotcha/backrest-volsync-operator/commit/2157b8f1c32a061ddff48997b59c1f0acd4109a9))
* add event recorder to BackrestVolSyncBinding and VolSyncAutoBinding controllers ([61b1c4e](https://github.com/jogotcha/backrest-volsync-operator/commit/61b1c4e3127ad80f91c812f55f4351b39209ba0b))
* add OperatorConfig and VolSync auto-binding ([4f740af](https://github.com/jogotcha/backrest-volsync-operator/commit/4f740afa30ed35be23060c39f68a781914795e83))
* enhance CI workflows with CVE scanning and testing, update Dockerfile for multi-architecture support ([aee6da5](https://github.com/jogotcha/backrest-volsync-operator/commit/aee6da570c98c6b67ded3fa775aeba633f53865f))
* refactor GitHub Actions workflows for CI and E2E testing, add caching and build options ([4f3e12d](https://github.com/jogotcha/backrest-volsync-operator/commit/4f3e12de3ef71d781331c1c96bf039772279eebe))


### Bug Fixes

* Change VolSyncAutoBindingReconciler to use Watches instead of For() for unstructured resources ([4267822](https://github.com/jogotcha/backrest-volsync-operator/commit/4267822c2717c5fc349ad334c7beb4256f39b79a))
* correct indentation in backrestvolsyncbinding.yaml and update CPU limits in deployment.yaml ([e41fc88](https://github.com/jogotcha/backrest-volsync-operator/commit/e41fc8866fd257d074dcf073134a7a7c1307e044))
* **deps:** update k8s.io/utils digest to 718f0e5 ([#2](https://github.com/jogotcha/backrest-volsync-operator/issues/2)) ([05be5bc](https://github.com/jogotcha/backrest-volsync-operator/commit/05be5bc7b5f3a38b72cfdf64219c21c36c80ac6b))
* **deps:** update kubernetes packages to v0.35.0 ([#9](https://github.com/jogotcha/backrest-volsync-operator/issues/9)) ([ea3233c](https://github.com/jogotcha/backrest-volsync-operator/commit/ea3233c404793915fb4d78a440209ad47b20d4af))
* **deps:** update module go.uber.org/zap to v1.27.1 ([#3](https://github.com/jogotcha/backrest-volsync-operator/issues/3)) ([40f5adc](https://github.com/jogotcha/backrest-volsync-operator/commit/40f5adc4f3d2ffd1ba6c8db0ac9872ead38eced1))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.22.4 ([#5](https://github.com/jogotcha/backrest-volsync-operator/issues/5)) ([528814f](https://github.com/jogotcha/backrest-volsync-operator/commit/528814ff11e16c51500b27c53945cc9a93832490))
* downgrade setup-go action version and specify golangci-lint version for consistency ([57a284d](https://github.com/jogotcha/backrest-volsync-operator/commit/57a284de8750c6147f91aff15ec15894701c9d51))
* optimize Dockerfile by adding caching for go mod download and build steps ([80c24a5](https://github.com/jogotcha/backrest-volsync-operator/commit/80c24a5e38fbf2ae804d430910c5a5c6d00bd0bb))
* update autoUnlock behavior and documentation across configurations and examples ([bbcc208](https://github.com/jogotcha/backrest-volsync-operator/commit/bbcc2089e08d8dfddad1874b8ba27b239b922e5d))
* update Dockerfile to support multi-platform builds ([ba9551c](https://github.com/jogotcha/backrest-volsync-operator/commit/ba9551cba213a2928dbd50d815785f510255889f))
* update permissions to include write access for packages ([4d865c6](https://github.com/jogotcha/backrest-volsync-operator/commit/4d865c674f2f00a86d319eb42ff0c54ad835b607))
* update release-please token handling and fix image tag format ([1f1247e](https://github.com/jogotcha/backrest-volsync-operator/commit/1f1247ec9ad1d5afc56c8e838cd46b71084b947c))
* update skip-ci condition in GHCR workflows to use toJson for better handling [skip-ci] ([520c251](https://github.com/jogotcha/backrest-volsync-operator/commit/520c2512832e48aaa8320cf8143f068ad6fc74aa))
