# Debugging flnd

`flnd` has several built-in features that make it easy to debug the system.
Specifically, it has several logging subsystems that can be controlled
independently.

## Logging

`flnd` uses several logging subsystems. Each subsystem can be controlled
independently by using the `debuglevel` flag.

The general format for the `debuglevel` flag is:
```shell
--debuglevel=<subsystem>=<level>,<subsystem2>=<level>,...
```

The available logging levels are: `trace`, `debug`, `info`, `warn`, `error`,
`critical`.

To list all available subsystems, you can run:
```shell
lncli debuglevel --show
```

### Logging to stdout

By default, `flnd` will log to a file. You can also tell `flnd` to log to stdout
by using the `--logconsole` flag.

### Changing the log level at runtime

You can change the log level of a subsystem at runtime by using the `lncli`
command:
```shell
lncli debuglevel --level=<subsystem>=<level>
```

### How are the logger prefixes created?

The logger prefixes are defined in the `log.go` file of the `flnd` package.
Each subsystem is registered using the `AddSubLogger` function:

```go
 AddSubLogger(root, "HSWC", interceptor, htlcswitch.UseLogger)
```

Caution: Some logger subsystems are overwritten during the instanziation. An
example here is the `neutrino/query` package which instead of using the `FLCN`
prefix is overwritten by the `LNWL` subsystem.

Moreover when using the `lncli` command the return value will provide the 
updated list of all subsystems and their associated logging levels. This makes
it easy to get an overview of the current logging level for the whole system.

Example:

```shell
{
    "sub_systems": "ARPC=INF, ATPL=INF, BLPT=INF, BRAR=INF, FLCN=INF, FLWL=INF, CHAC=INF, CHBU=INF, CHCL=INF, CHDB=INF, CHFD=INF, CHFT=INF, CHNF=INF, CHRE=INF, CLUS=INF, CMGR=INF, CNCT=INF, CNFG=INF, CRTR=INF, DISC=INF, DRPC=INF, FNDG=INF, GRPH=INF, HLCK=INF, HSWC=DBG, INVC=INF, IRPC=INF, LNWL=INF, LTND=INF, NANN=INF, NRPC=INF, NTFN=INF, NTFR=INF, PEER=INF, PRNF=INF, PROM=INF, PRPC=INF, RPCP=INF, RPCS=INF, RPWL=INF, RRPC=INF, SGNR=INF, SPHX=INF, SRVR=INF, SWPR=INF, TORC=INF, UTXN=INF, VRPC=INF, WLKT=INF, WTCL=INF, WTWR=INF"
}
```


## Built-in profiler in LND

`LND` has a built-in feature which allows you to capture profiling data at
runtime. This is very useful when you want to investigate performance issues.

The profiler is controlled by the `pprof` sub-command:

```shell
--pprof.cpuprofile=<file>
--pprof.profile=<port>
--pprof.blockingprofile=<file>
```

For more information, please see the [official pprof documentation](https://github.com/google/pprof/blob/master/doc/README.md).
