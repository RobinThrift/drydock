local_bin := absolute_path(".bin")

_default:
    @just --list

test +flags="-failfast": _install-tools
    {{ local_bin }}/gotestsum --format short-verbose -- {{ flags }} ./... 

alias tw := test-watch
test-watch +flags="-failfast": _install-tools
    {{ local_bin }}/gotestsum --format short-verbose --watch -- {{ flags }} ./...

test-ci format="short-verbose": _install-tools
    {{ local_bin }}/gotestsum --format {{ format }} --junitfile=test.junit.xml -- -timeout 10m ./...

lint: _install-tools
    {{ local_bin }}/staticcheck ./...
    {{ local_bin }}/golangci-lint run ./...

lint-ci: _install-tools
    {{ local_bin }}/golangci-lint run --timeout 5m --out-format=junit-xml ./... > lint.junit.xml
    {{ local_bin }}/staticcheck ./...

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
    git-cliff --config .tools/cliff.toml --prepend CHANGELOG.md --unreleased --tag {{ tag }}

_install-tools:
    @just _install-tool golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint
    @just _install-tool staticcheck honnef.co/go/tools/cmd/staticcheck
    @just _install-tool gotestsum gotest.tools/gotestsum

_install-tool bin mod:
    @[ -f {{ local_bin }}/{{bin}} ] || (cd .tools && GOBIN={{ local_bin }} go install -mod=readonly {{mod}})
