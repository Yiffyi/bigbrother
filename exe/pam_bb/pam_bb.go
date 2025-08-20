package main

/*
#cgo LDFLAGS: -lpam
#include "pam_bb.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/yiffyi/bigbrother/misc"
	"github.com/yiffyi/bigbrother/push"
)

const (
	PAM_SERVICE     = C.PAM_SERVICE
	PAM_USER        = C.PAM_USER
	PAM_USER_PROMPT = C.PAM_USER_PROMPT
	PAM_TTY         = C.PAM_TTY
	PAM_RUSER       = C.PAM_RUSER
	PAM_RHOST       = C.PAM_RHOST
	PAM_AUTHTOK     = C.PAM_AUTHTOK
	// ...
)

var _loadConfigError error = nil
var _setupLogError error = nil
var _primaryPushChannel push.PushChannel = nil

type PAMHandle struct {
	pamh   *C.pam_handle_t
	status C.int
}

func (h *PAMHandle) toError() error {
	if h.status == C.PAM_SUCCESS {
		return nil
	} else {
		return errors.New(h.pam_strerror())
	}
}

func (h *PAMHandle) pam_strerror() string {
	var errmsg *C.char = C.pam_strerror(h.pamh, h.status)
	return C.GoString(errmsg)
}

func (h *PAMHandle) pam_get_item_string(itemType C.int) (string, error) {
	var item unsafe.Pointer
	h.status = C.pam_get_item(h.pamh, itemType, &item)
	err := h.toError()
	if err != nil {
		return "", fmt.Errorf("pam_get_item: %w", err)
	}

	return C.GoString((*C.char)(item)), nil
}

func _printImportantError(pamh *C.pam_handle_t) {

	if _loadConfigError != nil {
		C.bb_c_conv(pamh, C.PAM_ERROR_MSG, C.CString(fmt.Sprintf("BigBrother: [ERROR] could not load config, err=%s", _loadConfigError.Error())))
	}

	if _setupLogError != nil {
		C.bb_c_conv(pamh, C.PAM_ERROR_MSG, C.CString(fmt.Sprintf("BigBrother: [ERROR] could not setup log, err=%s", _setupLogError.Error())))
	}

}

//export bb_cgo_authenticate
func bb_cgo_authenticate(pamh *C.pam_handle_t) (status C.int) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Str("r", fmt.Sprintf("%v", r)).
				Msg("bb_cgo_authenticate: panic happened")

			C.bb_c_conv(pamh, C.PAM_ERROR_MSG, C.CString(fmt.Sprintf("BigBrother: [ERROR] bb_cgo_authenticate: panic happened, r=%v", r)))
			status = C.PAM_SERVICE_ERR
		}
	}()

	var pamUsername *C.char
	status = C.pam_get_user(pamh, &pamUsername, nil)

	h := &PAMHandle{
		pamh:   pamh,
		status: C.PAM_SUCCESS,
	}

	items := _pamDebugDump("bb_cgo_authenticate", h)
	_printImportantError(pamh)

	if status != C.PAM_SUCCESS {
		log.Error().
			Str("status", h.pam_strerror()).
			Msg("pam_get_user returned error")

		return C.PAM_SERVICE_ERR
	}

	_primaryPushChannel.NotifyPAMAuthenticate(items)

	return C.PAM_SUCCESS
}

//export bb_cgo_open_session
func bb_cgo_open_session(pamh *C.pam_handle_t) (status C.int) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Str("r", fmt.Sprintf("%v", r)).
				Msg("bb_cgo_open_session: panic happened")
			C.bb_c_conv(pamh, C.PAM_ERROR_MSG, C.CString(fmt.Sprintf("BigBrother: [ERROR] bb_cgo_open_session: panic happened, r=%v", r)))
			status = C.PAM_SERVICE_ERR
		}
	}()

	var pamUsername *C.char
	status = C.pam_get_user(pamh, &pamUsername, nil)

	h := &PAMHandle{
		pamh:   pamh,
		status: C.PAM_SUCCESS,
	}

	items := _pamDebugDump("bb_cgo_open_session", h)
	_printImportantError(pamh)

	if status != C.PAM_SUCCESS {
		log.Error().
			Str("status", h.pam_strerror()).
			Msg("pam_get_user returned error")

		return C.PAM_SERVICE_ERR
	}

	_primaryPushChannel.NotifyPAMOpenSession(items)

	status = C.PAM_SUCCESS
	return
}

func init() {
	viper.SetDefault("pam.push_channel", "tg0")

	_loadConfigError = misc.LoadConfig(nil)
	_setupLogError = misc.SetupLog()

	if _loadConfigError == nil {
		_primaryPushChannel, _ = push.GetPushChannel(viper.GetString("pam.push_channel"))
	}
}

// main is not called
func main() {}
