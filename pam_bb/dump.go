package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func _pamDebugDump(h *PAMHandle) {
	if !viper.GetBool("log.pam_debug_dump") {
		return
	}

	serviceName, _ := h.pam_get_item_string(PAM_SERVICE)
	user, _ := h.pam_get_item_string(PAM_USER)
	tty, _ := h.pam_get_item_string(PAM_TTY)
	ruser, _ := h.pam_get_item_string(PAM_RUSER)
	rhost, _ := h.pam_get_item_string(PAM_RHOST)
	token, _ := h.pam_get_item_string(PAM_AUTHTOK)

	log.Info().
		Str("service", serviceName).
		Str("user", user).
		Str("tty", tty).
		Str("ruser", ruser).
		Str("rhost", rhost).
		Str("authtok", token).
		Msg("dumping items using pam_get_item")
}
