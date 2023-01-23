# manager

## Constants

```golang
const (
    SminitPidPath        = "/run/sminit/sminit.pid"
    SminitRunDir         = "/run/sminit"
    SminitLogPath        = "/run/sminit.log"
    SminitSocketPath     = "/run/sminit/sminit.sock"
    ServiceDefinitionDir = "/etc/sminit"
)
```

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

#### Constants

```golang
const (
    Running    Status = "running"
    Successful Status = "successful"
    Started    Status = "started"

    Pending Status = "pending"
    Failure Status = "failure"
    Stopped Status = "stopped"
)
```

### type [Watcher](/swatch.go#L19)

`type Watcher struct { ... }`

Watcher handles services through the manager, and accepts requests through Listener

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
