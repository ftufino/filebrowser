package fbhttp

import (
	"encoding/json"
	"net/http"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// effectiveSMTP returns the SMTP configuration to use for a user: the personal
// one when set, the global one otherwise. The bool is false if none is usable.
func effectiveSMTP(global settings.SMTPSettings, personal *users.SMTP) (settings.SMTPSettings, bool) {
	cfg := global
	if personal != nil && personal.Host != "" {
		cfg = settings.SMTPSettings{
			Host:            personal.Host,
			Port:            personal.Port,
			Username:        personal.Username,
			Password:        personal.Password,
			From:            personal.From,
			MaxAttachmentMB: global.MaxAttachmentMB,
		}
	}
	return cfg, cfg.Host != "" && cfg.From != ""
}

type userSMTPResponse struct {
	// Configured is true when an SMTP server (personal or global) can be used.
	Configured bool              `json:"configured"`
	Personal   *smtpSettingsData `json:"personal"`
}

var smtpGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	_, ok := effectiveSMTP(d.settings.SMTP, d.user.SMTP)
	res := userSMTPResponse{Configured: ok}
	if p := d.user.SMTP; p != nil && p.Host != "" {
		res.Personal = &smtpSettingsData{
			Host:        p.Host,
			Port:        p.Port,
			Username:    p.Username,
			From:        p.From,
			HasPassword: p.Password != "",
		}
	}
	return renderJSON(w, r, res)
})

var smtpPutHandler = withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Share {
		return http.StatusForbidden, nil
	}
	req := &smtpSettingsData{}
	if err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxEmailRequestSize)).Decode(req); err != nil {
		return http.StatusBadRequest, err
	}

	var current settings.SMTPSettings
	if d.user.SMTP != nil {
		current = settings.SMTPSettings{Password: d.user.SMTP.Password}
	}
	next, err := updateSMTPSettings(current, req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if next.Host == "" {
		d.user.SMTP = nil
	} else {
		d.user.SMTP = &users.SMTP{
			Host:     next.Host,
			Port:     next.Port,
			Username: next.Username,
			Password: next.Password,
			From:     next.From,
		}
	}
	if err := d.store.Users.Update(d.user, "SMTP"); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusNoContent, nil
})
