# Example boilerplate code for a fully featured connector (spooling, rendering, lookups)

connectorX is a sample boilerplate connector. 

Delivers job notification (rendered template) over a TCP connection to the specified address:port.
It is an example of how to use all of the 'extra' capabilities: lookups, rendertofile and spooling.

Files you'll need to get started:

* connector data structure [connector_data.go](./connector_data.go)
* connector code [connectorX.go](./connectorX.go)
* example [config file](./goslmailer.conf)
* example [template file](./conX.tmpl)

## Exercise for the reader:

To make this connector work, add the missing code block to the [connectors package](../../internal/connectors/connectors.go).
Recompile and try it out.

To verify it works:

```
[pja@red0 goslmailer]$ nc -lkv localhost 9999                                                                                                                                                                     
Ncat: Version 7.70 ( https://nmap.org/ncat )
Ncat: Listening on ::1:9999
Ncat: Connection from ::1.
Ncat: Connection from ::1:32848.
Job Name         : SendAllArrayJob
Job ID           : 1051492
User             : petar.jager
Partition        : c
Nodes Used       : stg-c2-0
Cores            : 4
Job state        : COMPLETED
Exit Code        : 0
Submit           : 2022-02-16T20:40:15
Start            : 2022-02-16T20:40:15
End              : 2022-02-17T01:11:04
Res. Walltime    : 08:00:00
Used Walltime    : 00:00:30
Used CPU time    : 01:57.511
% User (Comp)    : 86.81%
% System (I/O)   : 13.19%
Memory Requested : 34 GB
Max Memory Used  : 1.1 GB
Max Disk Write   : 10 kB
Max Disk Read    : 136 kB
TIP: Please consider lowering the ammount of requested memory in the future, your job has consumed less then half of the requested memory.                                                                        
TIP: Please consider lowering the amount of requested CPU cores in the future, your job has consumed less than half of requested CPU cores                                                                        
^C
```

**Good Luck!**