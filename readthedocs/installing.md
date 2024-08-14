# Installing

## Installing the go implementation

You need to have Go programming language installed (see [instructions](https://go.dev/doc/install)).

Assuming you want to install the interpreter to `~/bin/mg`, do the following:
```
GOBIN=~/bin go install github.com/google/mangle/interpreter/mg@latest
```

Then you can start the interpreter with `~/bin/mg`.

## Building the go implementation from source

```
git clone github.com/google/mangle
cd mangle
go get -t ./...
go build ./...
go test ./...
```

You can start the interpreter using `go run interpreter/mg/mg.go`

## Building the rust implementation from source

```
git clone github.com/google/mangle
cd mangle/rust
cargo build
cargo test
```

```{note}
The Rust implementation has no interactive interpreter yet.
```
