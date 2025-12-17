package main

import (
	"fmt"
	"os"
)

func main() {
	path := "C:\\Users\\hebeg\\Documents\\GitHub\\Legendary_Scalper\\.env"
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("❌ Failed to read: %v\n", err)
		return
	}

	fmt.Printf("📂 File Size: %d bytes\n", len(data))

	// Print first 20 bytes in hex to check for BOM
	fmt.Printf("🔍 First 20 bytes: %x\n", data[:min(20, len(data))])

	// Print as quoted string to show special chars
	content := string(data)
	if len(content) > 100 {
		content = content[:100] + "..."
	}
	fmt.Printf("📝 Content Preview: %q\n", content)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
