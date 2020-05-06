package monitoring

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"strconv"
	"sync"

	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/sirupsen/logrus"
)

const (
	AttMonName = "att"

	visTable   = "vis"
	attTable   = "source"
	sentBytes  = "sentBytes"
	rcvdBytes  = "rcvdBytes"
	sentPkts   = "sentPkts"
	rcvdPkts   = "rcvdPkts"
	conns      = "connections"
	queuedPkts = "queuedPkts"
)

var (
	monAddr      *net.UDPAddr
	defaultPoint point
	muConn       sync.Mutex
	connsSystems map[string]uint64
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
	connsSystems = initSystemsConns(systems)
	logrus.Infof("start sending metrics to influx to %+s:%+v, instance: %+s",
		monAddr.IP, monAddr.Port, util.Instance)
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

func SentBytes(systemName string, count int) {
	sendMetric(systemName, sentBytes, strconv.Itoa(count))
}

func RcvdBytes(systemName string, count int) {
	sendMetric(systemName, rcvdBytes, strconv.Itoa(count))
}

func SentPkts(systemName string, count int) {
	sendMetric(systemName, sentPkts, strconv.Itoa(count))
}

func RcvdPkts(systemName string, count int) {
	sendMetric(systemName, rcvdPkts, strconv.Itoa(count))
}

func QueuedPkts(systemName string, count int) {
	sendMetric(systemName, queuedPkts, strconv.Itoa(count))
}

func NewConn(systemName string) {
	muConn.Lock()
	count := connsSystems[systemName]
	if count < math.MaxUint64 {
		count++
	}
	connsSystems[systemName] = count
	muConn.Unlock()

	sendMetric(systemName, conns, strconv.FormatUint(count, 10))
}

func DeleteConn(systemName string) {
	muConn.Lock()
	count := connsSystems[systemName]
	if count > 0 {
		count--
	}
	connsSystems[systemName] = count
	muConn.Unlock()

	sendMetric(systemName, conns, strconv.FormatUint(count, 10))
}

func sendMetric(systemName string, metric string, count string) {
	newPoint := defaultPoint
	if systemName != AttMonName {
		newPoint.tags["system"] = systemName
		newPoint.table = visTable
	} else {
		newPoint.table = attTable
	}
	newPoint.values[metric] = count
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
