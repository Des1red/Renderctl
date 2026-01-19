package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"renderctl/internal/models"
	"renderctl/logger"
	"strings"
)

func NormalizeMode(mode string) string {
	return strings.ToLower(strings.TrimSpace(mode))
}

func ControlURL(cfg *models.Config) string {
	if cfg.TIP == "" || cfg.TPort == "" {
		return ""
	}

	path := cfg.TPath
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}

	if strings.HasPrefix(cfg.TPath, "http://") || strings.HasPrefix(cfg.TPath, "https://") {
		return cfg.TPath
	}
	return fmt.Sprintf("http://%s:%s%s", cfg.TIP, cfg.TPort, path)

}

func BaseUrl(c *models.Config) string {
	return "http://" + c.TIP + ":" + c.TPort
}

func ValidateFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}
	return nil
}

func MediaURL(cfg *models.Config) string {
	file := filepath.Base(cfg.LFile)
	return "http://" + cfg.LIP + ":" + cfg.ServePort + "/" + file
}

func LocalIP(ip string) string {
	if ip == "" {
		var newip string
		fmt.Print("Enter local IP: ")
		fmt.Scan(&newip)
		if strings.TrimSpace(newip) != "" {
			ip = newip
		} else {
			logger.Fatal("Missing -Lip (local IP for media serving)")
			os.Exit(1)
		}
	}
	return ip
}

func Confirm(msg string) bool {
	var ans string
	logger.Info("%s (y/n): ", msg)
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	return ans == "y" || ans == "yes"
}
