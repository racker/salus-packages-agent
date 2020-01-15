Provides an agent that gathers the installed software packages reports to monitoring systems such as telegraf.

## Usage

```
  -configs string
    	directory containing config files that define continuous monitoring (env AGENT_CONFIGS)
  -debug
    	enables debug logging (env AGENT_DEBUG)
  -include-debian
    	enables debian package listing, when not using configs (env AGENT_INCLUDE_DEBIAN) (default true)
  -include-rpm
    	enables rpm package listing, when not using configs (env AGENT_INCLUDE_RPM) (default true)
  -line-protocol-to-console
    	indicates that line-protocol lines should be output to stdout (env AGENT_LINE_PROTOCOL_TO_CONSOLE)
  -line-protocol-to-socket host:port
    	the host:port of a telegraf TCP socket_listener (env AGENT_LINE_PROTOCOL_TO_SOCKET)
  -version
    	show version and exit
```

When invoked without configs the agent will collect packages, report in the configured manner, and exit.

When no specific reporter options are given, the collected package info is output in a human-readable format.

## Continuous-Monitoring Config File Format

When running the agent with the `--configs` option, it will periodically collect package telemetry at the interval configured in each config file. The option specifies a directory where any files in that directory that have a name ending with ".json" will be processed. The structure of those JSON files is:

```json
{
  "interval": "6h",
  "include-debian": true,
  "include-rpm": false,
  "fail-when-not-supported": true
}
```

where:
- `interval` : a Go duration specifying the interval of package collection. The default is "1h".
- `include-debian` : indicates if debian packages should be collected. The default is false.
- `include-rpm` : indicates if RPM packages should be collected. The default is false.
- `fail-when-not-supported` : when true, reports a "packages_failure" measurement when the requested package manager(s) is not supported on the system. The default is false.

## Influx Line Protocol Modes

This agent supports reporting package telemetry in the form of [InfluxDB line protocol](https://docs.influxdata.com/influxdb/v1.7/write_protocols/line_protocol_tutorial/).

### Console

When using `--line-protocol-to-console`, Influx line protocol metrics will be written to stdout with a "> " prefix, such as:

```
> packages,system=debian,package=sensible-utils,arch=all version="0.0.12" 1579042018775063900
> packages,system=debian,package=sysvinit-utils,arch=amd64 version="2.88dsf-59.10ubuntu1" 1579042018775063900
> packages,system=debian,package=tar,arch=amd64 version="1.29b-2ubuntu0.1" 1579042018775063900
> packages,system=debian,package=ubuntu-keyring,arch=all version="2018.09.18.1~18.04.0" 1579042018775063900
> packages,system=debian,package=util-linux,arch=amd64 version="2.31.1-0.4ubuntu3.4" 1579042018775063900
> packages,system=debian,package=zlib1g,arch=amd64 version="1:1.2.11.dfsg-0ubuntu2" 1579042018775063900
```

### Socket

When using `--line-protocol-to-scoket`, Influx line protocol metrics will be sent to a remote endpoint, such as [telegraf's socket_listener with `data_format="influx"`](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/socket_listener). The lines sent would look like:

```
packages,system=rpm,package=tzdata,arch=noarch version="2019a-1.el8" 1136214245000000000
packages,system=rpm,package=libselinux,arch=x86_64 version="2.8-6.el8" 1136214245000000000
``` 

## Running an example via Docker

Docker can be used to build and run the example even when you don't have one of the supported package managers (Debian, RPM) installed on your host system:

```
docker build -f Dockerfile.example .
```

If you have `make`, you can instead use

```
make example
```

The command above will 
- exercise the unit tests, 
- build an executable, -
- run the agent in a "one shot" manner reporting the packages installed (in the ubuntu container) in a human-readable format
- run the agent again using the line protocol format reporting to stdout