package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Username       string
	Email          string
	Tag            string
	EndpointGroups []string `gorm:"serializer:json"`

	SubsToken   string `gorm:"uniqueIndex"`
	SubsBeginAt time.Time
	SubsEndAt   time.Time
}

type InvoiceState uint

const (
	INVOICE_STATE_INVALID = 0
	INVOICE_STATE_OPEN    = 1
	INVOICE_STATE_SUCCESS = 2
	INVOICE_STATE_FINISH  = 3
	INVOICE_STATE_CLOSED  = 4
)

type Invoice struct {
	gorm.Model

	Amount int
	Days   int
	State  InvoiceState

	Claimed bool

	UserID uint
	User   User
}
