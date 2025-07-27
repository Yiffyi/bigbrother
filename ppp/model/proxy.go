package model

import "time"

type UpdateProxyConfigRequest struct {
	ProxyType string

	ConfigFile []byte
	Restart    bool
}

type ReportStatusRequest struct {
	ProxyType string
	Running   bool

	SystemTime     time.Time
	CPUPercent     float64
	MemUsedPercent float64
}
