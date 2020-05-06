package monitoring

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/sirupsen/logrus"
)

const (
	visTable    = "vis"
	attTable    = "source"
	SentBytes   = "sentBytes"
	RcvdBytes   = "rcvdBytes"
	SentPackets = "sentPackets"
	RcvdPackets = "rcvdPackets"
	Conns       = "connections"
	QueuedPkts  = "queuedPkts"
)

var (
	monAddr      *net.UDPAddr
	defaultPoint point
	muConn       sync.Mutex
	connections  map[string]uint64
)

func Init(address string) (enable bool, err error) {
	if address == "" {
		logrus.Println("start without sending metrics to influx")
		return
	}
	monAddr, err = net.ResolveUDPAddr("udp", address)
	if err != nil {
		logrus.Errorf("error while connecting to influx: %s\n", err)
		return
	}
	host, err := getHost()
	if err != nil {
		logrus.Errorf("error while getting host IP: %s\n", err)
		return
	}
	defaultPoint = point{
		tags: map[string]string{
			"host":     host,
			"instance": util.Instance,
		},
		values: make(map[string]string),
	}
	connections = make(map[string]uint64)
	logrus.Infof("start sending metrics to influx to %+s:%+v, instance: %+s",
		monAddr.IP, monAddr.Port, util.Instance)
	return true, nil
}

func SendMetric(systemName string, metric string, count int) {
	newPoint := defaultPoint
	if "" != systemName {
		newPoint.tags["system"] = systemName
		newPoint.table = visTable
	} else {
		newPoint.table = attTable
	}
	newPoint.values[metric] = strconv.Itoa(count)
	record := newPoint.toRecord()
	send(record)
}

func NewConn(systemName string) {
	newPoint := defaultPoint
	if "" != systemName {
		newPoint.tags["system"] = systemName
		newPoint.table = visTable
	} else {
		newPoint.table = attTable
	}
	muConn.Lock()
	connections[systemName]++
	newPoint.values[Conns] = strconv.FormatUint(connections[systemName], 10)
	muConn.Unlock()
	record := newPoint.toRecord()
	send(record)
}

func CloseConn(systemName string) {
	newPoint := defaultPoint
	if "" != systemName {
		newPoint.tags["system"] = systemName
		newPoint.table = visTable
	} else {
		newPoint.table = attTable
	}
	muConn.Lock()
	if connections[systemName] > 0 {
		connections[systemName]--
	}
	newPoint.values[Conns] = strconv.FormatUint(connections[systemName], 10)
	muConn.Unlock()
	record := newPoint.toRecord()
	send(record)
}

func send(record string) error {
	logrus.Infof("send influx %s", record)
	conn, err := net.DialUDP("udp", nil, monAddr)
	if nil != err {
		return err
	}
	defer conn.Close()
	w := bufio.NewWriter(conn)
	_, err = fmt.Fprintf(w, record)
	if nil != err {
		return err
	}
	err = w.Flush()
	if nil != err {
		return err
	}
	return nil
}

func getHost() (string, error) {
	return "10_1_116_55", nil //TODO: динамическое вычисление
}
