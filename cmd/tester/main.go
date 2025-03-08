package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/willgorman/microcenter-inventory/config"
	"github.com/willgorman/microcenter-inventory/inventory"
)

func main() {
	// Parse flags
	storeID := flag.Int("store", 191, "MicroCenter store ID")
	url := flag.String("url", "", "Product URL to check")
	name := flag.String("name", "Test Product", "Product name")
	chromeDriver := flag.String("chrome-driver", "/home/will/.flox/run/x86_64-linux.default.dev/bin/chromedriver", "Path to chromedriver")
	headless := flag.Bool("headless", true, "Run Chrome in headless mode")
	timeout := flag.Duration("timeout", 30*time.Second, "Browser timeout")
	
	flag.Parse()

	if *url == "" {
		log.Fatal("Product URL is required. Use --url flag.")
	}

	// Create config
	cfg := &config.Config{
		StoreID:       *storeID,
		CheckInterval: 5 * time.Minute,
		SeleniumOpts: config.SeleniumOptions{
			ChromeDriverPath: *chromeDriver,
			Headless:         *headless,
			Timeout:          *timeout,
		},
	}

	// Create checker
	checker, err := inventory.NewChecker(cfg)
	if err != nil {
		log.Fatalf("Failed to create inventory checker: %v", err)
	}
	defer checker.Close()

	// Define product
	product := config.Product{
		Name: *name,
		URL:  *url,
	}

	// Check inventory
	fmt.Printf("Checking inventory for %s at store %d...\n", product.Name, cfg.StoreID)
	result := checker.CheckProductInventory(fmt.Sprintf("%d", cfg.StoreID), product)
	
	// Display results
	fmt.Println("\nResults:")
	fmt.Println("-----------------------------------")
	
	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	} else {
		fmt.Printf("Product: %s\n", result.ProductName)
		fmt.Printf("Store ID: %s\n", result.StoreID)
		fmt.Printf("Raw Text: %s\n", result.RawText)
		fmt.Printf("Parsed Count: %d\n", result.Count)
		fmt.Printf("Checked At: %s\n", result.LastChecked.Format(time.RFC3339))
	}
}
