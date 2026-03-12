package config_test

import (
	"testing"
	"time"

	"github.com/yourname/tunnel/internal/config"
)

func TestServerDefaults(t *testing.T) {
	c := config.ServerDefaults()
	if c.HTTPPort != config.DefaultHTTPPort {
		t.Errorf("HTTPPort = %d, want %d", c.HTTPPort, config.DefaultHTTPPort)
	}
	if c.ControlPort != config.DefaultControlPort {
		t.Errorf("ControlPort = %d, want %d", c.ControlPort, config.DefaultControlPort)
	}
	if c.AdminPort != config.DefaultAdminPort {
		t.Errorf("AdminPort = %d, want %d", c.AdminPort, config.DefaultAdminPort)
	}
	if c.Workers != 0 {
		t.Errorf("Workers = %d, want 0", c.Workers)
	}
	if c.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", c.LogLevel)
	}
	if c.PProfEnabled {
		t.Error("PProfEnabled = true, want false")
	}
}

func TestClientDefaults(t *testing.T) {
	c := config.ClientDefaults()
	if c.LocalURL != "http://127.0.0.1:5173" {
		t.Errorf("LocalURL = %q, want http://127.0.0.1:5173", c.LocalURL)
	}
	if c.ReconnectMax != 5*time.Minute {
		t.Errorf("ReconnectMax = %v, want 5m", c.ReconnectMax)
	}
	if c.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", c.LogLevel)
	}
}

func TestValidate_Server_Valid(t *testing.T) {
	c := config.ServerDefaults()
	if err := c.Validate(true); err != nil {
		t.Errorf("Validate(server): %v", err)
	}
}

func TestValidate_Server_InvalidPorts(t *testing.T) {
	c := config.ServerDefaults()
	c.HTTPPort = 0
	if err := c.Validate(true); err == nil {
		t.Error("Validate(server) with HTTPPort 0: want error")
	}
	c.HTTPPort = config.DefaultHTTPPort
	c.ControlPort = 70000
	if err := c.Validate(true); err == nil {
		t.Error("Validate(server) with ControlPort 70000: want error")
	}
	c.ControlPort = config.DefaultControlPort
	c.AdminPort = -1
	if err := c.Validate(true); err == nil {
		t.Error("Validate(server) with AdminPort -1: want error")
	}
}

func TestValidate_Server_InvalidWorkers(t *testing.T) {
	c := config.ServerDefaults()
	c.Workers = -1
	if err := c.Validate(true); err == nil {
		t.Error("Validate(server) with Workers -1: want error")
	}
}

func TestValidate_Server_InvalidLogLevel(t *testing.T) {
	c := config.ServerDefaults()
	c.LogLevel = "invalid"
	if err := c.Validate(true); err == nil {
		t.Error("Validate(server) with LogLevel invalid: want error")
	}
}

func TestValidate_Client_Valid(t *testing.T) {
	c := config.ClientDefaults()
	c.ServerAddr = "163.245.222.142:3000"
	c.LocalURL = "http://127.0.0.1:5173"
	if err := c.Validate(false); err != nil {
		t.Errorf("Validate(client): %v", err)
	}
}

func TestValidate_Client_MissingServerAddr(t *testing.T) {
	c := config.ClientDefaults()
	c.LocalURL = "http://127.0.0.1:5173"
	if err := c.Validate(false); err == nil {
		t.Error("Validate(client) without ServerAddr: want error")
	}
}

func TestValidate_Client_MissingLocalURL(t *testing.T) {
	c := config.ClientDefaults()
	c.ServerAddr = "163.245.222.142:3000"
	c.LocalURL = ""
	if err := c.Validate(false); err == nil {
		t.Error("Validate(client) without LocalURL: want error")
	}
}

func TestValidate_Client_InvalidLocalURL(t *testing.T) {
	c := config.ClientDefaults()
	c.ServerAddr = "163.245.222.142:3000"
	c.LocalURL = "://bad"
	if err := c.Validate(false); err == nil {
		t.Error("Validate(client) with invalid LocalURL: want error")
	}
}

func TestValidate_Client_SubdomainTooLong(t *testing.T) {
	c := config.ClientDefaults()
	c.ServerAddr = "163.245.222.142:3000"
	c.LocalURL = "http://127.0.0.1:5173"
	c.Subdomain = string(make([]byte, config.MaxDomainLen+1))
	if err := c.Validate(false); err == nil {
		t.Error("Validate(client) with Subdomain > MaxDomainLen: want error")
	}
}

func TestValidate_Client_NegativeReconnectMax(t *testing.T) {
	c := config.ClientDefaults()
	c.ServerAddr = "163.245.222.142:3000"
	c.LocalURL = "http://127.0.0.1:5173"
	c.ReconnectMax = -1
	if err := c.Validate(false); err == nil {
		t.Error("Validate(client) with ReconnectMax -1: want error")
	}
}

func TestValidate_Client_InvalidLogLevel(t *testing.T) {
	c := config.ClientDefaults()
	c.ServerAddr = "163.245.222.142:3000"
	c.LocalURL = "http://127.0.0.1:5173"
	c.LogLevel = "trace"
	if err := c.Validate(false); err == nil {
		t.Error("Validate(client) with LogLevel trace: want error")
	}
}

func TestValidate_LogLevel_EmptyAllowed(t *testing.T) {
	c := config.ServerDefaults()
	c.LogLevel = ""
	if err := c.Validate(true); err != nil {
		t.Errorf("Validate(server) with empty LogLevel: %v", err)
	}
}
