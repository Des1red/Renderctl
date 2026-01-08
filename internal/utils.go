package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tvctrl/logger"
)

func (c Config) ControlURL() string {
	path := c.TPath
	if path == "" {
		path = "/"
	}

	if len(path) > 0 && path[0] != '/' {
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
