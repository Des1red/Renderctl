package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tvctrl/logger"
)

func (c Config) ControlURL() string {
	if c._CachedControlURL != "" {
		return c._CachedControlURL
	}

	if c.TIP == "" || c.TPort == "" {
		return ""
	}

	path := c.TPath
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}

	return "http://" + c.TIP + ":" + c.TPort + path
}

func (c Config) BaseUrl() string {
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

func (cfg Config) MediaURL() string {
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

func confirm(msg string) bool {
	var ans string
	logger.Info("%s (y/n): ", msg)
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	return ans == "y" || ans == "yes"
}
