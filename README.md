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
