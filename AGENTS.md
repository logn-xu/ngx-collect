# Project Guidelines

## Overview

Go CLI tool that collects Nginx config files from multiple remote servers via SSH/SFTP. Uses Cobra for CLI, Viper for config, and logrus for logging.

## Architecture

```
cmd/              # CLI entry point and command definitions (Cobra)
  ngx-collect/    # main.go
  root.go         # root command with flags, config init, concurrency control
internal/
  config/         # YAML config structs (Viper + mapstructure tags)
  collector/      # SFTP file download logic (Collecter interface)
  sshclient/      # SSH connection setup (key + password auth)
config/           # Example YAML config files
```

- `collector.Collecter` is the core interface: `Connect()`, `Fetch()`, `Close()`
- Concurrency is managed via `sync.WaitGroup` + `golang.org/x/sync/semaphore` with configurable batch size
- Each machine runs in its own goroutine with a `context.WithTimeout`

## Build and Test

```bash
make build        # builds ./ngx-collect binary
make clean        # removes data/*
./ngx-collect -c config/simple.config.yaml   # run with custom config
./ngx-collect -b 10 -t 30                    # override batch-size and timeout
```

## Conventions

- Config structs use `mapstructure` tags (not `yaml`), matching Viper's unmarshaling
- Logging: use `log.Errorf` / `log.Infof` from `github.com/sirupsen/logrus`; no `fmt.Print` in library code
- Error handling: log at the point of failure and return the error; caller decides whether to continue
- File paths: downloaded files are organized as `<local_destination>/<alias>/<host>/...`
- Interface naming: `Collecter` (project convention, not `Collector`)
