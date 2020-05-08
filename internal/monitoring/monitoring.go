package monitoring

import (
	//"github.com/ashirko/tcpmirror/internal/monitoring/influx"
	"./influx"
	"github.com/sirupsen/logrus"
)

type monInfo struct {
	enabled bool
	client  *influx.client
}

func Init(address string) (monInfo *monInfo, err error) {
	if address == "" {
		logrus.Println("start without sending metrics to influx")
		return
	}
	monInfo.enabled = true
	monInfo.client, err = influx.NewInfluxClient(address)
	return
}
