package main

import (
	"github.com/rs/zerolog/log"
)

func dumpPAMItems(h *PAMHandle) {
	serviceName, err := h.pam_get_item_string(PAM_SERVICE)
	if err != nil {
		panic(err)
	}

	user, err := h.pam_get_item_string(PAM_USER)

	if err != nil {
		panic(err)
	}

	tty, err := h.pam_get_item_string(PAM_TTY)
	if err != nil {
		panic(err)
	}

	ruser, err := h.pam_get_item_string(PAM_RUSER)
	if err != nil {
		panic(err)
	}

	rhost, err := h.pam_get_item_string(PAM_RHOST)
	if err != nil {
		panic(err)
	}

	token, err := h.pam_get_item_string(PAM_AUTHTOK)
	if err != nil {
		panic(err)
	}

	log.Info().
		Str("service", serviceName).
		Str("user", user).
		Str("tty", tty).
		Str("ruser", ruser).
		Str("rhost", rhost).
		Str("authtok", token).
		Msg("dumping items using pam_get_item")
}
