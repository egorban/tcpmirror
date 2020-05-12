package monitoring

import (
	"math"
	"sync"

	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/egorban/influx/pkg/influx"
	"github.com/sirupsen/logrus"
)

const (
	TerminalName = "terminal"

	visTable = "vis"
	attTable = "source"

	SentBytes      = "sentBytes"
	RcvdBytes      = "rcvdBytes"
	SentPkts       = "sentPkts"
	RcvdPkts       = "rcvdPkts"
	QueuedPkts     = "queuedPkts"
	numConnections = "numConnections"
)

var (
	defaultTags  influx.Tags
	connsSystems map[string]uint64
	muConn       sync.Mutex
)

func Init(address string, systems []util.System) (monEnable bool, monClient *influx.Client, err error) {
	if address == "" {
		logrus.Println("start without sending metrics to influx")
		return
	}
	monClient, err = influx.NewClient(address, 30*1000, 5000)
	if err != nil {
		logrus.Println("error while connecting to influx", err)
		return
	}
	connsSystems = initSystemsConns(systems)
	defaultTags = setDefaultTags()
	monEnable = true
	return
}

func initSystemsConns(systems []util.System) map[string]uint64 {
	muConn.Lock()
	defer muConn.Unlock()

	connsSystems = make(map[string]uint64, len(systems)+1)

	connsSystems[TerminalName] = 0
	for _, sys := range systems {
		connsSystems[sys.Name] = 0
	}

	return connsSystems
}

func NewConn(monClient *influx.Client, systemName string) {
	muConn.Lock()
	numConn := connsSystems[systemName]
	if numConn < math.MaxUint64 {
		numConn++
		connsSystems[systemName] = numConn
	}
	muConn.Unlock()
	p := formPoint(systemName, numConnections, numConn)
	monClient.WritePoint(p)
}

func DelConn(monClient *influx.Client, systemName string) {
	muConn.Lock()
	numConn := connsSystems[systemName]
	if numConn > 0 {
		numConn--
		connsSystems[systemName] = numConn
	}
	muConn.Unlock()
	p := formPoint(systemName, numConnections, numConn)
	monClient.WritePoint(p)
}

func SendMetric(monClient *influx.Client, systemName string, metricName string, value interface{}) {
	logrus.Println("DEBUG Send Metric", systemName, metricName, value)
	p := formPoint(systemName, metricName, value)
	logrus.Println("DEBUG Send Metric Point", p)
	monClient.WritePoint(p)
}

func formPoint(systemName string, metricName string, value interface{}) *influx.Point {
	tags := defaultTags
	var table string
	if systemName == TerminalName {
		table = attTable
	} else {
		table = visTable
		tags["system"] = systemName
	}
	values := influx.Values{
		metricName: value,
	}
	p := influx.NewPoint(table, tags, values)
	return p
}

func setDefaultTags() influx.Tags {
	host := "10_1_116_55"
	return influx.Tags{
		"host":     host,
		"instance": util.Instance,
	}
}
