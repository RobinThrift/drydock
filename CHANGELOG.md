# Changelog

## [0.1.0](https://github.com/RobinThrift/drydock/releases/tag/v0.1.0) - 2026-01-08

### <!-- 0 -->New Features

- Add render function to visualise file tree [36d30771d1](https://github.com/RobinThrift/drydock/commit/36d30771d178e475fe3b38bc8b5da9de3ac503eb)(@Robin Thrift)

### <!-- 0 -->New Features (WritableFS)

- Add function to create WritableFS directory and any non-existent parent directory [e9bc305bd7](https://github.com/RobinThrift/drydock/commit/e9bc305bd70363741eb1d04876d06ab07c2873b9)(@Robin Thrift)

### <!-- 1 -->Fixes

- Fix DirP creating directories in reverse order [cbbf03ff1d](https://github.com/RobinThrift/drydock/commit/cbbf03ff1def3312e05c49925667fd2140671791)(@Robin Thrift)
- Add extra test for DirP [46e4e6c25a](https://github.com/RobinThrift/drydock/commit/46e4e6c25a7d8e1ccf11598779c94529d4ca2c3d)(@Robin Thrift)
- Fix current directory in DirP causing issues when ErrorOnExistingDir is true [efdc34a683](https://github.com/RobinThrift/drydock/commit/efdc34a6831972afe0328a0692cd15efcdba6d11)(@Robin Thrift)
- Remove golangci-lint from direct dev depdendencies [2b2359b5c0](https://github.com/RobinThrift/drydock/commit/2b2359b5c03abd0c71cd555cf26c5b4ed151f7f1)(@Robin Thrift)
- Fix automatic installation of golangci-lint [6d8517b8a4](https://github.com/RobinThrift/drydock/commit/6d8517b8a4e0446f3810eff6da4ecf52a4d0b774)(@Robin Thrift)

### <!-- 1 -->Fixes (build)

- Typo [3805767901](https://github.com/RobinThrift/drydock/commit/3805767901d81a74a1e4171a68b6590bc4ac3335)(@Robin Thrift)

### <!-- 2 -->Documentation

- Fix example in README [faa9f33c76](https://github.com/RobinThrift/drydock/commit/faa9f33c76938845bfda0be07a6a5a039e09c686)(@Robin Thrift)
- Remove duplicate badged in README [9201d609b3](https://github.com/RobinThrift/drydock/commit/9201d609b3d5ac94114aff569dfaae140ccbb027)(@Robin Thrift)

### <!-- 3 -->Dependencies (GitHub Actions)

- Upgrade Go and Just versions used in CI [f14256602f](https://github.com/RobinThrift/drydock/commit/f14256602fc447512b0625a2046d8447001006cc)(@Robin Thrift)

### <!-- 4 -->Refactored

- Move to functional options pattern and simplify generator and output filesystem requirements [a318b77d42](https://github.com/RobinThrift/drydock/commit/a318b77d4206c6c85797e28d57f4e591e767e6b8)(@Robin Thrift)
- Replace the Rename method with the Move method to allow for more flexible semantics [b7eb9ee7ff](https://github.com/RobinThrift/drydock/commit/b7eb9ee7ff6d4f560ca4b483b208ac20b70f9e31)(@Robin Thrift)

### <!-- 5 -->Removed

- Remove dead commented out code [f85110b733](https://github.com/RobinThrift/drydock/commit/f85110b733c325f669f4eaaa6aeceeb539623d1a)(@Robin Thrift)

### <!-- 7 -->Internal

- Init commit [6012e980ec](https://github.com/RobinThrift/drydock/commit/6012e980ec3c42d7d30353b2c92eb0e18a8f23d2)(@Robin Thrift)
- Fix changelog generation [0bbe62a05e](https://github.com/RobinThrift/drydock/commit/0bbe62a05e32ac5888db5ffb704249f59cfbc524)(@Robin Thrift)
- Add GitHub Actions output to test [8c7cbe6349](https://github.com/RobinThrift/drydock/commit/8c7cbe6349249f21ec03b2a4784d9d1eca33fbfe)(@Robin Thrift)
- Move tools to own package [3c0382e98d](https://github.com/RobinThrift/drydock/commit/3c0382e98d3b023689a0bb593fa2fbb7013c1081)(@Robin Thrift)
- Fix changelog categories [00ea4f6da6](https://github.com/RobinThrift/drydock/commit/00ea4f6da6bdb55ed369e3c51824a7b40637ea39)(@Robin Thrift)
- Add concurrency control to GitHub actions to prevent duplicate runs [b02a5bc2ec](https://github.com/RobinThrift/drydock/commit/b02a5bc2ec54b2f767a3f28be1a5e06145339e27)(@Robin Thrift)
- Fix CI GitHub Action workflow, move reference to code after checkout [d4b2440104](https://github.com/RobinThrift/drydock/commit/d4b24401040c41338d1ecb2609ab6a771d81ff81)(@Robin Thrift)
- Fix missing permissions for GitHub Actions CI workflow [3018211c07](https://github.com/RobinThrift/drydock/commit/3018211c07fb8cecf216ecb0fa11320390785f45)(@Robin Thrift)
- Rename ci workflow file to match title and fix badge [7f488216c1](https://github.com/RobinThrift/drydock/commit/7f488216c1566655234ef1b9bf68499a529a1518)(@Robin Thrift)
- Switch from odeshack/action-install-gh-release to RobinThrift/action-setup-just action to install just in CI [cea04626bd](https://github.com/RobinThrift/drydock/commit/cea04626bd8c5963c75737df1e4479fa0a804158)(@Robin Thrift)

### <!-- 8 -->Other Changes

- Bump github.com/golangci/golangci-lint in /.tools [809fe6e537](https://github.com/RobinThrift/drydock/commit/809fe6e5379d2764012f99a1e9acc03e0bb4170d)(@dependabot[bot])
- Bump github.com/stretchr/testify from 1.9.0 to 1.10.0 [4e8ce90273](https://github.com/RobinThrift/drydock/commit/4e8ce902731afb9a7f8ed5455a10e11c74f54d41)(@dependabot[bot])
- Bump github.com/golangci/golangci-lint in /.tools [f97a980515](https://github.com/RobinThrift/drydock/commit/f97a9805151fe57f3115de7ab183ca07d0425388)(@dependabot[bot])
- Bump github.com/golangci/golangci-lint in /.tools [466c86d647](https://github.com/RobinThrift/drydock/commit/466c86d6474c3607775c47d6a7da388d170462ff)(@dependabot[bot])
- Upgrade tooling to newest versions [6593e75afa](https://github.com/RobinThrift/drydock/commit/6593e75afacf92a082ef7ae2880e73c29934476a)(@Robin Thrift)

