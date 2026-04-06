package config

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
	"gopkg.in/yaml.v3"
)

type ConfigManager struct {
	mu       sync.RWMutex
	config   *domain.Config
	filepath string
	lastMod  time.Time
	OnReload func(added, removed []string)
}

var _ ports.ConfigProvider = (*ConfigManager)(nil)
var _ ports.ConfigWatcher = (*ConfigManager)(nil)
var _ ports.ConfigManager = (*ConfigManager)(nil)

func NewConfigManager(path string) (*ConfigManager, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}

	stat, _ := os.Stat(path)

	return &ConfigManager{
		config:   cfg,
		filepath: path,
		lastMod:  stat.ModTime(),
	}, nil
}

func (c *ConfigManager) Get() *domain.Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

func (c *ConfigManager) Watch() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stat, err := os.Stat(c.filepath)
		if err != nil {
			continue
		}

		if stat.ModTime().After(c.lastMod) {
			c.loadAndCompare(stat.ModTime())
		}
	}
}

func (c *ConfigManager) loadAndCompare(newModTime time.Time) {
	newConfig, err := Load(c.filepath)
	if err != nil {
		slog.Error("Error loading config via Hot Reload", "error", err)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	oldMap := make(map[string]bool)
	for _, s := range c.config.Searches {
		oldMap[s.Term] = true
	}

	newMap := make(map[string]bool)
	var added []string
	for _, s := range newConfig.Searches {
		newMap[s.Term] = true
		if !oldMap[s.Term] {
			added = append(added, s.Term)
		}
	}

	var removed []string
	for _, s := range c.config.Searches {
		if !newMap[s.Term] {
			removed = append(removed, s.Term)
		}
	}

	if len(added) > 0 || len(removed) > 0 {
		slog.Info("🔥 Hot Reload: Search configuration changed!", "added", added, "removed", removed)

		if c.OnReload != nil {
			c.OnReload(added, removed)
		}
	} else {
		slog.Info("🔥 Hot Reload: File updated (internal edit).")
	}

	c.config = newConfig
	c.lastMod = newModTime
}

func Load(filepath string) (*domain.Config, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var cfg domain.Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing yaml: %v", err)
	}

	return &cfg, nil
}
