package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// TelegramAuthData holds the fields sent by the Telegram Login Widget.
type TelegramAuthData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// Verify checks the Telegram Login Widget signature and recency (<24 h).
func (d *TelegramAuthData) Verify(botToken string) error {
	if time.Now().Unix()-d.AuthDate > 86400 {
		return errors.New("auth_date expired")
	}

	// Build data-check string: sorted key=value pairs (excluding hash) joined by \n
	fields := map[string]string{
		"id":        fmt.Sprintf("%d", d.ID),
		"auth_date": fmt.Sprintf("%d", d.AuthDate),
	}
	if d.FirstName != "" {
		fields["first_name"] = d.FirstName
	}
	if d.LastName != "" {
		fields["last_name"] = d.LastName
	}
	if d.Username != "" {
		fields["username"] = d.Username
	}
	if d.PhotoURL != "" {
		fields["photo_url"] = d.PhotoURL
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+fields[k])
	}
	dataCheckString := strings.Join(parts, "\n")

	// secret = SHA256(bot_token)
	h := sha256.New()
	h.Write([]byte(botToken))
	secret := h.Sum(nil)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(dataCheckString))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(d.Hash)) {
		return errors.New("invalid telegram auth hash")
	}
	return nil
}
