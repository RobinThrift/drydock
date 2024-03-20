# DryDock
### Simple Scaffolding Library for Go

[![MIT License](https://img.shields.io/github/license/spacefleet-dev/drydock?style=flat-square)](https://github.com/spacefleet-dev/drydock/blob/main/LICENSE)
![CI](https://github.com/spacefleet-dev/drydock/actions/workflows/ci.yaml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/spacefleet-dev/drydock.svg)](https://pkg.go.dev/github.com/spacefleet-dev/drydock)
[![Latest Release](https://img.shields.io/github/v/tag/spacefleet-dev/drydock?sort=semver&style=flat-square)](https://github.com/spacefleet-dev/drydock/releases/latest)



## Usage Example

```go
g := &FSGenerator{
	FS: NewWritableDirFS("out"),
}

err := g.Generate(
    context.Background(),
	PlainFile("README.md", "# drydock"),
	Dir("bin",
		Dir("cli",
			PlainFile("main.go", "package main"),
		),
	),
	Dir("pkg",
		PlainFile("README.md", "how to use this thing"),
		Dir("cli",
			PlainFile("cli.go", "package cli..."),
			PlainFile("run.go", "package cli...run..."),
		),
	),
)
````

## License

[MIT](https://github.com/spacefleet-dev/drydock/blob/main/LICENSE)
