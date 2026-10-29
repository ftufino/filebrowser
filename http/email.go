package fbhttp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/settings"
)

const (
	maxEmailAttachmentSize = 25 << 20
	maxEmailRequestSize    = 4 << 10
)

type emailRequest struct {
	To string `json:"to"`
}

var emailPostHandler = withPermShare(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.Check(r.URL.Path) {
		return http.StatusForbidden, nil
	}

	if d.settings.SMTP.Host == "" || d.settings.SMTP.From == "" {
		return http.StatusServiceUnavailable, fmt.Errorf("email sharing is not configured")
	}

	if r.Body == nil {
		return http.StatusBadRequest, fmt.Errorf("missing email request body")
	}
	defer r.Body.Close()
	requestBody, err := io.ReadAll(io.LimitReader(r.Body, maxEmailRequestSize+1))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("read email request: %w", err)
	}
	if len(requestBody) > maxEmailRequestSize {
		return http.StatusRequestEntityTooLarge, fmt.Errorf("email request is too large")
	}
	var req emailRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid email request: %w", err)
	}

	recipient, err := mail.ParseAddress(strings.TrimSpace(req.To))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid recipient address: %w", err)
	}

	info, err := d.user.Fs.Stat(r.URL.Path)
	if err != nil {
		return errToStatus(err), err
	}
	if !info.Mode().IsRegular() {
		return http.StatusBadRequest, fmt.Errorf("only regular files can be emailed")
	}
	if info.Size() > maxEmailAttachmentSize {
		return http.StatusRequestEntityTooLarge, fmt.Errorf("attachment exceeds the 25 MiB limit")
	}

	file, err := d.user.Fs.Open(r.URL.Path)
	if err != nil {
		return errToStatus(err), err
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxEmailAttachmentSize+1))
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if len(content) > maxEmailAttachmentSize {
		return http.StatusRequestEntityTooLarge, fmt.Errorf("attachment exceeds the 25 MiB limit")
	}

	if err := sendEmailAttachment(d.settings.SMTP, recipient, filepath.Base(r.URL.Path), content); err != nil {
		return http.StatusBadGateway, fmt.Errorf("failed to send email: %w", err)
	}

	return http.StatusNoContent, nil
})

func sendEmailAttachment(config settings.SMTPSettings, recipient *mail.Address, filename string, content []byte) error {
	from, err := mail.ParseAddress(config.From)
	if err != nil {
		return fmt.Errorf("invalid SMTP sender address: %w", err)
	}
	if config.Host == "" || config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("invalid SMTP server configuration")
	}
	if (config.Username == "") != (config.Password == "") {
		return fmt.Errorf("SMTP username and password must both be set or both be empty")
	}

	message, err := buildAttachmentMessage(from, recipient, filename, content)
	if err != nil {
		return err
	}

	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	if err := conn.SetDeadline(time.Now().Add(2 * time.Minute)); err != nil {
		_ = conn.Close()
		return fmt.Errorf("set SMTP connection deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("initialize SMTP client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("SMTP server does not support required STARTTLS")
	}
	if err := client.StartTLS(&tls.Config{
		ServerName: config.Host,
		MinVersion: tls.VersionTLS12,
	}); err != nil {
		return fmt.Errorf("start SMTP TLS: %w", err)
	}
	if config.Username != "" {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}
	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(recipient.Address); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP message: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finish SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("close SMTP session: %w", err)
	}

	return nil
}

func buildAttachmentMessage(from, recipient *mail.Address, filename string, content []byte) ([]byte, error) {
	var message bytes.Buffer
	writer := multipart.NewWriter(&message)
	fmt.Fprintf(&message, "From: %s\r\n", from.String())
	fmt.Fprintf(&message, "To: %s\r\n", recipient.String())
	fmt.Fprintf(&message, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", "File shared with you"))
	fmt.Fprintf(&message, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(&message, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&message, "Content-Type: multipart/mixed; boundary=%q\r\n\r\n", writer.Boundary())

	bodyHeader := make(textproto.MIMEHeader)
	bodyHeader.Set("Content-Type", "text/plain; charset=utf-8")
	bodyPart, err := writer.CreatePart(bodyHeader)
	if err != nil {
		return nil, fmt.Errorf("create email body: %w", err)
	}
	if _, err := io.WriteString(bodyPart, "The attached file was shared with you using File Browser.\r\n"); err != nil {
		return nil, fmt.Errorf("write email body: %w", err)
	}

	mediaType := mime.TypeByExtension(filepath.Ext(filename))
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	attachmentHeader := make(textproto.MIMEHeader)
	attachmentHeader.Set("Content-Type", mediaType)
	attachmentHeader.Set("Content-Transfer-Encoding", "base64")
	attachmentHeader.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": filename,
	}))
	attachmentPart, err := writer.CreatePart(attachmentHeader)
	if err != nil {
		return nil, fmt.Errorf("create email attachment: %w", err)
	}
	lineWriter := &base64LineWriter{writer: attachmentPart}
	encoder := base64.NewEncoder(base64.StdEncoding, lineWriter)
	if _, err := encoder.Write(content); err != nil {
		return nil, fmt.Errorf("encode email attachment: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("finish email attachment: %w", err)
	}
	if lineWriter.lineLength > 0 {
		if _, err := io.WriteString(attachmentPart, "\r\n"); err != nil {
			return nil, fmt.Errorf("finish email attachment line: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("finish email message: %w", err)
	}

	return message.Bytes(), nil
}

type base64LineWriter struct {
	writer     io.Writer
	lineLength int
}

func (w *base64LineWriter) Write(data []byte) (int, error) {
	written := 0
	for len(data) > 0 {
		available := 76 - w.lineLength
		chunk := data
		if len(chunk) > available {
			chunk = chunk[:available]
		}
		n, err := w.writer.Write(chunk)
		written += n
		w.lineLength += n
		data = data[n:]
		if err != nil {
			return written, err
		}
		if n != len(chunk) {
			return written, io.ErrShortWrite
		}
		if w.lineLength == 76 {
			if _, err := io.WriteString(w.writer, "\r\n"); err != nil {
				return written, err
			}
			w.lineLength = 0
		}
	}
	return written, nil
}
