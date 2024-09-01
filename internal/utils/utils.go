package utils

import (
	"os"
	"path/filepath"

	"github.com/Owbird/SVault-Engine/pkg/models"
	"github.com/atotto/clipboard"
	"github.com/martinlindhe/notify"
)

func GetSVaultDir() (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userDir, ".svault"), nil
}

func SendNotification(notification models.Notification) {
	notify.Notify("SVault", notification.Title, notification.Body, "")

	if notification.ClipboardText != "" {
		clipboard.WriteAll(notification.ClipboardText)
	}
}
