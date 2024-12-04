package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func GetSVaultDir() (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userDir, ".svault"), nil
}

func GetLocalIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "0.0.0.0", err
	}

	for _, addr := range addrs {
		ip, ok := addr.(*net.IPNet)
		if ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				return ip.IP.String(), nil
			}
		}
	}
	return "0.0.0.0", nil
}

func FmtBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
