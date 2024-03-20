# DryDock
### Simple Scaffolding Library for Go

[![MIT License](https://img.shields.io/github/license/RobinThrift/drydock?style=flat-square)](https://github.com/RobinThrift/drydock/blob/main/LICENSE)
![CI](https://github.com/RobinThrift/drydock/actions/workflows/ci.yaml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/RobinThrift/drydock.svg)](https://pkg.go.dev/github.com/RobinThrift/drydock)
[![Latest Release](https://img.shields.io/github/v/tag/RobinThrift/drydock?sort=semver&style=flat-square)](https://github.com/RobinThrift/drydock/releases/latest)



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

[MIT](https://github.com/RobinThrift/drydock/blob/main/LICENSE)
