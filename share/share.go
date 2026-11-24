package share

type CreateBody struct {
	Password string `json:"password"`
	Expires  string `json:"expires"`
	Unit     string `json:"unit"`
	// DeleteFiles physically deletes the shared files when the share ends
	// (expiration or download limit reached).
	DeleteFiles bool `json:"deleteFiles"`
	// MaxDownloads ends the share after this many downloads. 0 means unlimited.
	MaxDownloads int `json:"maxDownloads"`
}

// Link is the information needed to build a shareable link.
type Link struct {
	Hash         string `json:"hash" storm:"id,index"`
	Path         string `json:"path" storm:"index"`
	UserID       uint   `json:"userID"`
	Expire       int64  `json:"expire"`
	PasswordHash string `json:"password_hash,omitempty"`
	// Token is a random value that will only be set when PasswordHash is set. It is
	// URL-Safe and is used to download links in password-protected shares via a
	// query arg.
	Token string `json:"token,omitempty"`
	// DeleteFiles physically deletes the shared path when the share ends.
	DeleteFiles bool `json:"deleteFiles,omitempty"`
	// MaxDownloads is the download limit (0 = unlimited).
	MaxDownloads int `json:"maxDownloads,omitempty"`
	// Downloads is the number of downloads performed so far.
	Downloads int `json:"downloads,omitempty"`
}

// Exhausted tells whether the download limit has been reached.
func (l *Link) Exhausted() bool {
	return l.MaxDownloads > 0 && l.Downloads >= l.MaxDownloads
}
