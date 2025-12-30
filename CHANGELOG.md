# Changelog

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
