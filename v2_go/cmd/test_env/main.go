package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	cwd, _ := os.Getwd()
	fmt.Printf("📂 CWD: %s\n", cwd)

	// Test Absolute Path
	absPath := "C:\\Users\\hebeg\\Documents\\GitHub\\Legendary_Scalper\\.env"
	err := godotenv.Load(absPath)
	if err == nil {
		fmt.Printf("✅ Loaded ABOLUTE %s\n", absPath)
	} else {
		fmt.Printf("❌ Failed ABSOLUTE %s: %v\n", absPath, err)
	}

	// List parent dir
	files, _ := os.ReadDir("..")
	fmt.Println("📂 Listing .. :")
	for _, f := range files {
		if f.Name() == ".env" {
			fmt.Println("   FOUND .env!")
		}
	}

	key := os.Getenv("BINANCE_API_KEY")
	if len(key) > 5 {
		fmt.Printf("🔑 BINANCE_API_KEY: %s******\n", key[:4])
	} else {
		fmt.Printf("🔑 BINANCE_API_KEY: [EMPTY/MISSING] (Len: %d)\n", len(key))
	}
}
