package monitoring

import (
	"bytes"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ashirko/tcpmirror/internal/db"
	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/cakturk/go-netstat/netstat"
	"github.com/egorban/influx/pkg/influx"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

const (
	TerminalName = "terminal"

	visTable   = "vis"
	attTable   = "source"
	redisTable = "redis"

	SentBytes      = "sentBytes"
	RcvdBytes      = "rcvdBytes"
	SentPkts       = "sentPkts"
	RcvdPkts       = "rcvdPkts"
	QueuedPkts     = "queuedPkts"
	numConnections = "numConnections"
	unConfPkts     = "unConfPkts"
)

var (
	connsSystems map[string]uint64
	muConn       sync.Mutex
	host         string
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
	go periodicMonConns(monClient, 10*time.Second)
	go periodicMonRedisConns(monClient, 60*time.Second, dbAddress)
	go periodicMonRedisUnconfPkts(monClient, 60*time.Second, dbAddress)
	return
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

func periodicMonConns(monСlient *influx.Client, period time.Duration) {
	for {
		muConn.Lock()
		for systemName, numConn := range connsSystems {
			log.Println("DEBUG periodicMon", systemName, numConn)
			p := formPoint(systemName, numConnections, numConn)
			monСlient.WritePoint(p)
		}
		muConn.Unlock()
		time.Sleep(period)
	}
}

func periodicMonRedisConns(monСlient *influx.Client, period time.Duration, dbAddress string) {
	_, dbPortStr, err := net.SplitHostPort(dbAddress)
	if err != nil {
		logrus.Println("periodicMonRedisConns error", err)
		return
	}
	dbPort1, err := strconv.Atoi(dbPortStr)
	dbPort := uint16(dbPort1)
	for {
		tabs, err := netstat.TCPSocks(func(s *netstat.SockTabEntry) bool {
			return s.State == netstat.Established && s.LocalAddr.Port == dbPort
		})
		log.Println("DEBUG periodicMonRedisConns", tabs)
		n := len(tabs)
		if err == nil {
			p := formPoint(redisTable, numConnections, n)
			monСlient.WritePoint(p)
		} else {
			logrus.Println("error count redis connections", err)
		}
		time.Sleep(period)
	}
}

func periodicMonRedisUnconfPkts(monСlient *influx.Client, period time.Duration, dbAddress string) {
	for {
		conn := db.Connect(dbAddress)
		defer db.Connect(dbAddress)
		if conn != nil {
			nEgts, err := redis.Uint64(conn.Do("ZCOUNT", util.EgtsName, 0, util.Milliseconds()))
			if err == nil {
				p := formPoint(redisTable, unConfPkts, nEgts)
				monСlient.WritePoint(p)
			}
			log.Println("DEBUG periodicMonRedisUnconfPkts", nEgts)
			sessions, err := redis.ByteSlices(conn.Do("KEYS", "session:*"))
			if err == nil {
				nNdtp := uint64(0)
				prefix := []byte("session:")
				for _, s := range sessions {

					id := bytes.TrimPrefix(s, prefix)
					n, err := redis.Uint64(conn.Do("ZCOUNT", id, 0, util.Milliseconds()))
					if err == nil {
						nNdtp = nNdtp + n
					}
				}
				p := formPoint(redisTable, unConfPkts, nNdtp)
				monСlient.WritePoint(p)
			}
		}

		time.Sleep(period)
	}
}

func formPoint(systemName string, metricName string, value interface{}) *influx.Point {
	host := "10_1_116_55"
	tags := influx.Tags{
		"host":     host,
		"instance": util.Instance,
	}
	var table string
	if systemName == TerminalName {
		table = attTable
	} else if systemName == redisTable {
		table = redisTable
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
