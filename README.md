# modserver

[![Go Reference](https://pkg.go.dev/badge/github.com/mercari/modserver.svg)](https://pkg.go.dev/github.com/mercari/modserver)

modserver is a simple Go module server. It serves a collection of modules from a local directory. modserver comes with its own command, or it can be used as an library. This makes it easy to spin up a local Go module server for testing purposes.

## Installation

To install the latest released version of modserver command, simply run the following.

### Go version < 1.16

```bash
$ go get -u github.com/mercari/modserver/cmd/modserver
```

### Go 1.16+

```bash
$ go install github.com/mercari/modserver/cmd/modserver@latest
```

## Usage

### Command

modserver command can be used to quickly run a Go module server for the given module collection directory and address.

```shell
$ modserver [mod-directory] [address]
```

### Library

modserver can be run using `http.Server` directly. The example below run shows how to use modserver to serve the given modules collection directory.

```go
import (
    "net/http"
    "github.com/mercari/modserver"
)

func main () {
    h, err := modserver.NewHandler(&modserver.Config{
        ModDir: "directory/to/collection",
    })
    if err != nil {
        panic("failed initializing handler")
    }

    srv := &http.Server{
        Addr: ":8080",
        Handler: h,
    }

    // ...
}
```

modserver also provides the convenience function `modserver.NewTestServer()` to initialize and start an `httptest.Server` instance in a testing environment, as shown in the example below.

```go
import (
    "testing"
    "github.com/mercari/modserver"
)

func Test(t *testing.T) {
    srv, err := modserver.NewTestServer(&modserver.Config{
        ModDir: "directory/to/collection",
    })
    if err != nil {
        t.Fatal("unexpected error:", err)
    }

    // ...
}
```

## Module Collection

The module collection directory has to follow a certain structure to be able to be recognized by modserver. The example below represents the module collection directory tree on `modules/`.

```
modules
└── github.com
    └── mercari
        ├── example@v0.1.0
        │   ├── go.mod
        │   └── ...
        ├── example@v1.0.0
        │   ├── go.mod
        │   └── ...
        └── example@v2.0.0+incompatible
            └── ...
```

The collection above contains the modules `github.com/mercari/example` with versions `v0.1.0`, `v1.0.0`, and `v2.0.0+incompatible`. Note that the versions has to be in [canonical form](https://go.dev/ref/mod#glos-canonical-version), because non-canonical versions will not be served by modserver.

## Setting Up Environment Variables

The address where modserver listens to is typically set on the `GOPROXY` environment variable. Note that modserver does not proxy requests for modules not contained in the collection, thus it will not resolve other external modules. Another module proxy (e.g. https://proxy.golang.org) may need to be added to this environment variable as a fallback, if this behavior is to be expected.

At the moment, modserver does not implement the Go checksum database. This will cause Go commands to fail verifying the checksums of the modules retrieved from modserver as they might not be found by the public checksum database. To circumvent this, `GONOSUMDB` environment variable must be set to the glob pattern of the modules served by modserver.

Refer to the [Go Modules Reference](https://go.dev/ref/mod#goproxy-protocol) for more detail.

## Contribution

Please read the CLA carefully before submitting your contribution to Mercari. Under any circumstances, by submitting your contribution, you are deemed to accept and agree to be bound by the terms and conditions of the CLA.

https://www.mercari.com/cla/

## License

Copyright 2022 Mercari, Inc.

This project is licensed under the [MIT License](./LICENSE).
