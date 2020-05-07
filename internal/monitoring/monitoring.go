package monitoring

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/sirupsen/logrus"
)

const (
	AttMonName = "att"

	SentBytes  = "sentBytes"
	RcvdBytes  = "rcvdBytes"
	SentPkts   = "sentPkts"
	RcvdPkts   = "rcvdPkts"
	conns      = "connections"
	QueuedPkts = "queuedPkts"

	visTable = "vis"
	attTable = "source"
)

var (
	monAddr      *net.UDPAddr
	defaultPoint point
	muConn       sync.Mutex
	connsSystems map[string]uint64
	host         string
)

func Init(address string, systems []util.System) (enable bool, err error) {
	if address == "" {
		logrus.Println("start without sending metrics to influx")
		return
	}
	monAddr, err = net.ResolveUDPAddr("udp", address)
	if err != nil {
		logrus.Errorf("error while connecting to influx: %s\n", err)
		return
	}
	host, err = getHost()
	if err != nil {
		logrus.Errorf("error while getting host IP: %s\n", err)
		return
	}
	connsSystems = initSystemsConns(systems)
	logrus.Infof("start sending metrics to influx to %+s:%+v, instance: %+s",
		monAddr.IP, monAddr.Port, util.Instance)
	go periodicMon()
	return true, nil
}

func initSystemsConns(systems []util.System) map[string]uint64 {
	muConn.Lock()
	defer muConn.Unlock()

	connsSystems = make(map[string]uint64, len(systems)+1)

	connsSystems[AttMonName] = 0
	for _, sys := range systems {
		connsSystems[sys.Name] = 0
	}

	return connsSystems
}

func periodicMon() {
	for {
		muConn.Lock()
		for name, count := range connsSystems {
			SendMetric(name, conns, strconv.FormatUint(count, 10))
		}
		muConn.Unlock()
		time.Sleep(30 * time.Second)
	}
}

func NewConn(systemName string) {
	muConn.Lock()
	count := connsSystems[systemName]
	if count < math.MaxUint64 {
		count++
	}
	connsSystems[systemName] = count
	muConn.Unlock()

	SendMetric(systemName, conns, strconv.FormatUint(count, 10))
}

func DeleteConn(systemName string) {
	muConn.Lock()
	count := connsSystems[systemName]
	if count > 0 {
		count--
	}
	connsSystems[systemName] = count
	muConn.Unlock()

	SendMetric(systemName, conns, strconv.FormatUint(count, 10))
}

func SendMetric(systemName string, metric string, count string) {
	var newPoint = point{
		tags: map[string]string{
			"host":     host,
			"instance": util.Instance,
		},
		values: map[string]string{
			metric: count,
		},
	}
	if systemName != AttMonName {
		newPoint.table = visTable
		newPoint.tags["system"] = systemName
	} else {
		newPoint.table = attTable
	}
	logrus.Infof("newPoint %v", newPoint)
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
