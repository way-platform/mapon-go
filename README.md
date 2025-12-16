# Mapon Go

[![PkgGoDev](https://pkg.go.dev/badge/github.com/way-platform/mapon-go)](https://pkg.go.dev/github.com/way-platform/mapon-go)
[![GoReportCard](https://goreportcard.com/badge/github.com/way-platform/mapon-go)](https://goreportcard.com/report/github.com/way-platform/mapon-go)
[![CI](https://github.com/way-platform/mapon-go/actions/workflows/release.yaml/badge.svg)](https://github.com/way-platform/mapon-go/actions/workflows/release.yaml)

A Go SDK and CLI tool for the [Mapon API](https://mapon.com/api).

## SDK

### Features

- Read APIs for most core entities (Units, Routes, Objects, Drivers, Unit Groups, Alerts)
- Read APIs for unit data (CAN, ignition, humidity, etc)

### Installing

```bash
$ go get github.com/way-platform/mapon-go
```

### Using

```go
ctx := context.Background()
// Create a Mapon API client.
client, err := mapon.NewClient(
    ctx,
    mapon.WithAPIKey(os.Getenv("MAPON_API_KEY")),
)
if err != nil {
    panic(err)
}
// List vehicles in the account.
response, err := client.ListUnits(ctx, &mapon.ListUnitsRequest{})
if err != nil {
    panic(err)
}
for _, unit := range response.Units {
    fmt.Println(unit.GetUnitId())
}
// For all available methods, see the API documentation.
```

### Developing

#### Building

The project is built using [Mage](https://magefile.org), see
[tools/magefile.go](./tools/magefile.go) and the [tools/mage](./tools/mage)
helper script.

```bash
$ ./tools/mage build
```

For all available build tasks, see:

```bash
$ ./tools/mage -l
```

## CLI tool

The `mapon` CLI tool enables interaction with the APIs from the command line.

```bash
$ mapon

  Mapon API CLI

  USAGE

    mapon [command] [--flags]

  UNITS

    units [--flags]                                  List units

  UNIT DATA

    can-periods <unit-id> [--flags]                  List CAN data for a period
    can-point <unit-id> [--flags]                    Get CAN data at a specific time
    debug-info <unit-id ...>                         Get unit debug info
    digital-inputs <unit-id ...> [--flags]           List digital input events
    digital-inputs-extended <unit-id ...> [--flags]  List extended digital input events
    driving-time <unit-id>                           Get driving time extended
    fields <unit-id>                                 Get unit custom fields
    history-point <unit-id> [--flags]                Get historical data at a specific time
    humidity <unit-id ...> [--flags]                 List humidity data
    ibuttons <unit-id ...> [--flags]                 List iButton events
    ignitions <unit-id ...> [--flags]                List ignition events
    temperatures <unit-id ...> [--flags]             List temperature data

  UNIT GROUPS

    unit-groups list [--flags]                       List unit groups
    unit-groups units [--flags]                      List units in a group

  DRIVERS

    drivers [--flags]                                List drivers

  ROUTES

    routes [--flags]                                 List routes

  OBJECTS

    objects                                          List objects

  ALERTS

    alerts [--flags]                                 List alerts
```

### Installing

```bash
go install github.com/way-platform/mapon-go/cmd/mapon@latest
```

Prebuilt binaries for Linux, Windows, and Mac are available from the [Releases](https://github.com/way-platform/mapon-go/releases).

## License

This SDK is published under the [MIT License](./LICENSE).

## Security

Security researchers, see the [Security Policy](https://github.com/way-platform/mapon-go?tab=security-ov-file#readme).

## Code of Conduct

Be nice. For more info, see the [Code of Conduct](https://github.com/way-platform/mapon-go?tab=coc-ov-file#readme).
