// Taken from github.com/blacked/go-zabbix (discontinued)
// Package implement zabbix sender protocol for send metrics to zabbix.
package zabbix

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"time"
)

const (
	DEFAULT_TIMEOUT = 15 * time.Second
)

// Metric class.
type Metric struct {
	Host   string `json:"host"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	Clock  int64  `json:"clock,omitempty"`
	Active bool   `json:"-"`
}

// Metric class constructor.
// agentActive should be set to true if we are sending to a Zabbix Agent (active) item
func NewMetric(host, key, value string, agentActive bool, clock ...int64) *Metric {
	m := &Metric{Host: host, Key: key, Value: value, Active: agentActive}
	// do not send clock if not defined
	if len(clock) > 0 {
		m.Clock = int64(clock[0])
	}
	return m
}

// Packet class.
type Packet struct {
	Request string    `json:"request"`
	Data    []*Metric `json:"data"`
	Clock   int64     `json:"clock,omitempty"`
}

// Packet class cunstructor.
func NewPacket(data []*Metric, agentActive bool, clock ...int64) *Packet {
	var request string
	if agentActive {
		request = "agent data"
	} else {
		request = "sender data"
	}

	p := &Packet{Request: request, Data: data}

	// do not send clock if not defined
	if len(clock) > 0 {
		p.Clock = int64(clock[0])
	}
	return p
}

// DataLen Packet class method, return 8 bytes with packet length in little endian order.
func (p *Packet) DataLen() []byte {
	dataLen := make([]byte, 8)
	JSONData, _ := json.Marshal(p)
	binary.LittleEndian.PutUint32(dataLen, uint32(len(JSONData)))
	return dataLen
}

// Sender class.
type Sender struct {
	Host    string
	Port    int
	Timeout time.Duration
}

// Sender class constructor.
// Optional timeout parameter
func NewSender(host string, port int, timeout_opt ...time.Duration) *Sender {
	var timeout time.Duration
	if len(timeout_opt) == 0 {
		// Default timeout value
		timeout = DEFAULT_TIMEOUT
	} else {
		timeout = timeout_opt[0]
	}
	s := &Sender{Host: host, Port: port, Timeout: timeout}
	return s
}

// Method Sender class, return zabbix header.
// https://www.zabbix.com/documentation/4.0/manual/appendix/protocols/header_datalen
func (s *Sender) getHeader() []byte {
	return []byte("ZBXD\x01")
}

// Method Sender class, resolve uri by name:port.
func (s *Sender) getTCPAddr() (iaddr *net.TCPAddr, err error) {
	// format: hostname:port
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	// Resolve hostname:port to ip:port
	iaddr, err = net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		err = fmt.Errorf("Connection failed: %s", err.Error())
		return
	}

	return
}

// Method Sender class, make connection to uri.
func (s *Sender) connect() (conn *net.TCPConn, err error) {

	type DialResp struct {
		Conn  *net.TCPConn
		Error error
	}

	// Open connection to zabbix host
	iaddr, err := s.getTCPAddr()
	if err != nil {
		return
	}

	// dial tcp and handle timeouts
	ch := make(chan DialResp)

	go func() {
		conn, err = net.DialTCP("tcp", nil, iaddr)
		ch <- DialResp{Conn: conn, Error: err}
	}()

	select {
	case <-time.After(s.Timeout):
		err = fmt.Errorf("connection timeout (%v)", s.Timeout)
	case resp := <-ch:
		if resp.Error != nil {
			err = resp.Error
			break
		}

		conn = resp.Conn
	}

	return
}

// Method Sender class, read data from connection.
func (s *Sender) read(conn *net.TCPConn) (res []byte, err error) {
	res = make([]byte, 1024)
	res, err = ioutil.ReadAll(conn)
	if err != nil {
		err = fmt.Errorf("Error whule receiving the data: %s", err.Error())
		return
	}

	return
}

// SendMetrics send an array of metrics, making different packets for
// trapper and active items.
// The response for trapper metrics is in the first element of the res array and err array
// Response for active metrics is in the second element of the res array and error array
func (s *Sender) SendMetrics(metrics []*Metric) (resActive []byte, errActive error, resTrapper []byte, errTrapper error) {
	var trapperMetrics []*Metric
	var activeMetrics []*Metric

	for i := 0; i < len(metrics); i++ {
		if metrics[i].Active {
			activeMetrics = append(activeMetrics, metrics[i])
		} else {
			trapperMetrics = append(trapperMetrics, metrics[i])
		}
	}

	if len(trapperMetrics) > 0 {
		packetTrapper := NewPacket(trapperMetrics, false)
		resTrapper, errTrapper = s.Send(packetTrapper)
	}

	if len(activeMetrics) > 0 {
		packetActive := NewPacket(activeMetrics, true)
		resActive, errActive = s.Send(packetActive)
	}

	return
}

// Method Sender class, send packet to zabbix.
func (s *Sender) Send(packet *Packet) (res []byte, err error) {
	conn, err := s.connect()
	if err != nil {
		return
	}
	defer conn.Close()

	dataPacket, _ := json.Marshal(packet)

	// Fill buffer
	buffer := append(s.getHeader(), packet.DataLen()...)
	buffer = append(buffer, dataPacket...)

	// Sent packet to zabbix
	_, err = conn.Write(buffer)
	if err != nil {
		err = fmt.Errorf("Error while sending the data: %s", err.Error())
		return
	}

	res, err = s.read(conn)

	return
}
