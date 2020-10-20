go-zabbix
==============================================================================
Golang package, implement zabbix sender protocol for send metrics to zabbix.

Example:
```go
package main

import (
	"fmt"
	"time"

	"github.com/datadope-io/go-zabbix"
)

const (
	defaultHost  = `localhost`
	defaultPort = 10051
	agentActive = true
	trapper     = false
)

func main() {
	var metrics []*zabbix.Metric
	metrics = append(metrics, zabbix.NewMetric("localhost", "cpu", "1.22", agentActive, time.Now().Unix()))
	metrics = append(metrics, zabbix.NewMetric("localhost", "status", "OK", agentActive))
	metrics = append(metrics, zabbix.NewMetric("localhost", "someTrapper", "3.14", trapper))

	// Send metrics to zabbix
	z := zabbix.NewSender(defaultHost, defaultPort)
	resActive, errActive, resTrapper, errTrapper := z.SendMetrics(metrics)

	fmt.Printf("Agent active, response=%s, info=%s, error=%v\n", resActive.Response,resActive.Info, errActive)
	fmt.Printf("Trapper, response=%s, info=%s,error=%v\n", resTrapper.Response, resTrapper.Info,errTrapper)
}
```

# CHANGELOG

# 2020-10-20

* Now Active/Trapper response is formated in Response,Info strings and could be translated int ResponseInfo struct with Failed/Completed/Processed/Spent time.