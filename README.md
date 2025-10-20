# Medusa Exporter for Apache Cassandra

[![Actions Status](https://github.com/woblerr/medusa_exporter/workflows/build/badge.svg)](https://github.com/woblerr/medusa_exporter/actions)
[![Coverage Status](https://coveralls.io/repos/github/woblerr/medusa_exporter/badge.svg?branch=master)](https://coveralls.io/github/woblerr/medusa_exporter?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/woblerr/medusa_exporter)](https://goreportcard.com/report/github.com/woblerr/medusa_exporter)

Prometheus exporter for [cassandra-medusa](https://github.com/thelastpickle/cassandra-medusa).

The metrics are collected based on result of `medusa list-backups --output json` command. You need to run exporter on the same host where Medusa was installed or inside Docker.

## Grafana dashboard

To get a dashboard for visualizing the collected metrics, you can use a ready-made dashboard [DRAFT Medusa Exporter for Apache Cassandra Dashboard]() or make your own.

## Collected metrics
### Backup metrics

| Metric | Description |  Labels | Additional Info |
| ----------- | ------------------ | ------------- | --------------- |
| `medusa_backup_info` | backup info | backup_name, backup_type, prefix, start_time | Values description:<br> `1` - info about backup is exist. |
| `medusa_backup_status` | backup status | backup_name, backup_type | Values description:<br> `0` - backup is not complete,<br> `1` - backup is complete. |
| `medusa_backup_duration_seconds` | backup duration in seconds | backup_name, backup_type, start_time, stop_time | |
| `medusa_backup_size_bytes` | backup size in bytes | backup_name, backup_type | |
| `medusa_backup_objects` | number of objects in backup | backup_name, backup_type | |
| `medusa_backup_completed_nodes` | number of completed nodes in backup | backup_name, backup_type | |
| `medusa_backup_incomplete_nodes` | number of incomplete nodes in backup | backup_name, backup_type | |
| `medusa_backup_missing_nodes` | number of missing nodes in backup | backup_name, backup_type | |
| `medusa_node_backup_info` | node backup info | backup_name, backup_type, node_fqdn, prefix, release_version, server_type, start_time | Values description:<br> `1` - info about node backup is exist. |
| `medusa_node_backup_status` | node backup status | backup_name, backup_type, node_fqdn | Values description:<br> `0` - node backup is not complete,<br> `1` - node backup is complete. |
| `medusa_node_backup_duration_seconds` | node backup duration in seconds | backup_name, backup_type, node_fqdn, start_time, stop_time | |
| `medusa_node_backup_size` | node backup size in bytes | backup_name, backup_type, node_fqdn | |
| `medusa_node_backup_objects` | number of objects in node backup | backup_name, backup_type, node_fqdn | |

### Exporter metrics

| Metric | Description |  Labels | Additional Info |
| ----------- | ------------------ | ------------- | --------------- |
| `medusa_exporter_build_info` | information about Medusa exporter | branch, goarch, goos, goversion, revision, tags, version | |
| `medusa_exporter_status` | Medusa exporter get data status | prefix | Values description:<br> `0` - errors occurred when fetching information from Medusa,<br> `1` - information successfully fetched from Medusa. |

### Additional description of metrics

For `medusa_backup_duration_seconds` and `medusa_node_backup_duration_seconds` metrics the following logic is applied:
* if backup/node backup is complete then value calculated;
* if backup/node backup is not complete, then value is `0`, labels `stop_time` is `none`.


## Getting Started
### Building and running

```bash
git clone https://github.com/woblerr/medusa_exporter.git
cd medusa_exporter
make build
./medusa_exporter <flags>
```

Available configuration flags:

```bash
usage: medusa_exporter [<flags>]


Flags:
  -h, --[no-]help              Show context-sensitive help (also try --help-long and --help-man).
      --web.telemetry-path="/metrics"  
                               Path under which to expose metrics.
      --web.listen-address=:19500 ...  
                               Addresses on which to expose metrics and web interface. Repeatable for multiple
                               addresses. Examples: `:9100` or `[::1]:9100` for http, `vsock://:9100` for vsock
      --web.config.file=""     Path to configuration file that can enable TLS or authentication. See:
                               https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md
      --collect.interval=600   Collecting metrics interval in seconds.
      --medusa.config-file=""  Full path to Medusa configuration file.
      --medusa.prefix=""       Prefix for shared storage.
      --log.level=info         Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt      Output format of log messages. One of: [logfmt, json]
      --[no-]version           Show application version.
```

#### Additional description of flags

Custom `config` for `medusa` command can be specified via `--medusa.config` flag. Full paths must be specified.<br>
For example, `--medusa.config=/tmp/medusa.conf`.

When `--log.level=debug` is specified - information of values and labels for metrics is printing to the log.

The flag `--web.config.file` allows to specify the path to the configuration for TLS and/or basic authentication.<br>
The description of TLS configuration and basic authentication can be found at [exporter-toolkit/web](https://github.com/prometheus/exporter-toolkit/blob/v0.14.1/docs/web-configuration.md).


### Running as systemd service

* Register `medusa_exporter` (already builded, if not - exec `make build` before) as a systemd service:

```bash
make prepare-service
```

Validate prepared file `medusa_exporter.service` and run:

```bash
sudo make install-service
```

* View service logs:

```bash
journalctl -u medusa_exporter.service
```

* Delete systemd service:

```bash
sudo make remove-service
```

---
Manual register systemd service:

```bash
cp medusa_exporter.service.template medusa_exporter.service
```

In file `medusa_exporter.service` replace `{PATH_TO_FILE}` to full path to `medusa_exporter`.

```bash
sudo cp medusa_exporter.service /etc/systemd/system/medusa_exporter.service
sudo systemctl daemon-reload
sudo systemctl enable medusa_exporter.service
sudo systemctl restart medusa_exporter.service
systemctl -l status medusa_exporter.service
```

### RPM/DEB packages

You can use the already prepared rpm/deb package to install the exporter. Only the medusa_exporter binary  and the service file are installed by package.

For example:
```bash
rpm -ql medusa_exporter

/etc/systemd/system/medusa_exporter.service
/usr/bin/medusa_exporter
```