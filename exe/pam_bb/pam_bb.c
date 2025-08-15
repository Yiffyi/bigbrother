#include <stdlib.h>
#include <syslog.h>

#include "pam_bb.h"

/* Authentication API's */
int pam_sm_authenticate(pam_handle_t *pamh, int flags,
                        int argc, const char **argv)
{
    syslog(LOG_DEBUG, "pam_sm_authenticate");
    return bb_cgo_authenticate(pamh);
}

int pam_sm_setcred(pam_handle_t *pamh, int flags,
                   int argc, const char **argv)
{
    syslog(LOG_DEBUG, "pam_sm_setcred");
    return PAM_IGNORE;
}

/* Account Management API's */
int pam_sm_acct_mgmt(pam_handle_t *pamh, int flags,
                     int argc, const char **argv)
{
    syslog(LOG_DEBUG, "pam_sm_acct_mgmt");
    return PAM_IGNORE;
}

/* Session Management API's */
int pam_sm_open_session(pam_handle_t *pamh, int flags,
                        int argc, const char **argv)
{
    syslog(LOG_DEBUG, "pam_sm_open_session");
    return bb_cgo_open_session(pamh);
}

int pam_sm_close_session(pam_handle_t *pamh, int flags,
                         int argc, const char **argv)
{
    syslog(LOG_DEBUG, "pam_sm_close_session");
    return PAM_IGNORE;
}

/* Password Management API's */
int pam_sm_chauthtok(pam_handle_t *pamh, int flags,
                     int argc, const char **argv)
{
    syslog(LOG_DEBUG, "pam_sm_chauthtok");
    return PAM_IGNORE;
}

int bb_c_conv(const pam_handle_t *pamh, int msg_style, const char* msg)
{
    const struct pam_conv *conv;
    int status = pam_get_item(pamh, PAM_CONV, (const void**)&conv);
    if (status != PAM_SUCCESS) return status;

    const struct pam_message pmsg = {
        .msg_style = msg_style,
        .msg = msg
    };
    const struct pam_message *pmsgs = &pmsg;
    struct pam_response *resp = 0;
    status = conv->conv(1, &pmsgs, &resp, conv->appdata_ptr);

    if (resp) {
        if (resp->resp) free(resp->resp); // we don't read this
        free(resp);
    }
    return status;
}