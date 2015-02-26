go-zabbix
==============================================================================
Golang package, implement zabbix sender protocol for send metrics to zabbix.

Example:
```go
package main

import (
    "time"
    . "github.com/blacked/go-zabbix"
)

const (
    defaultHost  = `localhost`
    defaultPort  = 10051
)

func main() {
    var metrics []*Metric
    metrics = append(metrics, NewMetric("localhost", "cpu", "1.22", time.Now().Unix()))
    metrics = append(metrics, NewMetric("localhost", "status", "OK"))

    // Create instance of Packet class
    packet := NewPacket(metrics)

    // Send packet to zabbix
    z := NewSender(defaultHost, defaultPort)
    z.Send(packet)
}
```
