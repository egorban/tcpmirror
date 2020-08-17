package server

import (
	"net"
	"time"

	"github.com/ashirko/tcpmirror/internal/db"
	"github.com/ashirko/tcpmirror/internal/monitoring"
	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/egorban/navprot/pkg/egts"
	"github.com/sirupsen/logrus"
)

type egtsServer struct {
	conn net.Conn
	//terminalID int
	//sessionID  int
	logger   *logrus.Entry
	pool     *db.Pool
	exitChan chan struct{}
	*util.Options
	//masterIn    chan []byte
	//masterOut   chan []byte
	//ndtpClients []client.Client
	channels  []chan []byte
	recordNum uint32
	confChan  chan *db.ConfMsg
	name      string
}

func startEgtsServer(listen string, options *util.Options, channels []chan []byte, systems []util.System, confChan chan *db.ConfMsg) {
	pool := db.NewPool(options.DB)
	defer util.CloseAndLog(pool, logrus.WithFields(logrus.Fields{"main": "closing pool"}))
	l, err := net.Listen("tcp", listen)
	if err != nil {
		logrus.Fatalf("error while listening: %s", err)
		return
	}
	defer util.CloseAndLog(l, logrus.WithFields(logrus.Fields{"main": "closing listener"}))
	for {
		c, err := l.Accept()
		if err != nil {
			logrus.Errorf("error while accepting: %s", err)
		}
		logrus.Printf("accepted connection (%s <-> %s)", c.RemoteAddr(), c.LocalAddr())
		go initEgtsServer(c, pool, options, channels, systems, confChan)
	}
}

func initEgtsServer(c net.Conn, pool *db.Pool, options *util.Options, channels []chan []byte, systems []util.System,
	confChan chan *db.ConfMsg) {
	s, err := newEgtsServer(c, pool, options, channels, systems, confChan)
	if err != nil {
		logrus.Errorf("error during initialization new egts server: %s", err)
		return
	}
	s.logger.Tracef("newEgtsServer: %+v", s)
	go s.removeExpired()
	s.serverLoop()
}

func newEgtsServer(conn net.Conn, pool *db.Pool, options *util.Options, channels []chan []byte, systems []util.System,
	confChan chan *db.ConfMsg) (*egtsServer, error) {
	exitChan := make(chan struct{})
	return &egtsServer{
		conn:     conn,
		logger:   logrus.WithField("type", "egts_server"),
		pool:     pool,
		exitChan: exitChan,
		Options:  options,
		//masterIn:    master.InputChannel(),
		//masterOut:   master.OutputChannel(),
		//ndtpClients: append(clients, master),
		channels: channels,
		name:     monitoring.TerminalName,
	}, nil
}

func (s *egtsServer) serverLoop() {
	var buf []byte
	var b [defaultBufferSize]byte
	var ansPID uint16
	var ansRID uint16
	for {
		s.logger.Debug("start reading from client")
		if err := s.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			s.logger.Warningf("can't set read dead line: %s", err)
		}
		n, err := s.conn.Read(b[:])
		monitoring.SendMetric(s.Options, s.name, monitoring.RcvdBytes, n)
		//s.logger.Debugf("received %d from client", n)
		//util.PrintPacket(s.logger, "packet from client: ", b[:n])
		//todo remove after testing
		//	util.PrintPacketForDebugging(s.logger, "parsed packet from client:", b[:n])
		if err != nil {
			s.logger.Info("close ndtpServer: ", err)
			close(s.exitChan)
			return
		}
		buf = append(buf, b[:n]...)
		s.logger.Debugf("len(buf) = %d", len(buf))
		var numPacks uint
		buf, numPacks = s.processBuf(buf, ansPID, ansRID)
		monitoring.SendMetric(s.Options, s.name, monitoring.RcvdPkts, numPacks)
	}
}

