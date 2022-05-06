mackerel-plugin-fireworq
========================

A [Fireworq][] custom metrics plugin for mackerel.io agent.

## Synopsis

```shell
mackerel-plugin-fireworq [-scheme=<'http'|'https'>] [-host=<host>] [-port=<manage_port>] [-tempfile=<tempfile>] [-metric-key-prefix=<prefix>] [-metric-label-prefix=<label-prefix>] [-queue-stats=<comma-separated-queue-stats>]
```

Queue stats can contain `pushes`, `pops`, `successes`, `failed`, `permanent_failed`, `completes`. The default is nothing. If you want to record all, use `-queue-stats='pushes,pops,successes,failures,permanent_failures,completes'`.

## Example of mackerel-agent.conf

```
[plugin.metrics.fireworq]
command = "/path/to/mackerel-plugin-fireworq"
```

[Fireworq]: https://github.com/fireworq/fireworq

## Installation
### Install with mkr
```bash
mkr plugin install --upgrade fireworq/mackerel-plugin-fireworq
```

### Build from source
```bash
go get github.com/fireworq/mackerel-check-fireworq
```

## How to release
See Makefile for details.
```
make release
```
