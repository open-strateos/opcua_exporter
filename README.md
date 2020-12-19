opcua_exporter
==============

This is a Prometheus exporter for the [OPC Unified Architecture](https://en.wikipedia.org/wiki/OPC_Unified_Architecture) protocol.


Usage
-----
```
Usage of opcua_exporter:
  -buffer-size int
    	Maximum number of messages in the receive buffer (default 64)
  -config string
    	Path to a file from which to read the list of OPC UA nodes to monitor
  -config-b64 string
    	Base64-encoded config JSON. Overrides -config
  -debug
    	Enable debug logging
  -endpoint string
    	OPC UA Endpoint to connect to. (default "opc.tcp://localhost:4096")
  -max-timeouts int
    	The exporter will quit trying after this many read timeouts (0 to disable).
  -port int
    	Port to publish metrics on. (default 9686)
  -prom-prefix string
    	Prefix will be appended to emitted prometheus metrics
  -read-timeout duration
    	Timeout when waiting for OPCUA subscription messages (default 5s)
  -summary-interval duration
    	How frequently to print an event count summary (default 5m0s)

```