func (s *egtsServer) processBuf(buf []byte, ansPID uint16, ansRID uint16) ([]byte, uint) {
	countPack := uint(0)
	for len(buf) > 0 {
		//packet, rest, service, _, nphID, err := ndtp.SimpleParse(buf)
		var packet egts.Packet
		rest, err := packet.Parse(buf)
		//s.logger.Tracef("service: %d, nphID: %d, packet: %v, err: %v", service, nphID, packet, err)
		//s.logger.Tracef("len packet: %d, len buf: %d, service: %d", len(packet), len(rest), service)
		if err != nil {
			if len(rest) > defaultBufferSize {
				s.logger.Warningf("drop buffer: %s", err)
				return []byte(nil), countPack
			}
			return rest, countPack
		}
		buf = rest
		if packet.Type == egts.EgtsPtAppdata {
			err = s.processPacket(packet, ansPID, ansRID)
			if err != nil {
				s.logger.Warningf("can't process message from client: %s", err)
				return []byte(nil), countPack
			}
			countPack++
		}
	}
	return buf, countPack
}

func (s *egtsServer) processPacket(packet egts.Packet, ansPID uint16, ansRID uint16) (err error) {
	var recNums []uint16
	for _, rec := range packet.Records {
		data := util.Data_Egts{
			OID:    rec.ID,
			PackID: packet.ID,
			RecID:  rec.RecNum,
			Record: rec.recBin, //record из которой потом формировать пакет (заголовок записи + подзаписи)
		}
		s.recordNum++
		sdata := util.Serialize_Egts(data)
		err = db.Write2DB_Egts(s.pool, int(data.OID), sdata, s.logger)
		if err != nil {
			return
		}
		s.send2Channels(sdata)
		recNums = append(recNums, rec.RecNum)
	}
	reply, _ := makeReply(packet.ID, recNums, ansPID, ansRID)
	err = s.send2terminal(reply)
	return
}

func makeReply(packetID uint16, recNums []uint16, ansPID uint16, ansRID uint16) ([]byte, error) {
	subRecords := make([]*egts.SubRecord, 0, 1)
	for _, num := range recNums {
		subData := egts.Confirmation{
			CRN: num,
			RST: 0,
		}
		sub := &egts.SubRecord{
			Type: egts.EgtsSrResponse,
			Data: &subData,
		}
		subRecords = append(subRecords, sub)
	}
	data := egts.Response{
		RPID:    packetID,
		ProcRes: 0,
	}

	rec := egts.Record{
		RecNum:  ansRID,
		Service: egts.EgtsTeledataService,
		Data:    subRecords,
	}
	packetData := &egts.Packet{
		Type:    egts.EgtsPtResponse,
		ID:      ansPID,
		Records: []*egts.Record{&rec},
		Data:    &data,
	}
	return packetData.Form()
}

func (s *egtsServer) send2terminal(packet []byte) (err error) {
	util.PrintPacket(s.logger, "send to terminal:", packet)
	err = s.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	//todo remove after testing
	util.PrintPacketForDebugging(s.logger, "parsed packet to client:", packet)
	if err != nil {
		return
	}
	var n int
	n, err = s.conn.Write(packet)
	monitoring.SendMetric(s.Options, s.name, monitoring.SentBytes, n)
	return
}

func (s *egtsServer) send2Channels(data []byte) {
	s.logger.Tracef("send2Channels %v : %v", s.channels, data)
	for _, channel := range s.channels {
		s.send2Channel(channel, data)
	}
}

func (s *egtsServer) send2Channel(channel chan []byte, data []byte) {
	copyData := util.Copy(data)
	select {
	case channel <- copyData:
		return
	default:
		s.logger.Warningln("channel is full")
	}
}

func (s *egtsServer) removeExpired() {
	tickerEx := time.NewTicker(1 * time.Hour)
	defer tickerEx.Stop()
	for {
		select {
		case <-s.exitChan:
			return
		case <-tickerEx.C:
			err := db.RemoveExpired(s.pool, 1, s.logger)
			if err != nil {
				s.logger.Errorf("can't remove expired data ndtp: %s", err)
			}
		}
	}
}
