package model

import "time"

type ProgramType string

const (
	PROGRAM_TYPE_SINGBOX ProgramType = "sing-box"
	PROGRAM_TYPE_CLASH   ProgramType = "clash"
)

type UpdateServerConfigRequest struct {
	ServerType ProgramType

	Config  []byte
	Restart bool
}

type ReportStatusRequest struct {
	ServerType ProgramType
	Running    bool

	SystemTime     time.Time
	CPUPercent     float64
	MemUsedPercent float64
}
