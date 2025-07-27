package model

import "time"

type UpdateProxyConfigRequest struct {
	DaemonType string

	ConfigFile []byte
	Restart    bool
}

type UpdateProxyConfigResponse struct {
	Error error
}

type ReportStatusRequest struct {
	DaemonType string
	Running    bool

	SystemTime     time.Time
	CPUPercent     float64
	MemUsedPercent float64
}
