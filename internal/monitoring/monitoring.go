package monitoring

import (
	"github.com/egorban/influx/pkg/influx"
	"github.com/sirupsen/logrus"
)

const (
	visTable = "vis"
	attTable = "source"
)

var (
	Instance    string
	defaultTags influx.Tags
)

type MonInfo struct {
	enabled bool
	client  *influx.Client
}

func Init(address string) (monInfo *MonInfo, err error) {
	if address == "" {
		logrus.Println("start without sending metrics to influx")
		return
	}
	monClient, err := influx.NewClient(address, 30*1000, 5000)
	if err != nil {
		logrus.Println("error while connecting to influx", err)
		return
	}
	monInfo = new(MonInfo)
	monInfo.client = monClient
	defaultTags = setDefaultTags()
	monInfo.enabled = true
	return
}

func (info *MonInfo) AttSendMetric(metricName string, value interface{}) {
	info.SendMetric(attTable, "", metricName, value)
}

func (info *MonInfo) VisSendMetric(systemName string, metricName string, value interface{}) {
	info.SendMetric(visTable, systemName, metricName, value)
}

func (info *MonInfo) SendMetric(table string, systemName string, metricName string, value interface{}) {
	if !info.enabled {
		return
	}
	tags := defaultTags
	values := influx.Values{
		metricName: value,
	}
	if table == visTable {
		tags["system"] = systemName
	}
	p := influx.NewPoint(table, tags, values)
	info.client.WritePoint(p)
}

func setDefaultTags() influx.Tags {
	host := "10_1_116_55"
	return influx.Tags{
		"host":     host,
		"instance": Instance,
	}
}
