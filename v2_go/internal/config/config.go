package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// AppConfig holds the entire application configuration
type AppConfig struct {
	Strategy StrategyConfig `yaml:"strategy"`
	Scanning ScanningConfig `yaml:"scanning"`
	Analysis AnalysisConfig `yaml:"analysis"`
	System   SystemConfig   `yaml:"system"`

	// Secrets (Loaded from .env, not yaml)
	Binance struct {
		APIKey     string
		APISecret  string
		UseTestnet bool
	}
	Supabase struct {
		URL string
		Key string
	}
	Grok struct {
		APIKey string
	}
}

// StrategyConfig holds specific strategy settings
type StrategyConfig struct {
	Martingale MartingaleConfig `yaml:"martingale"`
	Risk       RiskConfig       `yaml:"risk_management"`
	Safety     SafetyConfig     `yaml:"safety"`
	Trailing   TrailingConfig   `yaml:"trailing_stop"`
	PartialTP  PartialTPConfig  `yaml:"partial_tp"`
}

type MartingaleConfig struct {
	Enabled        bool      `yaml:"enabled"`
	Leverage       int       `yaml:"leverage"` // NEW
	Steps          []int     `yaml:"steps"`
	StepDistances  []float64 `yaml:"step_distances"`
	StepWaitTimes  []int     `yaml:"step_wait_times"`
	MaxPositions   int       `yaml:"max_positions"`
	MinPumpPercent float64   `yaml:"min_pump_percent"`
	MinRSIEntry    float64   `yaml:"min_rsi_entry"`
	TakeProfit     float64   `yaml:"take_profit_percent"`
	HalfClose      float64   `yaml:"half_close_percent"`
}

type RiskConfig struct {
	EmergencyStopUSD     float64 `yaml:"emergency_stop_usd"`
	EmergencyStopPercent float64 `yaml:"emergency_stop_percent"`
	MaxDailyLoss         float64 `yaml:"max_daily_loss_percent"`
	DailyLossEnabled     bool    `yaml:"daily_loss_limit_enabled"`
	DynamicBlacklist     struct {
		Enabled  bool `yaml:"enabled"`
		StopLoss int  `yaml:"stop_losses_trigger"`
		Window   int  `yaml:"window_hours"`
		Duration int  `yaml:"duration_hours"`
	} `yaml:"dynamic_blacklist"`
}

type SafetyConfig struct {
	RSICircuitBreaker float64 `yaml:"rsi_circuit_breaker_limit"`
	MaxVolMultiplier  float64 `yaml:"max_volatility_multiplier"`
}

type TrailingConfig struct {
	Enabled             bool    `yaml:"enabled"`
	Activation          float64 `yaml:"activation_percent"`
	Callback            float64 `yaml:"callback_percent"`
	BreakevenEnabled    bool    `yaml:"breakeven_enabled"`
	BreakevenActivation float64 `yaml:"breakeven_activation"`
}

type PartialTPConfig struct {
	Enabled    bool    `yaml:"enabled"`
	Percent    float64 `yaml:"percent"`
	Activation float64 `yaml:"activation_percent"`
}

type ScanningConfig struct {
	Interval      int      `yaml:"interval_seconds"`
	TopPairs      int      `yaml:"top_pairs_count"`
	QuoteAsset    string   `yaml:"quote_asset"`
	UseVolatility bool     `yaml:"use_volatility_ranking"` // Implicit check if min_volatility > 0? No, let's look for field
	MinVolatility float64  `yaml:"min_volatility_percent"`
	MinVolume     float64  `yaml:"min_24h_volume_usdt"`
	Blacklist     []string `yaml:"blacklist"`
}

type AnalysisConfig struct {
	Timeframes map[string]string `yaml:"timeframes"`
	Indicators struct {
		EMA struct{ Fast, Slow int } `yaml:"ema"`
		RSI struct {
			Period               int
			Oversold, Overbought float64
		} `yaml:"rsi"`
		MACD struct{ Fast, Slow, Signal int } `yaml:"macd"`
		ATR  struct{ Period int }             `yaml:"atr"`
	} `yaml:"indicators"`
	AI struct {
		Grok struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"grok"`
		Vision struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"vision"`
	} `yaml:"ai"`
}

type SystemConfig struct {
	LogLevel           string `yaml:"log_level"`
	DisplayRefreshRate int    `yaml:"display_refresh_rate"` // Added missing field definition while here
	ShowIndicators     bool   `yaml:"show_indicators"`      // Added missing field definition while here
	UseTestnet         bool   `yaml:"use_testnet"`          // NEW
}

// LoadConfig reads .env and config.yaml and returns the AppConfig
func LoadConfig() (*AppConfig, error) {
	cfg := &AppConfig{}

	// 1. Manual Env Loading (Robust)
	// Try multiple paths
	paths := []string{".env", "../.env", "../../.env", "C:\\Users\\hebeg\\Documents\\GitHub\\Legendary_Scalper\\.env"}
	envLoaded := false

	for _, path := range paths {
		if file, err := os.Open(path); err == nil {
			fmt.Printf("📂 Loading .env from: %s\n", path)
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				// Skip comments and empty lines
				if len(line) == 0 || strings.HasPrefix(strings.TrimSpace(line), "#") {
					continue
				}
				// Parse KEY=VALUE
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					// Remove quotes if present
					val = strings.Trim(val, `"'`)
					os.Setenv(key, val)
				}
			}
			file.Close()
			envLoaded = true
			break // Stop after first successful load
		}
	}

	if !envLoaded {
		fmt.Println("⚠️  No .env file found in any search path.")
	}

	cfg.Binance.APIKey = os.Getenv("BINANCE_API_KEY")
	cfg.Binance.APISecret = os.Getenv("BINANCE_API_SECRET")
	cfg.Binance.UseTestnet = os.Getenv("USE_TESTNET") == "True"

	cfg.Supabase.URL = os.Getenv("SUPABASE_URL")
	cfg.Supabase.Key = os.Getenv("SUPABASE_KEY")
	cfg.Grok.APIKey = os.Getenv("GROK_API_KEY")

	// 2. Load Strategy from config.yaml
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		// Try parent directory if running from cmd subfolder
		yamlFile, err = os.ReadFile("../../config.yaml")
		if err != nil {
			return nil, fmt.Errorf("failed to read config.yaml: %v", err)
		}
	}

	err = yaml.Unmarshal(yamlFile, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config.yaml: %v", err)
	}

	// 3. Apply Testnet Overrides
	// If config.yaml says use_testnet: true, OR env var is "True", enable it.
	if cfg.System.UseTestnet {
		cfg.Binance.UseTestnet = true
	} else if os.Getenv("USE_TESTNET") == "True" {
		cfg.Binance.UseTestnet = true
	}

	return cfg, nil
}
