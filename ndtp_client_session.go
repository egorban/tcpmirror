package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"time"
)

import nav "github.com/ashirko/navprot"

//session with NDTP client
func clientSession(s *session) {
	if err := removeServerExt(s); err != nil {
		s.logger.Warningf("can't removeServerExt: %s", err)
	}
	go oldFromClient(s)
	go ndtpRemoveExpired(s)
	clientSessionLoop(s)
}

func clientSessionLoop(s *session) {
	var restBuf []byte
	var b [defaultBufferSize]byte
	for {
		s.logger.Debug("start reading from client")
		if err := s.clientConn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			s.logger.Warningf("can't set read dead line: %s", err)
		}
		n, err := s.clientConn.Read(b[:])
		s.logger.Debugf("received %d from client", n)
		printPacket(s.logger, "packet from client: ", b[:n])
		if err != nil {
			s.errClientCh <- err
			return
		}
		if enableMetrics {
			countFromServerNDTP.Inc(1)
		}
		restBuf = append(restBuf, b[:n]...)
		s.logger.Debugf("len(restBuf) = %d", len(restBuf))
		restBuf = processRestBuf(s, restBuf)
	}
}

func processRestBuf(s *session, restBuf []byte) []byte {
	var err error
	for len(restBuf) > 0 {
		ndtp := new(nav.NDTP)
		s.logger.Tracef("before parsing len(restBuf) = %d", len(restBuf))
		printPacket(s.logger, "before parsing restBuf", restBuf)
		restBuf, err = ndtp.Parse(restBuf)
		if err != nil {
			if len(restBuf) > defaultBufferSize {
				return []byte(nil)
			}
			return restBuf
		}
		if enableMetrics {
			countClientNDTP.Inc(1)
		}
		err = processNdtp(s, ndtp)
		if err != nil {
			s.logger.Warningf("can't process message from client: %s", err)
			return []byte(nil)
		}
		time.Sleep(1 * time.Millisecond)
	}
	return restBuf
}

func processNdtp(s *session, ndtp *nav.NDTP) (err error) {
	if ndtp.IsResult() {
		err = nphResultFromClient(s, ndtp)
	} else if ndtp.Service() == nav.NphSrvExternalDevice {
		err = extFromClient(s, ndtp)
	} else if !ndtp.NeedReply() {
		err = noNeedReplyFromClient(s, ndtp)
	} else {
		err = ndtpFromClient(s, ndtp)
	}
	return
}

// nphResultFromClient processes client response to NPH_ SRV_GENERIC_CONTROLS message from server
func nphResultFromClient(s *session, ndtp *nav.NDTP) (err error) {
	controlReplyID, err := readControlID(s, int(ndtp.Nph.ReqID))
	if err != nil {
		s.logger.Warningf("can't find controlReplyID: %s", err)
		return
	}
	s.logger.Debugf("old nphReqID: %d; new nphReqID: %d", ndtp.Nph.ReqID, controlReplyID)
	printPacket(s.logger, "control message before changing: ", ndtp.Packet)
	changes := map[string]int{nav.NplReqID: int(s.serverNplID()), nav.NphReqID: controlReplyID}
	ndtp.ChangePacket(changes)
	printPacket(s.logger, "control message after changing: ", ndtp.Packet)
	err = sendToServer(s, ndtp)
	return
}

// noNeedReplyFromClient processes messages from client to server which do not need reply
func noNeedReplyFromClient(s *session, ndtp *nav.NDTP) (err error) {
	s.logger.Debugf("no need to reply on message ServiceID: %d, NPHReqID: %d", ndtp.Nph.ServiceID, ndtp.Nph.ReqID)
	changes := map[string]int{nav.NplReqID: int(s.serverNplID()), nav.NphReqID: int(s.serverNPHReqID)}
	ndtp.ChangePacket(changes)
	printPacket(s.logger, "send message to server (no reply): ", ndtp.Packet)
	err = sendToServer(s, ndtp)
	return
}

