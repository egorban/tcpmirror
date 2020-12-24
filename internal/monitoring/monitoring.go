package monitoring

import (
	"os"
	"time"

	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/egorban/influx/pkg/influx"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

const (
	TerminalName = "terminal"

	visTable = "vis"
	attTable = "source"

	periodMonSystemConns = 10 * time.Second
)

var (
	host        string
	listenPort  uint16
	pidInstance int
	systems     []sysInfo
)

type sysInfo struct {
	name      string
	ipAddress string
	port      uint16
}

func Init(args *util.Args) (monEnable bool, monClient *influx.Client, err error) {
	monAddress := args.Monitoring
	if monAddress == "" {
		logrus.Println("start without sending metrics to influx")
		return
	}
	monClient, err = influx.NewClient(monAddress)
	if err != nil {
		logrus.Errorln("error while connecting to influx", err)
		return
	}
	host = getHost()
	monEnable = true
	startSystemsPeriodicMon(monClient, args)
	return
}
func startSystemsPeriodicMon(monClient *influx.Client, args *util.Args) {
	if args.Systems == nil {
		logrus.Errorln("error get systems, start without monitoring system connections")
		return
	}
	systems = make([]sysInfo, 0)
	for _, sys := range args.Systems {
		ipAddress, port, err := splitAddrPort(sys.Address)
		if err != nil {
			logrus.Errorln("error get systems, start without monitoring system connections", err)
			return
		}
		if ipAddress == "localhost" {
			ipAddress = "127.0.0.1"
		}
		system := sysInfo{
			name:      sys.Name,
			ipAddress: ipAddress,
			port:      port,
		}
		systems = append(systems, system)
	}

	listenAddress := args.Listen
	if listenAddress == "" {
		logrus.Errorln("error get listen address, start without monitoring redis")
		return
	}

	_, lp, err := splitAddrPort(listenAddress)
	if err != nil {
		logrus.Errorln("error get listen port, start without monitoring redis", err)
		return
	}
	listenPort = lp

	pidInstance = os.Getpid()
	if pidInstance == 0 {
		logrus.Errorln("error get pid, start without monitoring system connections")
		return
	}
	go monSystemConns(monClient)
}

func MonPoolDB(monClient *influx.Client, pool *redis.Pool) {
	logrus.Println("start monitoring pool")
	for {
		time.Sleep(10 * time.Second)
		st := pool.Stats()
		values := influx.Values{
			"active": st.ActiveCount,
			"idle":   st.IdleCount,
		}

		monClient.WritePoint(influx.NewPoint("info", influx.Tags{}, values))
	}
}
