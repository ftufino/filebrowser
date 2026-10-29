package fbhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
)

type settingsData struct {
	Signup                bool                  `json:"signup"`
	HideLoginButton       bool                  `json:"hideLoginButton"`
	CreateUserDir         bool                  `json:"createUserDir"`
	MinimumPasswordLength uint                  `json:"minimumPasswordLength"`
	UserHomeBasePath      string                `json:"userHomeBasePath"`
	Defaults              settings.UserDefaults `json:"defaults"`
	AuthMethod            settings.AuthMethod   `json:"authMethod"`
	Rules                 []rules.Rule          `json:"rules"`
	Branding              settings.Branding     `json:"branding"`
	SMTP                  *smtpSettingsData     `json:"smtp,omitempty"`
	Tus                   settings.Tus          `json:"tus"`
	Shell                 []string              `json:"shell"`
	Commands              map[string][]string   `json:"commands"`
}

type smtpSettingsData struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	From        string `json:"from"`
	HasPassword bool   `json:"hasPassword"`
}

func smtpSettingsResponse(s settings.SMTPSettings) *smtpSettingsData {
	return &smtpSettingsData{
		Host:        s.Host,
		Port:        s.Port,
		Username:    s.Username,
		From:        s.From,
		HasPassword: s.Password != "",
	}
}

func updateSMTPSettings(current settings.SMTPSettings, req *smtpSettingsData) (settings.SMTPSettings, error) {
	if req == nil {
		return current, nil
	}
	if strings.TrimSpace(req.Host) == "" {
		return settings.SMTPSettings{}, nil
	}

	password := req.Password
	if password == "" && req.Username != "" {
		password = current.Password
	}
	port := req.Port
	if port == 0 {
		port = 587
	}
	next := settings.SMTPSettings{
		Host:     strings.TrimSpace(req.Host),
		Port:     port,
		Username: strings.TrimSpace(req.Username),
		Password: password,
		From:     strings.TrimSpace(req.From),
	}

	if strings.ContainsAny(next.Host, "\r\n\t /") {
		return current, fmt.Errorf("invalid SMTP host")
	}
	if next.Port < 1 || next.Port > 65535 {
		return current, fmt.Errorf("invalid SMTP port")
	}
	if (next.Username == "") != (next.Password == "") {
		return current, fmt.Errorf("SMTP username and password must both be set or both be empty")
	}
	if _, err := mail.ParseAddress(next.From); err != nil {
		return current, fmt.Errorf("invalid SMTP sender address: %w", err)
	}

	return next, nil
}

var settingsGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	data := &settingsData{
		Signup:                d.settings.Signup,
		HideLoginButton:       d.settings.HideLoginButton,
		CreateUserDir:         d.settings.CreateUserDir,
		MinimumPasswordLength: d.settings.MinimumPasswordLength,
		UserHomeBasePath:      d.settings.UserHomeBasePath,
		Defaults:              d.settings.Defaults,
		AuthMethod:            d.settings.AuthMethod,
		Rules:                 d.settings.Rules,
		Branding:              d.settings.Branding,
		SMTP:                  smtpSettingsResponse(d.settings.SMTP),
		Tus:                   d.settings.Tus,
		Shell:                 d.settings.Shell,
		Commands:              d.settings.Commands,
	}

	return renderJSON(w, r, data)
})

var settingsPutHandler = withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	req := &settingsData{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	d.settings.Signup = req.Signup
	d.settings.CreateUserDir = req.CreateUserDir
	d.settings.MinimumPasswordLength = req.MinimumPasswordLength
	d.settings.UserHomeBasePath = req.UserHomeBasePath
	d.settings.Defaults = req.Defaults
	d.settings.Rules = req.Rules
	d.settings.Branding = req.Branding
	smtpSettings, err := updateSMTPSettings(d.settings.SMTP, req.SMTP)
	if err != nil {
		return http.StatusBadRequest, err
	}
	d.settings.SMTP = smtpSettings
	d.settings.Tus = req.Tus
	d.settings.Shell = req.Shell
	d.settings.Commands = req.Commands
	d.settings.HideLoginButton = req.HideLoginButton

	err = d.store.Settings.Save(d.settings)
	return errToStatus(err), err
})
