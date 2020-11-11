# SINGLE BACKUPER
Creates backups, saves them to a local machine and/or remote SSH server. Backups of the following types are supported:
- MongoDB
- MySQL 
- Clickhouse
- Files

## BUILD 
Requires [Go](https://golang.org/doc/install). Tested with Go 1.15.

Clone this repo locally and run test, build:
```
mkdir -p $HOME/singlebackuper && \
cd $HOME/singlebackuper && \
git clone https://github.com/denfm/singlebackuper ./ && \
make test && make build && \
cd bin && ls -la
```

## Running

```
# module=mongodb|mysqldb|files|clickhouse
./singlebackuper --module=mongodb --config-path=/etc/singlebackuper/singlebackuper.toml
```

## PLAN Release 1.0
- feature/rotation
- feature/api (systemd service)

## PLAN Release 2.0
- feature/prometheus_metrics_exporter

## PLAN Release 3.0
- feature/gui

LICENSE
========

See [LICENSE](./LICENSE)