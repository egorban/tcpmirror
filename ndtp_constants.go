package main

//
//import (
//	"bytes"
//	"encoding/binary"
//)
//
////NPL constants
//
//var (
//	nplSignature = []byte{0x7E, 0x7E}
//)
//
//const (
//	NPL_HEADER_LEN = 15
//	NPH_RESULT     = 0
//	NPH_HEADER_LEN = 10
//
//	// NPH service types
//	NPH_SRV_GENERIC_CONTROLS = 0
//	NPH_SRV_NAVDATA          = 1
//	NPH_SRV_FILE_TRANSFER    = 3
//	NPH_SRV_CLIENT_LIST      = 4
//	NPH_SRV_EXTERNAL_DEVICE  = 5
//	NPH_SRV_DEBUG            = 6
//
//	// NPH_SRV_GENERIC_CONTROLS packets
//	NPH_SGC_RESULT            = NPH_RESULT
//	NPH_SGC_CONN_REQUEST      = 100
//	NPH_SGC_CONN_AUTH_STRING  = 101
//	NPH_SGC_SERVICE_REQUEST   = 110
//	NPH_SGC_SERVICES_REQUEST  = 111
//	NPH_SGC_SERVICES          = 112
//	NPH_SGC_PEER_DESC_REQUEST = 120
//	NPH_SGC_PEER_DESC         = 121
//
//	// PHOTO
//	NPH_GET_PAR_PHOTO  = 141
//	NPH_SET_PAR_PHOTO  = 142
//	NPH_PAR_PHOTO_SD   = 143
//	NPH_PAR_PHOTO_GPRS = 144
//	NPH_GET_PHOTO      = 145
//
//	// DIGITAL OUTPUT
//	NPH_SET_PRDO = 150
//	NPH_GET_PRDO = 151
//	NPH_PRDO     = 152
//
//	// MODE ALARM
//	NPH_SET_MODALARM = 160
//	NPH_GET_MODALARM = 161
//	NPH_MODALARM     = 162
//
//	// Program cmd for connect server
//	NPH_SET_PRIA = 170
//	NPH_GET_PRIA = 171
//	NPH_PRIA     = 172
//
//	// Program parametrs for navigation system
//	NPH_SET_PRNAV = 175
//	NPH_GET_PRNAV = 176
//	NPH_PRNAV     = 177
//
//	// Command for read firmware
//	NPH_SET_LOADFIRM = 180
//
//	// Get information about navigator
//	NPH_GET_INFO = 185
//	NPH_INFO     = 186
//
//	// Get balance on SIM card
//	NPH_GET_BALANCE = 190
//	NPH_BALANCE     = 191
//
//	// Command for set current time
//	NPH_SET_CURTIME = 195
//
//	// Command for Aotoinformer
//	NPH_SET_ROUTE_AUTOINFORMER = 200
//	NPH_GET_ROUTE_AUTOINFORMER = 201
//	NPH_ROUTE_AUTOINFORMER     = 202
//
//	// Command for reset internal state
//	NPH_RESET_INT_STATE = 205
//	// Bit mask for reset state
//	MASK_RESET_VOICE = 1
//	MASK_RESET_SOS   = 2
//
//	// Command for IMSI SIM code
//	NPH_GET_SIM_IMSI = 210
//	NPH_SIM_IMSI     = 211
//
//	// Command for set parameters in roaming
//	NPH_GPRS_ROMING = 215
//
//	// Command for get min navigation
//	NPH_GET_NAVINFO = 220
//
//	// Programm cmd for connect server extended
//	NPH_SET_PRIA_EXT = 230
//	NPH_GET_PRIA_EXT = 231
//	NPH_PRIA_EXT     = 232
//
//	// Command for set parameters in roaming Extended
//	NPH_GPRS_ROMING_EXT = 235
//
//	// Programm cmd enable second server
//	NPH_SET_SECSERVER = 235
//
//	// SMS (This command may receive only from sms channel)
//	NPH_SET_PRIA_EXT_SMS    = 1000
//	NPH_SET_PRNAV_SMS       = 1005
//	NPH_SET_LOADFIRM_SMS    = 1010
//	NPH_GET_INFO_SMS        = 1015
//	NPH_GET_BALANCE_SMS     = 1020
//	NPH_RESET_SMS           = 1025
//	NPH_ADD_TEL_SMS         = 1030
//	NPH_DEL_TEL_SMS         = 1031
//	NPH_GPRS_ROMING_EXT_SMS = 1035
//	NPH_GET_NAVINFO_SMS     = 1040
//	NPH_SET_SECSERVER_SMS   = 1050
//
//	NPH_SET_PRIA_SMS    = 1200
//	NPH_GPRS_ROMING_SMS = 1205
//
//	// NPH_SRV_NAVDATA packets
//	NPH_SND_RESULT   = NPH_RESULT
//	NPH_SND_HISTORY  = 100
//	NPH_SND_REALTIME = 101
//
//	// NPH_SRV_CLIENT_LIST packets
//	NPH_SCL_RESULT                = NPH_RESULT
//	NPH_SCL_CLIENT_LIST_REQUEST   = 100
//	NPH_SCL_CLIENT_LIST           = 101
//	NPH_SCL_CLIENT_STATUS_REQUEST = 102
//
//	// NPH_SRV_EXTERNAL_DEVICE packets
//	NPH_SED_DEVICE_TITLE_DATA = 100
//	NPH_SED_DEVICE_DATA       = 101
//	NPH_SED_DEVICE_RESULT     = 102
//	// service errors
//	NPH_RESULT_SERVICE_NOT_SUPPORTED      = 100
//	NPH_RESULT_SERVICE_NOT_ALLOWED        = 101
//	NPH_RESULT_SERVICE_NOT_AVIALABLE      = 102
//	NPH_RESULT_SERVICE_BUSY               = 103
//	NPH_RESULT_SERVICE_NOT_CONSECUTION    = 104
//	NPH_RESULT_SERVICE_NOT_DEVICE_ADDRESS = 105
//
//	// packet errors
//	NPH_RESULT_PACKET_NOT_SUPPORTED     = 200
//	NPH_RESULT_PACKET_INVALID_SIZE      = 201
//	NPH_RESULT_PACKET_INVALID_FORMAT    = 202
//	NPH_RESULT_PACKET_INVALID_PARAMETER = 203
//	NPH_RESULT_PACKET_UNEXPECTED        = 204
//
//	// connection errors
//	NPH_RESULT_PROTO_VER_NOT_SUPPORTED   = 300
//	NPH_RESULT_CLIENT_NOT_REGISTERED     = 301
//	NPH_RESULT_CLIENT_TYPE_NOT_SUPPORTED = 302
//	NPH_RESULT_CLIENT_AUTH_FAILED        = 303
//	NPH_RESULT_INVALID_ADDRESS           = 304
//	NPH_RESULT_CLIENT_ALREADY_REGISTERED = 305
//
//	// database errors
//	NPH_RESULT_DATABASE_QUERY_FAILED = 400
//	NPH_RESULT_DATABASE_DOWN         = 401
//
//	// message errors
//	NPH_RESULT_CODEC_NOT_SUPPORTED = 500
//)
//
//var navDataLength = map[uint8]int{
//	0:  26,
//	2:  26,
//	3:  14,
//	4:  15,
//	5:  6,
//	6:  9,
//	7:  1,
//	8:  6,
//	9:  40,
//	10: 37,
//	12: 5,
//	13: 13,
//	14: 15,
//	15: 50,
//	16: 8,
//	17: 46,
//	18: 50,
//	19: 40,
//	20: 8,
//	21: 180,
//	22: 24,
//	23: 16,
//	24: 4,
//	25: 41,
//}
//
//var okResult = make([]byte, 4)
//var okResultExt = make([]byte, 8)
//var errResult = make([]byte, 4)
//var nphResultType = make([]byte, 2)
//var extResultType = make([]byte, 2)
//var NPH_NO_NEED_REPLY uint16 = 0
//
//func init() {
//	body := new(bytes.Buffer)
//	binary.Write(body, binary.LittleEndian, uint32(NPH_RESULT_SERVICE_BUSY))
//	errResult = body.Bytes()
//	binary.LittleEndian.PutUint16(extResultType, uint16(NPH_SED_DEVICE_RESULT))
//}
//
//func crc16(bs []byte) (crc uint16) {
//	l := len(bs)
//	crc = 0xFFFF
//	for i := 0; i < l; i++ {
//		crc = (crc >> 8) ^ crc16tab[(crc&0xff)^uint16(bs[i])]
//	}
//	return
//}
//
//var crc16tab = [256]uint16{
//	0x0000, 0xC0C1, 0xC181, 0x0140, 0xC301, 0x03C0, 0x0280, 0xC241,
//	0xC601, 0x06C0, 0x0780, 0xC741, 0x0500, 0xC5C1, 0xC481, 0x0440,
//	0xCC01, 0x0CC0, 0x0D80, 0xCD41, 0x0F00, 0xCFC1, 0xCE81, 0x0E40,
//	0x0A00, 0xCAC1, 0xCB81, 0x0B40, 0xC901, 0x09C0, 0x0880, 0xC841,
//	0xD801, 0x18C0, 0x1980, 0xD941, 0x1B00, 0xDBC1, 0xDA81, 0x1A40,
//	0x1E00, 0xDEC1, 0xDF81, 0x1F40, 0xDD01, 0x1DC0, 0x1C80, 0xDC41,
//	0x1400, 0xD4C1, 0xD581, 0x1540, 0xD701, 0x17C0, 0x1680, 0xD641,
//	0xD201, 0x12C0, 0x1380, 0xD341, 0x1100, 0xD1C1, 0xD081, 0x1040,
//	0xF001, 0x30C0, 0x3180, 0xF141, 0x3300, 0xF3C1, 0xF281, 0x3240,
//	0x3600, 0xF6C1, 0xF781, 0x3740, 0xF501, 0x35C0, 0x3480, 0xF441,
//	0x3C00, 0xFCC1, 0xFD81, 0x3D40, 0xFF01, 0x3FC0, 0x3E80, 0xFE41,
//	0xFA01, 0x3AC0, 0x3B80, 0xFB41, 0x3900, 0xF9C1, 0xF881, 0x3840,
//	0x2800, 0xE8C1, 0xE981, 0x2940, 0xEB01, 0x2BC0, 0x2A80, 0xEA41,
//	0xEE01, 0x2EC0, 0x2F80, 0xEF41, 0x2D00, 0xEDC1, 0xEC81, 0x2C40,
//	0xE401, 0x24C0, 0x2580, 0xE541, 0x2700, 0xE7C1, 0xE681, 0x2640,
//	0x2200, 0xE2C1, 0xE381, 0x2340, 0xE101, 0x21C0, 0x2080, 0xE041,
//	0xA001, 0x60C0, 0x6180, 0xA141, 0x6300, 0xA3C1, 0xA281, 0x6240,
//	0x6600, 0xA6C1, 0xA781, 0x6740, 0xA501, 0x65C0, 0x6480, 0xA441,
//	0x6C00, 0xACC1, 0xAD81, 0x6D40, 0xAF01, 0x6FC0, 0x6E80, 0xAE41,
//	0xAA01, 0x6AC0, 0x6B80, 0xAB41, 0x6900, 0xA9C1, 0xA881, 0x6840,
//	0x7800, 0xB8C1, 0xB981, 0x7940, 0xBB01, 0x7BC0, 0x7A80, 0xBA41,
//	0xBE01, 0x7EC0, 0x7F80, 0xBF41, 0x7D00, 0xBDC1, 0xBC81, 0x7C40,
//	0xB401, 0x74C0, 0x7580, 0xB541, 0x7700, 0xB7C1, 0xB681, 0x7640,
//	0x7200, 0xB2C1, 0xB381, 0x7340, 0xB101, 0x71C0, 0x7080, 0xB041,
//	0x5000, 0x90C1, 0x9181, 0x5140, 0x9301, 0x53C0, 0x5280, 0x9241,
//	0x9601, 0x56C0, 0x5780, 0x9741, 0x5500, 0x95C1, 0x9481, 0x5440,
//	0x9C01, 0x5CC0, 0x5D80, 0x9D41, 0x5F00, 0x9FC1, 0x9E81, 0x5E40,
//	0x5A00, 0x9AC1, 0x9B81, 0x5B40, 0x9901, 0x59C0, 0x5880, 0x9841,
//	0x8801, 0x48C0, 0x4980, 0x8941, 0x4B00, 0x8BC1, 0x8A81, 0x4A40,
//	0x4E00, 0x8EC1, 0x8F81, 0x4F40, 0x8D01, 0x4DC0, 0x4C80, 0x8C41,
//	0x4400, 0x84C1, 0x8581, 0x4540, 0x8701, 0x47C0, 0x4680, 0x8641,
//	0x8201, 0x42C0, 0x4380, 0x8341, 0x4100, 0x81C1, 0x8081, 0x4040}
