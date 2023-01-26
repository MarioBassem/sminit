# manager

## Variables

```golang
var (
    ErrSminitInternalError = errors.New("sminit internal error")
    ErrBadRequest          = errors.New("bad request")
)
```

```golang
var (
    Address = "127.0.0.1"
    Port    = 8080
)
```

```golang
var (
    // SminitLog is the default logger used in sminit
    SminitLog = log.Output(zerolog.ConsoleWriter{
        Out: os.Stdout,
        FieldsExclude: []string{
            "component",
        },
        PartsOrder: []string{
            "level",
            "component",
            "message",
        },
    }).With().Str("component", "sminit:").Logger()
)
```

## Types

### type [App](/swatch.go#L15)

`type App struct { ... }`

App handles services through the manager, and accepts requests through Listener

### type [Manager](/manager.go#L24)

`type Manager struct { ... }`

Manager handles service manipulation

### type [Service](/manager.go#L50)

`type Service struct { ... }`

Service contains all needed information during the lifetime of a service

### type [ServiceOptions](/loader.go#L14)

`type ServiceOptions struct { ... }`

### type [Status](/manager.go#L37)

`type Status string`

Status presents service status

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
