#include <syslog.h>

#include "pam_bb.h"

/* Authentication API's */
int pam_sm_authenticate(pam_handle_t *pamh, int flags,
                        int argc, const char **argv)
{
    syslog(LOG_DEBUG, "pam_sm_authenticate");
    return go_authenticate(pamh);
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
    syslog(LOG_DEBUG, "pam_sm_open_session");
    return PAM_IGNORE;
}

/* Session Management API's */
int pam_sm_open_session(pam_handle_t *pamh, int flags,
                        int argc, const char **argv)
{
    syslog(LOG_DEBUG, "pam_sm_open_session");
    return PAM_IGNORE;
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