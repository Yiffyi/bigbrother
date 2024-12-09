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

// export cgo_authenticate
func cgo_authenticate(pamh *C.pam_handle_t) C.int {
	var pamUsername *C.char
	status := C.pam_get_user(pamh, &pamUsername, nil)

	dumpPAMItems(&PAMHandle{
		pamh:   pamh,
		status: C.PAM_SUCCESS,
	})

	if status != C.PAM_SUCCESS {
		return C.PAM_SERVICE_ERR
	} else {
		return C.PAM_SUCCESS
	}
}

//export cgo_open_session
func cgo_open_session(pamh *C.pam_handle_t) C.int {
	var pamUsername *C.char
	status := C.pam_get_user(pamh, &pamUsername, nil)

	dumpPAMItems(&PAMHandle{
		pamh:   pamh,
		status: C.PAM_SUCCESS,
	})

	if status != C.PAM_SUCCESS {
		return C.PAM_SERVICE_ERR
	} else {
		return C.PAM_SUCCESS
	}
}
