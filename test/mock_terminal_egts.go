package test

import (
	"net"
	"testing"
	"time"

	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/sirupsen/logrus"
)

var packetNav_Egts = []byte{1, 0, 0, 11, 0, 45, 0, 0, 0, 1, 47,
	34, 0, 0, 0, 1, 239, 0, 0, 0, 2, 2,
	16, 21, 0, 210, 49, 43, 16, 79, 186, 58, 158, 210, 39, 188, 53, 3, 0, 0, 178, 0, 0, 0, 0, 0,
	27, 7, 0, 32, 0, 0, 20, 0, 0, 0,
	148, 199}

func mockTerminal_Egts(t *testing.T, addr string, num int) {
	logger := logrus.WithFields(logrus.Fields{"test": "mock_terminal"})
	time.Sleep(100 * time.Millisecond)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Error(err)
	}
	defer util.CloseAndLog(conn, logger)
	for i := 0; i < num; i++ {
		err = sendNewMessage_Egts(t, conn, i, logger)
		if err != nil {
			logger.Errorf("got error: %v", err)
			t.Error(err)
		}
	}
	time.Sleep(10 * time.Second)
}

func sendNewMessage_Egts(t *testing.T, conn net.Conn, i int, logger *logrus.Entry) error {
	return sendAndReceive(t, conn, packetNav_Egts, logger)
}
