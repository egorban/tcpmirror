package monitoring

import (
	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/egorban/influx/pkg/influx"
)

const (
	TypeNdtp     = "ndtp"
	TypeEgts     = "egts"
	TypeTerminal = "terminal"

	GetRedisPool = "getRedisPool"

	TerminalConn       = "terminalConn"
	TerminalFirstMsg   = "terminalFirstMsg"
	TerminalDisconnect = "terminalDisconnect"
	TerminalDropBuf    = "terminalDropBuf"
	TerminalProcMsg    = "terminalProcMsg"
	TerminalSend       = "terminalSend"
	TerminalRemoveExp  = "terminalRemoveExp"

	NdtpVisConn               = "NdtpVisConn"
	NdtpVisFirstMsg           = "NdtpVisFirstMsg"
	NdtpVisDisconnect         = "NdtpVisDisconnect"
	NdtpVisTerminalDisconnect = "NdtpVisTerminalDisconnect"
	NdtpVisProcTerminalMsg    = "NdtpVisProcTerminalMsg"
	NdtpVisProcMsgFrom        = "ndtpVisProcMsgFrom"
	NdtpVisSend               = "ndtpVisSend"
	NdtpVisDropBuf            = "ndtpVisDropBuf"
	NdtpVisMasterChannelFull  = "ndtpVisMasterChannelFull"
	NdtpVisGetOld             = "ndtpVisGetOld"

	EgtsVisConn            = "EgtsVisConn"
	EgtsVisSend            = "EgtsVisSend"
	EgtsVisProcMsgFrom     = "EgtsVisProcMsgFrom"
	EgtsVisDisconnect      = "EgtsVisDisconnect"
	EgtsVisProcTerminalMsg = "EgtsVisProcTerminalMsg"
	EgtsVisGetOld          = "EgtsVisGetOld"
)

func SendMetricInfo(options *util.Options, metricName string, typeSystem string) {
	if !options.MonEnable {
		return
	}

	tags := influx.Tags{
		"host":     host,
		"instance": util.Instance,
		"system":   typeSystem,
	}

	values := influx.Values{
		metricName: 1,
	}

	options.MonСlient.WritePoint(influx.NewPoint("info", tags, values))
}
