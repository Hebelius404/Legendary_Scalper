package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"legendary-scalper/internal/bot"
	"legendary-scalper/internal/config"
)

func main() {
	fmt.Println("🤖 Legendary Scalper v2.0 (Go Edition) Starting...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	fmt.Println("✅ Configuration loaded successfully")
	fmt.Printf("   Strategy: Martingale (Enabled: %v)\n", cfg.Strategy.Martingale.Enabled)
	fmt.Printf("   Pairs to Scan: %d\n", cfg.Scanning.TopPairs)
	if cfg.Binance.UseTestnet {
		fmt.Println("   Mode: TESTNET 🛡️")
	} else {
		fmt.Println("   Mode: PRODUCTION 🚀")
	}

	// DEBUG: Check API Key (Masked)
	if len(cfg.Binance.APIKey) > 10 {
		fmt.Printf("   ApiKey: %s******\n", cfg.Binance.APIKey[:4])
	} else {
		fmt.Println("   ApiKey: [MISSING OR SHORT]")
	}

	// Initialize Bot Engine
	engine := bot.NewBotEngine(cfg)

	// Verify Connectivity (Optional, Engine does it internally too but good for sanity)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t, err := engine.Client.GetServerTime(ctx)
	if err != nil {
		log.Fatalf("❌ Connectivity verification failed: %v", err)
	}
	fmt.Printf("✅ Connected! Server Time: %d\n", t)

	// START THE ENGINE
	fmt.Println("🚦 Starting Main Loop...")
	engine.Run()
}
