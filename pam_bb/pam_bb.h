#ifndef BIGBROTHER_PAM_H
#define BIGBROTHER_PAM_H
#include <security/pam_modules.h>
#include <security/pam_ext.h>

extern int cgo_authenticate(pam_handle_t *pamh);
extern int cgo_open_session(pam_handle_t *pamh);

#endif  // BIGBROTHER_PAM_H