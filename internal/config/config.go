package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is application configuration (non-secret knobs + pointers to env for secrets).
type Config struct {
	HTTP struct {
		Addr string `yaml:"addr"`
	} `yaml:"http"`
	Auth struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"auth"`
	Catalog struct {
		Backend string `yaml:"backend"` // memory | supabase
		Memory  struct {
			ProductsPath string `yaml:"products_path"`
		} `yaml:"memory"`
		Supabase struct {
			DSNEnv string `yaml:"dsn_env"` // env var name holding postgres URL
		} `yaml:"supabase"`
	} `yaml:"catalog"`
	Promo struct {
		DataDir string `yaml:"data_dir"` // directory containing couponbase1-3.gz
	} `yaml:"promo"`
}

// Load reads YAML from path and applies environment overrides.
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	applyEnv(&c)
	if err := validate(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func applyEnv(c *Config) {
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		c.HTTP.Addr = v
	}
	if v := os.Getenv("API_KEY"); v != "" {
		c.Auth.APIKey = v
	}
	if v := os.Getenv("DATABASE_URL"); v != "" && c.Catalog.Supabase.DSNEnv == "" {
		c.Catalog.Supabase.DSNEnv = "DATABASE_URL"
	}
	if v := os.Getenv("COUPON_DATA_DIR"); v != "" {
		c.Promo.DataDir = v
	}
}

func validate(c *Config) error {
	if c.HTTP.Addr == "" {
		c.HTTP.Addr = ":8080"
	}
	if c.Auth.APIKey == "" {
		c.Auth.APIKey = "apitest"
	}
	c.Catalog.Backend = strings.ToLower(strings.TrimSpace(c.Catalog.Backend))
	if c.Catalog.Backend == "" {
		c.Catalog.Backend = "memory"
	}
	switch c.Catalog.Backend {
	case "memory", "supabase":
	default:
		return fmt.Errorf("catalog.backend must be memory or supabase, got %q", c.Catalog.Backend)
	}
	if c.Catalog.Backend == "memory" && c.Catalog.Memory.ProductsPath == "" {
		c.Catalog.Memory.ProductsPath = "data/products.json"
	}
	if c.Catalog.Backend == "supabase" && c.Catalog.Supabase.DSNEnv == "" {
		c.Catalog.Supabase.DSNEnv = "DATABASE_URL"
	}
	if c.Promo.DataDir == "" {
		c.Promo.DataDir = "data"
	}
	return nil
}

// DSN returns the Postgres connection string for supabase backend.
func (c *Config) DSN() (string, error) {
	name := c.Catalog.Supabase.DSNEnv
	if name == "" {
		return "", fmt.Errorf("supabase dsn env not configured")
	}
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("environment %q is not set", name)
	}
	return v, nil
}
