package monitoring

import (
	"log"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ashirko/tcpmirror/internal/db"
	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/egorban/influx/pkg/influx"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

const (
	TerminalName = "terminal"

	visTable   = "vis"
	attTable   = "source"
	redisTable = "redis"

	SentBytes       = "sentBytes"
	RcvdBytes       = "rcvdBytes"
	SentPkts        = "sentPkts"
	RcvdPkts        = "rcvdPkts"
	QueuedPkts      = "queuedPkts"
	numConnections  = "numConnections"
	unconfirmedPkts = "unconfirmedPkts"
)

var (
	host         string
	connsSystems map[string]uint64
	muConn       sync.Mutex
	connsRedis   uint64
	terminals    map[int]uint64
	muRedis      sync.Mutex
)

func Init(address string, systems []util.System, dbAddress string) (monEnable bool, monClient *influx.Client, err error) {
	if address == "" {
		logrus.Println("start without sending metrics to influx")
		return
	}
	monClient, err = influx.NewClient(address)
	if err != nil {
		logrus.Println("error while connecting to influx", err)
		return
	}
	connsSystems = initSystemsConns(systems)
	monEnable = true
	host = getHost()
	go periodicMon(monClient, dbAddress)
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

func periodicMon(monСlient *influx.Client, dbAddress string) {
	conn := db.Connect(dbAddress)
	defer conn.Close()
	for {
		muConn.Lock()
		for systemName, numConn := range connsSystems {
			log.Println("DEBUG periodicMon", systemName, numConn)
			p := formPoint(systemName, numConnections, numConn)
			monСlient.WritePoint(p)
		}
		muConn.Unlock()

		p := formRedisPoint(redisTable, numConnections, connsRedis)
		monСlient.WritePoint(p)

		if conn != nil {
			sessions, _ := redis.ByteSlices(conn.Do("KEYS", "session:*"))
			for _, s := range sessions {
				id := strings.TrimPrefix(string(s[:]), "sessions:")
				n, _ := redis.Uint64(conn.Do("ZCOUNT", id, 0, util.Milliseconds()))
				p := formRedisPoint("NDTP", unconfirmedPkts, n)
				monСlient.WritePoint(p)
			}
			n, err := redis.Uint64(conn.Do("ZCOUNT", util.EgtsName, 0, util.Milliseconds()))
			if err == nil {
				p := formRedisPoint("NDTP", unconfirmedPkts, n)
				monСlient.WritePoint(p)
			}
			p = formRedisPoint("EGTS", unconfirmedPkts, n)
			monСlient.WritePoint(p)
			muRedis.Unlock()
		}

		time.Sleep(10 * time.Second)
	}
}

func NewConn(options *util.Options, systemName string) {
	log.Println("DEBUG NewConn", systemName)
	if !options.MonEnable {
		return
	}
	muConn.Lock()
	numConn := connsSystems[systemName]
	if numConn < math.MaxUint64 {
		numConn++
		connsSystems[systemName] = numConn
	}
	muConn.Unlock()
	p := formPoint(systemName, numConnections, numConn)
	options.MonСlient.WritePoint(p)
}

func DelConn(options *util.Options, systemName string) {
	log.Println("DEBUG DelConn", systemName)
	if !options.MonEnable {
		return
	}
	muConn.Lock()
	numConn := connsSystems[systemName]
	if numConn > 0 {
		numConn--
		connsSystems[systemName] = numConn
	}
	muConn.Unlock()
	p := formPoint(systemName, numConnections, numConn)
	options.MonСlient.WritePoint(p)
}

func SendMetric(options *util.Options, systemName string, metricName string, value interface{}) {
	if !options.MonEnable {
		return
	}
	p := formPoint(systemName, metricName, value)
	options.MonСlient.WritePoint(p)
}

func formPoint(systemName string, metricName string, value interface{}) *influx.Point {
	tags := influx.Tags{
		"host":     host,
		"instance": util.Instance,
	}
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

func formRedisPoint(systemName string, metricName string, value interface{}) *influx.Point {
	tags := influx.Tags{
		"host":     host,
		"instance": util.Instance,
	}
	if systemName != redisTable {
		tags["system"] = systemName
	}
	values := influx.Values{
		metricName: value,
	}
	p := influx.NewPoint(redisTable, tags, values)
	return p
}

func getHost() (ipStr string) {
	ipStr = "localhost"
	host, err := os.Hostname()
	if err != nil {
		return
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		return
	}
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			logrus.Println("IPv4: ", ipv4)
			ipStr = strings.ReplaceAll(ipv4.String(), ".", "_")
		}
	}
	return
}
