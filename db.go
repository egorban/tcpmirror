package main

import (
	"bytes"
	"encoding/binary"
	"github.com/gomodule/redigo/redis"
	"log"
	"strconv"
	"strings"
)

func writeConnDB(c redis.Conn, id uint32, message []byte) error {
	key := "conn:" + strconv.Itoa(int(id))
	log.Printf("writeConnDB key: %s; message: %v", key, message)
	_, err := c.Do("SET", key, message)
	return err
}

func readConnDB(c redis.Conn, id int) ([]byte, error) {
	key := "conn:" + strconv.Itoa(id)
	res, err := redis.Bytes(c.Do("GET", key))
	log.Printf("readConnDB key: %s; message: %v", key, res)
	return res, err
}

func writeNDTPid(c redis.Conn, id, nphID uint32, mill int64) error {
	key := "ntn:" + strconv.FormatUint(uint64(id), 10) + ":" + strconv.FormatUint(uint64(nphID), 10)
	log.Printf("writeNDTPid key: %s; val: %d", key, mill)
	_, err := c.Do("SET", key, mill, "ex", 50)
	return err
}

func readNDTPid(c redis.Conn, id int, nphID uint32) (int64, error) {
	key := "ntn:" + strconv.FormatUint(uint64(id), 10) + ":" + strconv.FormatUint(uint64(nphID), 10)
	res, err := redis.Int64(c.Do("GET", key))
	log.Printf("readNDTPid key: %s; res: %d", key, res)
	return res, err
}

func write2DB(c redis.Conn, data ndtpData, s *session, packet []byte, time int64) (err error) {
	err = write2NDTP(c, s.id, time, packet)
	if err != nil {
		return
	}
	if data.ToRnis.time != 0 {
		err = write2EGTS(c, s.id, time, packet)
	}
	return
}

func write2NDTP(c redis.Conn, id int, time int64, packet []byte) error {
	log.Printf("write2NDTP id: %d; time: %d; packet: %v", id, time, packet)
	_, err := c.Do("ZADD", id, time, packet)
	return err
}

func write2EGTS(c redis.Conn, id int, time int64, packet []byte) error {
	idB := new(bytes.Buffer)
	binary.Write(idB, binary.LittleEndian, uint32(id))
	packet = append(idB.Bytes(), packet...)
	log.Printf("write2EGTS id: %d; time: %d; packet: %v", id, time, packet)
	_, err := c.Do("ZADD", "rnis", time, packet)
	return err
}

func removeFromNDTP(c redis.Conn, id int, NPHReqID uint32) error {
	time, err := readNDTPid(c, id, NPHReqID)
	if err != nil {
		return err
	}
	log.Printf("removeFromNDTP: id: %d; time: %d", id, time)
	_, err = c.Do("ZREMRANGEBYSCORE", id, time, time)
	return err
}

func getOldNDTP(c redis.Conn, id int) ([][]byte, error) {
	max := getMill() - 60000
	res, err := redis.ByteSlices(c.Do("ZRANGEBYSCORE", id, 0, max, "LIMIT", 0, 10))
	log.Printf("getOldNDTP: id: %d; max: %d; len(res): %d", id, max, len(res))
	return res, err
}

func writeEGTSid(c redis.Conn, egtsMessageID uint16, messageID string) (err error) {
	key := "egts:" + strconv.Itoa(int(egtsMessageID))
	log.Printf("writeEGTSid: key: %s; messageID: %s", key, messageID)
	_, err = c.Do("SET", key, messageID, "ex", 50)
	return
}

func deleteEGTS(c redis.Conn, egtsMessageID uint16) (err error) {
	key := "egts:" + strconv.Itoa(int(egtsMessageID))
	messageID, err := redis.String(c.Do("GET", key))
	log.Printf("deleteEGTS 1: key: %s; mssageID: %s", key, messageID)
	if err != nil {
		log.Println("error get EGTS message id from db: ", err)
		return
	}
	log.Println("get messageID: ", messageID)
	messageIDSplit := strings.Split(messageID, ":")
	log.Println("messageIDSplit: ", messageIDSplit)
	id, err := strconv.ParseUint(messageIDSplit[1], 10, 32)
	time, err := strconv.ParseInt(messageIDSplit[1], 10, 64)
	idB := new(bytes.Buffer)
	binary.Write(idB, binary.LittleEndian, id)

	packets, err := redis.ByteSlices(c.Do("ZRANGEBYSCORE", "rnis", time, time))
	log.Printf("deleteEGTS 2: key: %s; mssageID: %d", key, time)
	if err != nil {
		log.Println("error get EGTS packets from db: ", err)
		return
	}
	numPackets := len(packets)
	switch {
	case numPackets > 1:
		for _, pack := range packets {
			if bytes.Compare(pack[0:4], idB.Bytes()) == 0 {
				_, err = c.Do("ZREM", "rnis", pack)
				if err != nil {
					log.Println("error while deleting EGTS packet from db")
				}
				return
				log.Println("where is no EGTS packets for EGTSMessageID: ", egtsMessageID, "; messageID: ", messageID)
			}
		}
	case numPackets == 1:
		_, err = c.Do("ZREM", "rnis", packets[0])
		if err != nil {
			log.Println("error while deleting EGTS packet from db")
		}
	default:
		log.Println("where is no EGTS packets for ", egtsMessageID)
		return
	}
	return
}

func removeExpiredDataEGTS(c redis.Conn) (err error) {
	max := getMill() - 259200000 //3*24*60*60*1000
	_, err = c.Do("ZREMRANGEBYSCORE", "rnis", 0, max)
	log.Printf("removeExpiredDataEGTS: max: %d", max)
	return
}

func getOldEGTS(c redis.Conn) (res [][]byte, err error) {
	max := getMill() - 60000
	res, err = redis.ByteSlices(c.Do("ZRANGEBYSCORE", "rnis", 0, max, "LIMIT", 0, 10000))
	log.Printf("getOldEGTS: max: %d; len(res): %d", max, len(res))
	return
}
