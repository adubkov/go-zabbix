package zabbix

import (
	"fmt"
	"testing"
)

func TestSendMetric(t *testing.T) {
	m := NewMetric("zabbixAgent1", "active", "13", true)
	m2 := NewMetric("zabbixAgent1", "trapper", "12", false)
	s := NewSender("172.17.0.3", 10051)
	res, err := s.SendMetrics([]*Metric{m, m2})
	if err[0] != nil || err[1] != nil {
		t.Fatalf("erroe enviando la metrica: %v", err)
	}
	fmt.Println("trapper:", string(res[0]))
	fmt.Println("active:", string(res[1]))
}
