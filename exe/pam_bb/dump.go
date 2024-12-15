package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func _pamDebugDump(from string, h *PAMHandle) map[string]string {

	serviceName, _ := h.pam_get_item_string(PAM_SERVICE)
	user, _ := h.pam_get_item_string(PAM_USER)
	tty, _ := h.pam_get_item_string(PAM_TTY)
	ruser, _ := h.pam_get_item_string(PAM_RUSER)
	rhost, _ := h.pam_get_item_string(PAM_RHOST)
	authtok, _ := h.pam_get_item_string(PAM_AUTHTOK)

	if viper.GetBool("log.pam_debug_dump") {
		log.Info().
			Str("from", from).
			Str("service", serviceName).
			Str("user", user).
			Str("tty", tty).
			Str("ruser", ruser).
			Str("rhost", rhost).
			Str("authtok", authtok).
			Msg("dumping items using pam_get_item")
	}

	return map[string]string{
		"service": serviceName,
		"user":    user,
		"tty":     tty,
		"ruser":   ruser,
		"rhost":   rhost,
		"authtok": authtok,
	}
}
