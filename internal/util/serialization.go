package util

import "encoding/binary"

// PacketStart defines number of byte with metadata in front of binary packet.
// Structure of serialized data is 'PacketStart' bytes with metadata, then binary packet.
const PacketStart = 10

const PacketStart_Egts = 8

// Data defines deserialized data
type Data struct {
	TerminalID uint32
	SessionID  uint16
	PacketNum  uint32
	Packet     []byte
	ID         []byte
}

type Data_Egts struct {
	OID    uint32
	PackID uint16
	RecID  uint16
	Record []byte
	ID     []byte
}

// Serialize is using for serializing data
func Serialize(data Data) []byte {
	bin := make([]byte, PacketStart)
	binary.LittleEndian.PutUint32(bin[:4], data.TerminalID)
	binary.LittleEndian.PutUint16(bin[4:6], data.SessionID)
	binary.LittleEndian.PutUint32(bin[6:], data.PacketNum)
	return append(bin, data.Packet...)
}

// Deserialize is using for deserializing data
func Deserialize(bin []byte) Data {
	data := Data{}
	data.TerminalID = binary.LittleEndian.Uint32(bin[0:4])
	//data.SessionID = binary.LittleEndian.Uint16(bin[4:6])
	//data.PacketNum = binary.LittleEndian.Uint32(bin[6:10])
	data.ID = bin[:10]
	data.Packet = bin[PacketStart:]
	return data
}

func Serialize_Egts(data Data_Egts) []byte {
	bin := make([]byte, PacketStart_Egts)
	binary.LittleEndian.PutUint32(bin[:4], data.OID)
	binary.LittleEndian.PutUint16(bin[4:6], data.PackID)
	binary.LittleEndian.PutUint16(bin[6:], data.RecID)
	return append(bin, data.Record...)
}

// Deserialize is using for deserializing data
func Deserialize_Egts(bin []byte) Data_Egts {
	data := Data_Egts{}
	data.OID = binary.LittleEndian.Uint32(bin[0:4])
	//data.SessionID = binary.LittleEndian.Uint16(bin[4:6])
	//data.PacketNum = binary.LittleEndian.Uint32(bin[6:10])
	data.ID = bin[:PacketStart_Egts]
	data.Record = bin[PacketStart_Egts:]
	return data
}

// TerminalID returns TerminalID from serialized data
func TerminalID(bin []byte) uint32 {
	return binary.LittleEndian.Uint32(bin[0:4])
}