// extFromClient processes NPH_SRV_EXTERNAL_DEVICE messages from client to server
func extFromClient(s *session, ndtp *nav.NDTP) (err error) {
	mill := milliseconds()
	s.logger.Debugf("handle NPH_SRV_EXTERNAL_DEVICE")
	pType := ndtp.PacketType()
	if pType == nav.NphSedDeviceTitleData {
		err = extTitleFromClient(s, ndtp, mill)
	} else if pType == nav.NphSedDeviceResult {
		err = extResFromClient(s, ndtp, mill)
	} else {
		err = fmt.Errorf("unknown NPHType: packet %v", ndtp.Packet)
	}
	return
}

// extTitleFromClient processes NPH_SED_DEVICE_TITLE_DATA messages from client to server
func extTitleFromClient(s *session, ndtp *nav.NDTP, mill int64) (err error) {
	c := pool.Get()
	defer closeAndLog(c, s.logger)
	packetCopy := append([]byte(nil), ndtp.Packet...)
	err = writeExtClient(c, s, mill, packetCopy)
	if err != nil {
		s.logger.Errorf("extTitleFromClient: send ext error reply to client because of: %s", err)
		err1 := replyExt(s, ndtp, nphResultError)
		if err1 != nil {
			return err1
		}
		return
	}
	s.logger.Println("send ext device message to server")
	if s.servConn.closed != true {
		err = writeNDTPIdExt(c, s, ndtp.Nph.Data.(nav.ExtDevice).MesID, mill)
		if err != nil {
			s.logger.Errorf("error writeNDTPIdExt: %v", err)
		} else {
			packetCopy1 := append([]byte(nil), ndtp.Packet...)
			changes := map[string]int{nav.NplReqID: int(s.serverNplID()), nav.NphReqID: int(s.serverNPHReqID)}
			ndtp.ChangePacket(changes)
			printPacket(s.logger, "ext device message to server after change: ", ndtp.Packet)
			err = sendToServer(s, ndtp)
			ndtp.Packet = packetCopy1
			if err != nil {
				s.logger.Warningf("can't send ext device message to NDTP server: %s", err)
				ndtpConStatus(s)
			} else {
				if enableMetrics {
					countToServerNDTP.Inc(1)
				}
			}
		}
	}
	s.logger.Debugln("start to reply to ext device message")
	err = replyExt(s, ndtp, nphResultOk)
	if err != nil {
		s.logger.Warningf("can't reply to ext device message: ", err)
		s.errClientCh <- err
	}
	return
}

// extResFromClient processes NPH_SED_DEVICE_RESULT messages from client to server
func extResFromClient(s *session, ndtp *nav.NDTP, mill int64) (err error) {
	c := pool.Get()
	defer closeAndLog(c, s.logger)
	_, _, _, mesID, err := servExt(c, s)
	if err != nil {
		s.logger.Warningf("can't servExt: %v", err)
	} else if mesID == uint64(ndtp.Nph.Data.(nav.ExtDevice).MesID) {
		if ndtp.Nph.Data.(nav.ExtDevice).Res == 0 {
			s.logger.Debugf("received result and remove data from db")
			err = removeServerExt(s)
			if err != nil {
				s.logger.Warningf("can't removeFromNDTPExt : %v", err)
			}
		} else {
			s.logger.Println("received result with error status")
			err = setFlagServerExt(c, s, "1")
			if err != nil {
				s.logger.Warningf("can't setFlagServerExt: %v", err)
			}
		}
	} else {
		s.logger.Warningf("receive reply with mesID: %d; messID stored in DB: %d", ndtp.Nph.Data.(nav.ExtDevice).MesID, mesID)
	}
	packetCopy := append([]byte(nil), ndtp.Packet...)
	changes := map[string]int{nav.NplReqID: int(s.serverNplID()), nav.NphReqID: int(s.serverNPHReqID)}
	ndtp.ChangePacket(changes)
	printPacket(s.logger, "send ext device message to server: ", ndtp.Packet)
	err = sendToServer(s, ndtp)
	if err != nil {
		s.logger.Warningf("can't write ext device result to server: %v", err)
		err = writeExtClient(c, s, mill, packetCopy)
		if err != nil {
			s.logger.Errorf("can't write2DB ext device result: %v", err)
		}
	}
	return
}

