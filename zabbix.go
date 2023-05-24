// Package implement zabbix sender protocol for send metrics to zabbix.
package zabbix

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

// Metric class.
type Metric struct {
	Host  string `json:"host"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Clock int64  `json:"clock"`
}

// Metric class constructor.
func NewMetric(host, key, value string, clock ...int64) *Metric {
	m := &Metric{Host: host, Key: key, Value: value}
	// use current time, if `clock` is not specified
	if m.Clock = time.Now().Unix(); len(clock) > 0 {
		m.Clock = int64(clock[0])
	}
	return m
}

// Packet class.
type Packet struct {
	Request string    `json:"request"`
	Data    []*Metric `json:"data"`
	Clock   int64     `json:"clock"`
}

// Packet class cunstructor.
func NewPacket(data []*Metric, clock ...int64) *Packet {
	p := &Packet{Request: `sender data`, Data: data}
	// use current time, if `clock` is not specified
	if p.Clock = time.Now().Unix(); len(clock) > 0 {
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
	Host string
	Port int
}

// Sender class constructor.
func NewSender(host string, port int) *Sender {
	s := &Sender{Host: host, Port: port}
	return s
}

// Method Sender class, return zabbix header.
func (s *Sender) getHeader() []byte {
	return []byte("ZBXD\x01")
}

// Method Sender class, make connection to uri.
func (s *Sender) connect() (tcpConn *net.TCPConn, err error) {
	// parses addr as "hostname:port"
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	// resolves hostname, dials TCP and handles timeout
	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		return
	}
	tcpConn, _ = conn.(*net.TCPConn)
	return
}

// Method Sender class, read data from connection.
func (s *Sender) read(conn *net.TCPConn) (res []byte, err error) {
	res, err = io.ReadAll(conn)
	if err != nil {
		err = fmt.Errorf("error whule receiving the data: %s", err.Error())
		return
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

	/*
	   fmt.Printf("HEADER: % x (%s)\n", s.getHeader(), s.getHeader())
	   fmt.Printf("DATALEN: % x, %d byte\n", packet.DataLen(), len(packet.DataLen()))
	   fmt.Printf("BODY: %s\n", string(dataPacket))
	*/

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

	/*
	   fmt.Printf("RESPONSE: %s\n", string(res))
	*/
	return
}
