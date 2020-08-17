package client

import (
	"net"
	"time"

	"github.com/ashirko/tcpmirror/internal/db"
	"github.com/ashirko/tcpmirror/internal/monitoring"
	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/egorban/navprot/pkg/egts"
)

// EgtsChanSize defines size of EGTS client input chanel buffer
//const EgtsChanSize = 10000

// Egts describes EGTS client
// type Egts struct {
// 	Input  chan []byte
// 	dbConn db.Conn
// 	*info
// 	*egtsSession
// 	*connection
// 	confChan chan *db.ConfMsg
// }

// type egtsSession struct {
// 	egtsMessageID uint16
// 	egtsRecID     uint16
// 	mu            sync.Mutex
// }

// NewEgts creates new Egts client
// func NewEgts(sys util.System, options *util.Options, confChan chan *db.ConfMsg) *Egts {
// 	c := new(Egts)
// 	c.info = new(info)
// 	c.egtsSession = new(egtsSession)
// 	c.connection = new(connection)
// 	c.id = sys.ID
// 	c.name = sys.Name
// 	c.address = sys.Address
// 	c.logger = logrus.WithFields(logrus.Fields{"type": "egts_client", "vis": sys.ID})
// 	c.Options = options
// 	c.Input = make(chan []byte, EgtsChanSize)
// 	c.confChan = confChan
// 	return c
// }

// InputChannel implements method of Client interface
// func (c *Egts) InputChannel() chan []byte {
// 	return c.Input
// }

// // OutputChannel implements method of Client interface
// func (c *Egts) OutputChannel() chan []byte {
// 	return nil
// }

func (c *Egts) start_Egts() {
	c.logger.Traceln("start")
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		c.logger.Errorf("error while connecting to EGTS server: %s", err)
		c.reconnect()
	} else {
		c.conn = conn
		c.open = true
	}
	go c.old()
	go c.replyHandler()
	c.clientLoop_Egts()
}

func (c *Egts) clientLoop_Egts() {
	dbConn := db.Connect(c.DB)
	defer c.closeDBConn(dbConn)
	err := c.getID(dbConn)
	if err != nil {
		c.logger.Errorf("can't getID: %v", err)
	}
	var buf []byte
	var records [][]byte
	countRec := 0
	countPack := 0
	sendTicker := time.NewTicker(100 * time.Millisecond)
	defer sendTicker.Stop()
	for {
		if c.open {
			select {
			case record := <-c.Input:
				monitoring.SendMetric(c.Options, c.name, monitoring.QueuedPkts, len(c.Input))
				if db.CheckOldData(dbConn, record, c.logger) {
					continue
				}
				records = c.processRecord(dbConn, record, records)
				countRec++
				if countRec == 3 {
					buf, _ = c.formPacket(records)
					countPack++
					records = [][]byte(nil)
					countRec = 0
				}
				if countPack == 10 {
					err := c.send(buf)
					if err == nil {
						//monitoring.SendMetric(c.Options, c.name, monitoring.SentPkts, count)
					}
					buf = []byte(nil)
					countPack = 0
				}
			case <-sendTicker.C:
				if (countRec > 0) && (countRec < 3) {
					buf, _ = c.formPacket(records)
					countPack++
					records = [][]byte(nil)
					countRec = 0
				}
				if (countPack > 0) && (countPack <= 10) {
					err := c.send(buf)
					if err == nil {
						//	monitoring.SendMetric(c.Options, c.name, monitoring.SentPkts, count)
					}
					buf = []byte(nil)
					countPack = 0
				}
			}
		} else {
			time.Sleep(time.Duration(TimeoutClose) * time.Second)
			buf = []byte(nil)
			records = [][]byte(nil)
			countRec = 0
			countPack = 0
		}
	}
}

func (c *Egts) processRecord(dbConn db.Conn, record []byte, records [][]byte) [][]byte {
	util.PrintPacket(c.logger, "serialized data: ", record)
	data := util.Deserialize_Egts(record)
	c.logger.Tracef("data: %+v", data)
	_, recID, err := c.ids(dbConn)
	if err != nil {
		c.logger.Errorf("can't get ids: %s", err)
		return records
	}
	records = append(records, record)
	err = db.WriteEgtsID(dbConn, c.id, recID, data.ID)
	if err != nil {
		c.logger.Errorf("error WriteEgtsID: %s", err)
	}
	return records
}

func (c *Egts) formPacket(recordsBin [][]byte) ([]byte, error) {
	var records []*egts.Record
	for _, recBin := range recordsBin {
		record := &egts.Record{
			RecBin: recBin,
		}
		records = append(records, record)
	}
	packetData := egts.Packet{
		Type:    egts.EgtsPtAppdata,
		ID:      0,
		Records: records,
		Data:    nil,
	}
	return packetData.Form()
}
