package client

import (
	"net"
	"sync"
	"time"

	"github.com/ashirko/tcpmirror/internal/monitoring"
)

type ndtpSession struct {
	nplID uint16
	nphID uint32
	mu    sync.Mutex
}

func reverseSlice(res [][]byte) [][]byte {
	for i := len(res)/2 - 1; i >= 0; i-- {
		opp := len(res) - 1 - i
		res[i], res[opp] = res[opp], res[i]
	}
	return res
}

func send(conn net.Conn, name string, packet []byte) error {
	err := conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		return err
	}
	n, err := conn.Write(packet)
	monitoring.SendMetric(name, monitoring.SentBytes, n)
	return err
}