func ndtpFromClient(s *session, ndtp *nav.NDTP) (err error) {
	c := pool.Get()
	defer closeAndLog(c, s.logger)
	mill := milliseconds()
	packetCopy := append([]byte(nil), ndtp.Packet...)
	err = write2DB(c, s, packetCopy, mill, toEGTS(ndtp))
	if err != nil {
		s.logger.Errorf("can't write2DB: ", err)
		err1 := reply(s, ndtp, nphResultError)
		if err1 != nil {
			return err1
		}
		return
	}
	s.logger.Debugf("start send to NDTP server")
	if s.servConn.closed != true {
		fromCtoS(s, ndtp, c, mill)
	}
	if egtsConn.closed != true {
		if toEGTS(ndtp) {
			sendToEgts(s, ndtp, mill)
		}
	}
	s.logger.Debugln("start to reply")
	err = reply(s, ndtp, nphResultOk)
	if err != nil {
		s.logger.Warningf("ndtpFromClient: error replying to att: ", err)
		s.errClientCh <- err
	}
	return
}

func sendToEgts(s *session, ndtp *nav.NDTP, mill int64) {
	m := new(egtsMsg)
	m.msgID = strconv.Itoa(s.id) + ":" + strconv.FormatInt(mill, 10)
	var err error
	m.msg, err = nav.NDTPtoEGTS(*ndtp, uint32(s.id))
	if err != nil {
		s.logger.Errorf("error converting to egts: %s", err)
	}
	s.logger.Debugln("start to send to EGTS goroutine")
	select {
	case egtsCh <- m:
	default:
		s.logger.Errorln("egtsCh is full")
	}
}

func fromCtoS(s *session, ndtp *nav.NDTP, c redis.Conn, mill int64) {
	changes := map[string]int{nav.NplReqID: int(s.serverNplID()), nav.NphReqID: int(s.serverNPHReqID)}
	ndtp.ChangePacket(changes)
	printPacket(s.logger, "packet after changing: ", ndtp.Packet)
	err := writeNDTPid(c, s, ndtp.Nph.ReqID, mill)
	if err != nil {
		s.logger.Errorf("error writingNDTPid: %v", err)
	} else {
		printPacket(s.logger, "send message to server: ", ndtp.Packet)
		err = sendToServer(s, ndtp)
		if err != nil {
			s.logger.Warningf("can't send to NDTP server: %s", err)
			ndtpConStatus(s)
		} else {
			if enableMetrics {
				countToServerNDTP.Inc(1)
			}
		}
	}
}

func toEGTS(ndtp *nav.NDTP) bool {
	switch ndtp.Nph.Data.(type) {
	case *nav.NavData:
		return true
	default:
		return false
	}
}

func reply(s *session, ndtp *nav.NDTP, result uint32) error {
	if ndtp.IsResult() {
		return nil
	}
	reply := ndtp.Reply(result)
	printPacket(s.logger, "reply: send answer: ", reply)
	err := sendToClient(s, reply)
	return err
}

func replyExt(s *session, ndtp *nav.NDTP, result uint32) error {
	ans, err := ndtp.ReplyExt(result)
	if err != nil {
		s.logger.Errorf("ReplyExt error: %s", err)
		return err
	}
	printPacket(s.logger, "send answer: ", ans)
	err = sendToClient(s, ans)
	return err
}

func sendToClient(s *session, packet []byte) error {
	err := s.clientConn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		return err
	}
	_, err = s.clientConn.Write(packet)
	return err
}

func sendToServer(s *session, ndtp *nav.NDTP) error {
	err := s.servConn.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		return err
	}
	_, err = s.servConn.conn.Write(ndtp.Packet)
	return err
}
