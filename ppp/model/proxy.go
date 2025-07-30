package model

import "time"

type UpdateServerConfigRequest struct {
	ServerType string

	Config  []byte
	Restart bool
}

type ReportStatusRequest struct {
	ServerType string
	Running    bool

	SystemTime     time.Time
	CPUPercent     float64
	MemUsedPercent float64
}
