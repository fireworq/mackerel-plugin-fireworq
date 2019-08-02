mackerel-plugin-fireworq
========================

A [Fireworq][] custom metrics plugin for mackerel.io agent.

## Synopsis

```shell
mackerel-plugin-fireworq [-scheme=<'http'|'https'>] [-host=<host>] [-port=<manage_port>] [-tempfile=<tempfile>] [-metric-key-prefix=<prefix>] [-metric-label-prefix=<label-prefix>]
```

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
