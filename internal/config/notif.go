package config

import (
	"github.com/Owbird/SVault-Engine/pkg/models"
	"github.com/atotto/clipboard"
	"github.com/martinlindhe/notify"
)

type NotifConfig struct {
	// Should uploads be allowed
	allowNotifs bool
}

func NewNotifConfig() *NotifConfig {
	return &NotifConfig{}
}

// SetAllowNotif sets if notifications are allowed
// Defaults to true
func (nc *NotifConfig) SetAllowNotif(allowNotifs bool) *NotifConfig {
	nc.allowNotifs = allowNotifs
	return nc
}

// GetName returns whether notifications are allowed
func (nc *NotifConfig) GetAllowNotif() bool {
	return nc.allowNotifs
}

func (nc *NotifConfig) SendNotification(notification models.Notification) {
	if nc.GetAllowNotif() {
		notify.Notify("SVault", notification.Title, notification.Body, "")
	}

	if notification.ClipboardText != "" {
		clipboard.WriteAll(notification.ClipboardText)
	}
}
