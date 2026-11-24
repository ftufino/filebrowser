package fbhttp

import (
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestEffectiveSMTP(t *testing.T) {
	global := settings.SMTPSettings{Host: "global", Port: 587, From: "g@example.com", MaxAttachmentMB: 10}
	personal := &users.SMTP{Host: "mine", Port: 465, From: "me@example.com"}

	if _, ok := effectiveSMTP(settings.SMTPSettings{}, nil); ok {
		t.Fatal("expected no SMTP when nothing is configured")
	}
	if cfg, ok := effectiveSMTP(global, nil); !ok || cfg.Host != "global" {
		t.Fatalf("expected global SMTP, got %+v", cfg)
	}
	cfg, ok := effectiveSMTP(global, personal)
	if !ok || cfg.Host != "mine" || cfg.MaxAttachmentMB != 10 {
		t.Fatalf("expected personal SMTP with global limit, got %+v", cfg)
	}
	if cfg, ok := effectiveSMTP(settings.SMTPSettings{}, personal); !ok || cfg.Host != "mine" {
		t.Fatalf("expected personal SMTP without global one, got %+v", cfg)
	}
}
