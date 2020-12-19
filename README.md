opcua_exporter
==============

This is a Prometheus exporter for the [OPC Unified Architecture](https://en.wikipedia.org/wiki/OPC_Unified_Architecture) protocol.

It uses [gopcua/opcua](https://github.com/gopcua/opcua) to communicate with an OPCUA endpoint, subscribes to 
selected channels, and republishes them as Promtheus metrics on a port of your choice.


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

Node Configuration
------------------
You need to supply a mapping of stringified OPC-UA node names to Prometheus metric names.
This is nessecary because the OPC-UA node strings use `;` and `=` characters to seperate
fields from each other and names from values, while Prometheus metric names must match 
the regex `[a-zA-Z_:][a-zA-Z0-9_:]*`. For good advice on Prometheus metric naming, refer 
to the [Prometheus docs](https://prometheus.io/docs/practices/naming/).

The exporter uses a YAML config file to specify the mapping.


An example config file might look like:
```yaml
- nodeName: ns=1;s=Voltmeter
  metricName: circuit_input_volts
- nodeName: ns=1;s=Ammeter
  metricName: circuit_input_amps
- nodeName: ns=1;s=CircuitBreakerStates
  extractBit: 3 # pull just this bit from a bit-vector channel
  metricName: circuit_breaker_three_tripped
```

Bit Vectors
-----------
Some of our OPC-UA devices send alarm states as binary bit-vector values,
for example, a 32-bit unsinged integer where each bit represents the state of a circuit breaker.

If you want to monitor a single bit, you can specify an `extractBit` in the node configuration file. 
The exporter will pull just that bit (zero-indexed) from the value of the OPC-UA channel, and export it
as a 0.0 or 1.0 Prometheus metric value.