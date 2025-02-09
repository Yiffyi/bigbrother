package honeypot

import (
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"gorm.io/gorm"
)

type AuthAttempt struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time

	Username   string
	Password   string
	RemoteIP   string
	RemotePort int
}

type CounterAttackAttempt struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time

	Username string
	Password string

	RemoteIP      string
	RemotePort    int
	ServerVersion string

	AuthSuccess bool
	ExecSuccess bool

	ExecUname     string
	ExecProcesses string
	ExecLsRoot    string
	ExecFree      string
	ExecDf        string
}

func OpenDatabase(v *viper.Viper) (db *gorm.DB, err error) {

	dsn := v.GetString("db")
	// github.com/mattn/go-sqlite3
	db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&AuthAttempt{}, &CounterAttackAttempt{})
	if err != nil {
		return db, err
	}

	return
}
