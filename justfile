gobin := absolute_path(".gobin")

_default:
    @just --list

test +flags="-failfast": _install-tools
    {{ gobin }}/gotestsum --format short-verbose -- {{ flags }} ./... 

alias tw := test-watch
test-watch +flags="-failfast": _install-tools
    {{ gobin }}/gotestsum --format short-verbose --watch -- {{ flags }} ./...

test-ci: _install-tools
    {{ gobin }}/gotestsum --format short-verbose --junitfile=test.junit.xml -- -timeout 10m ./...

lint: _install-tools
    {{ gobin }}/staticcheck ./...
    {{ gobin }}/golangci-lint run ./...

lint-ci: _install-tools
    {{ gobin }}/golangci-lint run --timeout 5m --out-format=junit-xml ./... > lint.junit.xml
    {{ gobin }}/staticcheck ./...

fmt:
	@go fmt ./...

clean:
	go clean -cache

release tag:
    just changelog {{ tag }}
    git add CHANGELOG.md
    git commit -m "release: Releasing version {{tag}}"
    git tag {{tag}}
    git push
    git push origin {{tag}}

changelog tag:
    git-cliff --config cliff.toml --prepend CHANGELOG.md --unreleased --tag {{ tag }}

_install-tools:
    @[ -x .bin/golangci-lint ] || gobin={{ gobin }} go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
    @just _install-tool gotestsum gotest.tools/gotestsum
    @just _install-tool staticcheck honnef.co/go/tools/cmd/staticcheck

_install-tool bin mod:
    @[ -f .bin/{{bin}} ] || GOBIN={{ gobin }} go install -mod=readonly {{mod}}
