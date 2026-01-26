package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type CloudflaredConfig struct {
	Tunnel  string        `yaml:"tunnel"`
	Ingress []IngressRule `yaml:"ingress"`
}

type IngressRule struct {
	Hostname string `yaml:"hostname,omitempty"`
	Service  string `yaml:"service"`
}

func Parse(data string) (*CloudflaredConfig, error) {
	var cfg CloudflaredConfig
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse cloudflared config: %w", err)
	}
	return &cfg, nil
}

func (c *CloudflaredConfig) Hostnames() []string {
	hostnames := make([]string, 0, len(c.Ingress))
	for _, rule := range c.Ingress {
		if rule.Hostname != "" {
			hostnames = append(hostnames, rule.Hostname)
		}
	}
	return hostnames
}

func (c *CloudflaredConfig) TunnelTarget() string {
	return fmt.Sprintf("%s.cfargotunnel.com", c.Tunnel)
}
