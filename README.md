# grpc-service

Endpoints every gRPC server should expose: health, reflection, info, and
dependency diagnostics, assembled in one call.

## About

[grpc-foundation](https://github.com/sonic-original-software/grpc-foundation)
decides how a server serves: its options, telemetry, errors, and shutdown. This
library decides what a server exposes beyond its own services. A caller that
wants the first without the second takes foundation alone.

The contract a consumer signs up for are the two protos in
[grpc-protos](https://github.com/sonic-original-software/grpc-protos) plus the
mounted set:

- `grpc.health.v1.Health` — the standard health service
- `grpc.reflection.v1.ServerReflection` — the standard reflection service
- `info.InfoService` — the server's version, read from `GRPC_SERVER_VERSION`
- `diagnostics.DiagnosticsService` — the state of each dependency the caller
  names

## Installation

```bash
go get git.sonicoriginal.software/grpc-service
```

## What's Included

- **`service/`** — `Register`, which attaches the four services above alongside
  the caller's own
- **`diagnostics/`** — the diagnostics server and the checks it runs

## Usage

### Assembly

`service.Register` attaches the caller's services to the server, then the four
infrastructure services, marks each of the caller's services SERVING, and
returns their fully qualified method names.

It assembles only. The listener, the server, the background goroutines, and the
blocking `Serve` call are the caller's, because those are the pieces that differ
between production and a test.

`service/example_test.go` holds the whole sequence as a Go `Example`. It has no
`// Output:` comment, so `go test` compiles it and never runs it, which keeps it
type-checked against the real API. Read it there rather than from a copy here.

The returned method list excludes the `grpc.`, `info.`, and `diagnostics.`
services, which are infrastructure rather than something a caller advertises.

### Health Status

`health.NewServer()` marks the `""` entry SERVING, which answers "is this
process alive". `Register` adds an entry per service the caller registered,
under its fully qualified gRPC name, so a probe asks about
`yourpackage.YourService` rather than a logical name of your choosing.

Those entries start SERVING and stay there until you change them. Only your
application knows whether a given upstream being down means it can still do its
job, so deciding that is yours:

```go
healthSrv.SetServingStatus(
    yourpb.YourService_ServiceDesc.ServiceName,
    grpc_health_v1.HealthCheckResponse_NOT_SERVING,
)
```

The generated `_ServiceDesc.ServiceName` constant is the same name `Register`
used, so the two cannot drift.

### Diagnostics

`diagnostics.Checks` maps a dependency name to a `Check`. `Register` mounts a
diagnostics server over the map; `GetDiagnostics` runs every check concurrently
and reports each dependency's address, health, and connectivity state under its
name.

Three constructors cover the usual cases:

- `NewDependencyCheck(conn)` reports on a connection the caller already holds.
  Prefer this: what is reported is what is in use.
- `NewUpstreamCheck(conn, upstream)` is `NewDependencyCheck` for a connection
  whose target is a resolver URL rather than an address. `upstream` supplies the
  replica address the connection is on right now.
- `NewTargetCheck(address)` dials a fresh connection on every check, for a
  dependency the caller does not otherwise hold a connection to.

A `Check` is a plain function, so a dependency that is not a gRPC connection — a
database, a cache — is one the caller writes.

## Configuration

- `GRPC_SERVER_VERSION` — what the info service answers with. Read by
  `grpc-foundation`.
