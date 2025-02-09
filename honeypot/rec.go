package honeypot

import (
	"net"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func RecordAuthAttempt(db *gorm.DB, username, password string, remoteAddr net.Addr) error {
	a := AuthAttempt{
		Username: username,
		Password: password,
	}

	switch addr := remoteAddr.(type) {
	case *net.UDPAddr: // though this is impossible
		a.RemoteIP = addr.IP.String()
		a.RemotePort = addr.Port
	case *net.TCPAddr:
		a.RemoteIP = addr.IP.String()
		a.RemotePort = addr.Port
	}

	result := db.Create(&a)
	if err := result.Error; err != nil {
		log.Error().Err(err).Interface("authAttempt", a).Msg("could not insert AuthAttempt into DB")
		return err
	}

	return nil
}

func ShallAllowConnection(db *gorm.DB, username, password string, remoteAddr net.Addr) (bool, error) {
	a := AuthAttempt{}
	switch addr := remoteAddr.(type) {
	case *net.UDPAddr: // though this is impossible
		a.RemoteIP = addr.IP.String()
		a.RemotePort = addr.Port
	case *net.TCPAddr:
		a.RemoteIP = addr.IP.String()
		a.RemotePort = addr.Port
	}

	var count int64
	result := db.Model(&a).Where("username = ? AND password = ? AND remote_ip = ?", username, password, a.RemoteIP).Count(&count)
	if err := result.Error; err != nil {
		log.Error().Err(err).Str("username", username).Str("password", password).Str("remoteAddr", remoteAddr.String()).Msg("could not count AuthAttempt")
		return false, err
	}
	return count >= 2, nil
}
