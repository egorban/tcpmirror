package monitoring

import (
	"github.com/ashirko/tcpmirror/internal/util"
	"github.com/egorban/influx/pkg/influx"
)

const (
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
)

func SendMetricInfo(options *util.Options, metricName string) {
	if !options.MonEnable {
		return
	}

	tags := influx.Tags{
		"host":     host,
		"instance": util.Instance,
		//"type":     typeMetric,
	}

	values := influx.Values{
		metricName: 1,
	}

	options.MonСlient.WritePoint(influx.NewPoint("info", tags, values))
}
