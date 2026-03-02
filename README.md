# Go IMS Client

[![CI](https://github.com/adobe/ims-go/actions/workflows/ci.yml/badge.svg)](https://github.com/adobe/ims-go/actions/workflows/ci.yml)
[![govulncheck](https://github.com/adobe/ims-go/actions/workflows/govulncheck.yml/badge.svg)](https://github.com/adobe/ims-go/actions/workflows/govulncheck.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/adobe/ims-go)](https://goreportcard.com/report/github.com/adobe/ims-go)
![Go Version](https://img.shields.io/github/go-mod/go-version/adobe/ims-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/adobe/ims-go/ims.svg)](https://pkg.go.dev/github.com/adobe/ims-go/ims)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

This project is a Go library for accessing the IMS API. 
The goal of this project is to provide an easy-to-use binding to the IMS API and a set of common utilities for working efficiently with IMS.

## Installation

Use the standard Go toolchain to use this library in your project.

Example:

```
go get -u github.com/adobe/ims-go
```

## Usage

Once installed, you can start interacting with IMS by instantiating a new client.

Example:

```go
import "github.com/adobe/ims-go/ims"

c, err := ims.NewClient(&ims.ClientConfig{
    URL: imsEndpoint,
})
```

## Contributing

Contributions are welcomed! 
Read the [Contributing Guide](./.github/CONTRIBUTING.md) for more information.

## Licensing

This project is licensed under the Apache V2 License. 
See [LICENSE](LICENSE) for more information.
