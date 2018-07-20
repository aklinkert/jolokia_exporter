# jolokia_exporter for prometheus

This is an exporter for prometheus, written as a [cobra](https://github.com/spf13/cobra) application to export the data from [Jolokia](jolokia.org/reference/html/protocol.html).

# docker image

There is an automatically build docker image out there: [scalify/jolokia_exporter](https://hub.docker.com/r/scalify/jolokia_exporter/)

# usage

The exporter is configured using command line flags and arguments. Usage is as follows:

```
Exports jolokia metrics from given endpoint, using given metrics mapping config

Usage:
  jolokia_exporter export <metrics-config-file> <endpoint> [flags]

Flags:
      --basic-auth-password string   HTTP Basic auth password for authentication on the jolokia endpoint
      --basic-auth-user string       HTTP Basic auth user for authentication on the jolokia endpoint
  -e, --endpoint string              Path the exporter should listen listen on (default "/metrics")
  -h, --help                         help for export
  -i, --insecure                     Whether to use insecure https mode, i.e. skip ssl cert validation (only useful with https endpoint)
  -l, --listen string                Host/Port the exporter should listen listen on (default ":9422")
  -v, --verbose                      Whether to use verbose https mode
```

Example usage in a docker-compose file:

```yaml
version: "3"
services:
  fixtures:
    build:
      context: fixtures
    image: scalify/jolokia_exporter_test_server
    ports:
      - "3000:3000"
  exporter:
    build:
      context: .
    image: scalify/jolokia_exporter
    volumes:
      - "./fixtures:/fixtures"
    command: [
      "export", "/fixtures/config.yaml", "http://fixtures:3000/manage/jolokia",
      "--basic-auth-user", "admin",
      "--basic-auth-password", "secret",
      "--insecure"
    ]
    links:
      - fixtures
    ports:
      - "9422:9422"
```

The exporter loads a configuration file, containing information about how to resolve mbean data:

```yaml
metrics:
- source:
    mbean: java.lang:type=Memory
    attribute: HeapMemoryUsage
    path: used
  target: java_memory_heap_memory_usage_used
- source:
    mbean: java.lang:type=Memory
    attribute: HeapMemoryUsage
    path: max
  target: java_memory_max
- source:
    mbean: java.lang:type=Threading
    attribute: ThreadCount
  target: java_threading_thread_count
- source:
    mbean: java.lang:type=OperatingSystem
  target: java_os
```

More information on how to specify mbeans can be found in the [Jolokia docs](https://jolokia.org/reference/html/protocol.html#post-request). For a complete example have a look into the `fixtures` directory and the `docker-compose.yml`

# license

MIT License

Copyright (c) 2018 Scalify GmbH
