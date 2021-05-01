package config

import (
	"errors"
	"net"
)

type Config struct {
	Name       string   `json:"name"`
	Port       int      `json:"port"`
	TargetPort int      `json:"target_port"`
	IPList     []string `json:"ip_list"`
}

func (cfg *Config) Valid() error {
	if !validPort(cfg.Port) {
		return errors.New("error Port is not valid")
	}
	if !validPort(cfg.TargetPort) {
		return errors.New("error target_port is not valid")
	}
	for _, ip := range cfg.IPList {
		if !validIP(ip) {
			return errors.New("error ip is not valid: " + ip)
		}
	}
	return nil
}

func (cfg *Config) Same(targetCfg *Config) bool {
	if cfg.Name != targetCfg.Name {
		return false
	}
	if cfg.Port != targetCfg.Port {
		return false
	}
	if cfg.TargetPort != targetCfg.TargetPort {
		return false
	}
	return sameStringArray(cfg.IPList, targetCfg.IPList)
}

func sameStringArray(arr1, arr2 []string) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for _, ele1 := range arr1 {
		find := false
		for _, ele2 := range arr2 {
			if ele1 == ele2 {
				find = true
			}
		}
		if !find {
			return false
		}
	}
	return true
}

func validPort(p int) bool {
	if p <= 0 {
		return false
	}
	if p > 65535 {
		return false
	}
	return true
}

func validIP(ip string) bool {
	if net.ParseIP(ip) == nil {
		return false
	}
	return true
}
