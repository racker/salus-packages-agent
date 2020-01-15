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
make example
```

The command above will 
- exercise the unit tests, 
- build a local image base on Ubuntu with the executable,
- run the agent in a "one shot" manner reporting the packages installed (in the ubuntu container) in a human-readable format
- run the agent again using the line protocol format reporting to stdout

## Running an integration test with telegraf

A Docker compose file is provided to run a build and integration test with a telegraf container using a socket_listener input plugin. The following command will start the test and Control-C to stop the containers.

```
docker-compose up
```

The configuration file used is located in [testdata/compose-configs](testdata/compose-configs) and instructs the agent to re-collect every 1 minute; however, you will see output such as the following shortly after startup since the interval kicks in after the first collection.

```
telegraf_1  | 2020-01-15T17:45:53Z I! Starting Telegraf 1.13.1
telegraf_1  | 2020-01-15T17:45:53Z I! Using config file: /etc/telegraf/telegraf.conf
telegraf_1  | 2020-01-15T17:45:53Z I! Loaded inputs: socket_listener
telegraf_1  | 2020-01-15T17:45:53Z I! Loaded aggregators:
telegraf_1  | 2020-01-15T17:45:53Z I! Loaded processors:
telegraf_1  | 2020-01-15T17:45:53Z I! Loaded outputs: file
telegraf_1  | 2020-01-15T17:45:53Z I! Tags enabled: host=40fa37e3c964
telegraf_1  | 2020-01-15T17:45:53Z I! [agent] Config: Interval:10s, Quiet:false, Hostname:"40fa37e3c964", Flush Interval:10s
telegraf_1  | 2020-01-15T17:45:53Z I! [inputs.socket_listener] Listening on tcp://[::]:8094
sut_1       | 2020-01-15T17:45:55.493Z	DEBUG	build/lister.go:69	calling packaging tool	{"name": "dpkg-query", "args": ["--show", "--showformat", "${Package} ${Version} ${Architecture}\\n"]}
telegraf_1  | packages,arch=amd64,host=40fa37e3c964,package=zlib1g,system=debian version="1:1.2.11.dfsg-0ubuntu2" 1579110355492834600
telegraf_1  | packages,arch=amd64,host=40fa37e3c964,package=util-linux,system=debian version="2.31.1-0.4ubuntu3.4" 1579110355492834600
telegraf_1  | packages,arch=all,host=40fa37e3c964,package=ubuntu-keyring,system=debian version="2018.09.18.1~18.04.0" 1579110355492834600
telegraf_1  | packages,arch=amd64,host=40fa37e3c964,package=tar,system=debian version="1.29b-2ubuntu0.1" 1579110355492834600
telegraf_1  | packages,arch=amd64,host=40fa37e3c964,package=sysvinit-utils,system=debian version="2.88dsf-59.10ubuntu1" 1579110355492834600
```