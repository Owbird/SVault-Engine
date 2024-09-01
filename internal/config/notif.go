package config

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
