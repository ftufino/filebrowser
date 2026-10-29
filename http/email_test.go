package fbhttp

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestBuildAttachmentMessage(t *testing.T) {
	from, err := mail.ParseAddress("files@example.com")
	if err != nil {
		t.Fatal(err)
	}
	to, err := mail.ParseAddress("recipient@example.com")
	if err != nil {
		t.Fatal(err)
	}

	wantContent := []byte("email attachment content")
	message, err := buildAttachmentMessage(from, to, "résumé.txt", wantContent)
	if err != nil {
		t.Fatalf("build message: %v", err)
	}

	parsed, err := mail.ReadMessage(strings.NewReader(string(message)))
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	_, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("parse message content type: %v", err)
	}
	reader := multipart.NewReader(parsed.Body, params["boundary"])
	if _, err := reader.NextPart(); err != nil {
		t.Fatalf("read email body: %v", err)
	}
	attachment, err := reader.NextPart()
	if err != nil {
		t.Fatalf("read attachment part: %v", err)
	}
	if got := attachment.FileName(); got != "résumé.txt" {
		t.Fatalf("attachment filename = %q, want résumé.txt", got)
	}
	encoded, err := io.ReadAll(attachment)
	if err != nil {
		t.Fatalf("read encoded attachment: %v", err)
	}
	gotContent, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(string(encoded))))
	if err != nil {
		t.Fatalf("decode attachment: %v", err)
	}
	if string(gotContent) != string(wantContent) {
		t.Fatalf("attachment = %q, want %q", gotContent, wantContent)
	}
}

func TestEmailHandlerRejectsDirectoriesAndOversizedFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "folder"), 0o755); err != nil {
		t.Fatal(err)
	}
	large, err := os.Create(filepath.Join(root, "large.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if err := large.Truncate(maxEmailAttachmentSize + 1); err != nil {
		t.Fatal(err)
	}
	if err := large.Close(); err != nil {
		t.Fatal(err)
	}

	key := []byte("test-signing-key")
	perm := users.Permissions{Share: true, Download: true}
	st := scopedUserStorage(t, root, perm, key)
	appSettings, err := st.Settings.Get()
	if err != nil {
		t.Fatal(err)
	}
	appSettings.SMTP = settings.SMTPSettings{
		Host: "localhost",
		Port: 587,
		From: "sender@example.com",
	}
	if err := st.Settings.Save(appSettings); err != nil {
		t.Fatal(err)
	}

	handler := handle(emailPostHandler, "", st, &settings.Server{Root: root})
	signed := signToken(t, perm, key)
	for _, tc := range []struct {
		name string
		path string
		want int
	}{
		{name: "directory", path: "/folder/", want: http.StatusBadRequest},
		{name: "oversized file", path: "/large.bin", want: http.StatusRequestEntityTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(`{"to":"recipient@example.com"}`))
			req.Header.Set("X-Auth", signed)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			if recorder.Code != tc.want {
				t.Fatalf("status = %d, want %d; body=%q", recorder.Code, tc.want, recorder.Body.String())
			}
		})
	}
}

func TestSMTPSettingsResponseMasksPasswordAndUpdatePreservesIt(t *testing.T) {
	current := settings.SMTPSettings{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "secret",
		From:     "sender@example.com",
	}
	response := smtpSettingsResponse(current)
	if response.Password != "" || !response.HasPassword {
		t.Fatalf("SMTP password response leaked or omitted saved-password state: %#v", response)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("encode SMTP settings response: %v", err)
	}
	if strings.Contains(string(encoded), current.Password) {
		t.Fatal("SMTP settings response JSON contains the saved password")
	}

	updated, err := updateSMTPSettings(current, &smtpSettingsData{
		Host:     current.Host,
		Port:     current.Port,
		Username: current.Username,
		From:     current.From,
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if updated.Password != current.Password {
		t.Fatal("blank password did not preserve the saved SMTP password")
	}
}
