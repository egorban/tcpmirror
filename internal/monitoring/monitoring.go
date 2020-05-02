package monitoring

import (
	"bufio"
	"fmt"
	"net"
	"strconv"

	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/sirupsen/logrus"
)

var (
	addr     *net.UDPAddr
	host     string
	instance string
)

func Init(address string) (enable bool, err error) {
	if address == "" {
		logrus.Println("start without sending metrics to influx")
		return
	}
	addr, err = net.ResolveUDPAddr("udp", address)
	if err != nil {
		logrus.Errorf("error while connecting to influx: %s\n", err)
		return
	}
	host = "10_1_116_55"           //TODO: динамическое вычисление
	instance = util.InstancePrefix //TODO: динамическое вычисление
	logrus.Infof("start sending metrics to influx to %+s:%+v, prefix: %+s",
		addr.IP, addr.Port, util.InstancePrefix)
	return true, nil
}

//формирование строки для отправки
func SendBytes(table string, system string, count int) { // int?
	record := table + ",host=" + host + ",instance=" + instance + ",system=" + system +
		" sentBytes=" + strconv.Itoa(count)
	logrus.Println("SendBytes %+v", record)
	sendToInflux(record)
}

// отправка строки по udp
func sendToInflux(record string) error {
	conn, err := net.DialUDP("udp", nil, addr)
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
