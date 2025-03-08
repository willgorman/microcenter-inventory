package inventory

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"github.com/willgorman/microcenter-inventory/config"
)

// TestCheckProductInventory tests the product inventory checker
func TestCheckProductInventory(t *testing.T) {
	// Skip this test if the environment variable is not set
	if os.Getenv("RUN_SELENIUM_TESTS") != "1" {
		t.Skip("Skipping Selenium test; set RUN_SELENIUM_TESTS=1 to run")
	}

	// You'll need ChromeDriver installed for this test
	chromeDriverPath := os.Getenv("CHROME_DRIVER_PATH")
	if chromeDriverPath == "" {
		chromeDriverPath = "/usr/local/bin/chromedriver" // default path
	}

	// Create a test config
	cfg := &config.Config{
		StoreID:       141, // Example store ID
		CheckInterval: 5 * time.Minute,
		SeleniumOpts: config.SeleniumOptions{
			ChromeDriverPath: chromeDriverPath,
			Headless:         true,
			Timeout:          30 * time.Second,
		},
	}

	// Setup Selenium manually for testing
	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(cfg.SeleniumOpts.ChromeDriverPath, 4444, opts...)
	if err != nil {
		t.Fatalf("Failed to create Chrome driver service: %v", err)
	}
	defer service.Stop()

	// Configure Chrome
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	chromeCaps := chrome.Capabilities{
		Args: []string{"--headless"},
	}

	caps.AddChrome(chromeCaps)

	// Create WebDriver
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 4444))
	if err != nil {
		t.Fatalf("Failed to create WebDriver: %v", err)
	}
	defer wd.Quit()

	// Set timeout
	if err := wd.SetImplicitWaitTimeout(cfg.SeleniumOpts.Timeout); err != nil {
		t.Fatalf("Failed to set timeout: %v", err)
	}

	// Create a checker with our test components
	checker := &Checker{
		config:    cfg,
		service:   service,
		webDriver: wd,
	}

	// Test cases
	testCases := []struct {
		name    string
		product config.Product
	}{
		{
			name: "Raspberry Pi 4",
			product: config.Product{
				Name: "Raspberry Pi 4",
				URL:  "https://www.microcenter.com/product/621439/raspberry-pi-4-model-b-4gb-ddr4",
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := checker.CheckProductInventory(fmt.Sprintf("%d", cfg.StoreID), tc.product)
			
			if result.Error != nil {
				t.Errorf("Error checking inventory: %v", result.Error)
			}
			
			t.Logf("Product: %s", result.ProductName)
			t.Logf("Store ID: %s", result.StoreID)
			t.Logf("Raw Text: %s", result.RawText)
			t.Logf("Count: %d", result.Count)
			
			// Here we're just logging the results, but you could add assertions
			// once you know what to expect from the website
		})
	}
}

// TestParseInventoryCount tests the inventory count parsing logic
func TestParseInventoryCount(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
	}{
		{"5 in stock at Tustin Store", 5},
		{"In Stock at Tustin Store: 3", 3},
		{"Out of Stock at Tustin Store", 0},
		{"Limited Stock at Tustin Store", 0}, // You might want to handle this differently
		{"10+ in stock", 10},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			count := parseInventoryCount(tc.input)
			if count != tc.expected {
				t.Errorf("Expected %d, got %d for input: %s", tc.expected, count, tc.input)
			}
		})
	}
}

// This is a simple standalone test runner that can be used during development
func Example_manualTest() {
	// This function won't run as a test but shows how to use the checker manually
	// for debugging purposes.
	
	chromeDriverPath := "/usr/local/bin/chromedriver" // Update with your path
	
	// Create a test config
	cfg := &config.Config{
		StoreID:       141, // Example store ID
		CheckInterval: 5 * time.Minute,
		SeleniumOpts: config.SeleniumOptions{
			ChromeDriverPath: chromeDriverPath,
			Headless:         true,
			Timeout:          30 * time.Second,
		},
	}

	// Setup Selenium manually
	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(cfg.SeleniumOpts.ChromeDriverPath, 4444, opts...)
	if err != nil {
		log.Fatalf("Failed to create Chrome driver service: %v", err)
		return
	}
	defer service.Stop()

	// Configure Chrome
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	chromeCaps := chrome.Capabilities{
		Args: []string{"--headless"},
	}

	caps.AddChrome(chromeCaps)

	// Create WebDriver
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 4444))
	if err != nil {
		log.Fatalf("Failed to create WebDriver: %v", err)
		return
	}
	defer wd.Quit()

	// Set timeout
	if err := wd.SetImplicitWaitTimeout(cfg.SeleniumOpts.Timeout); err != nil {
		log.Fatalf("Failed to set timeout: %v", err)
		return
	}

	// Create a checker with our test components
	checker := &Checker{
		config:    cfg,
		service:   service,
		webDriver: wd,
	}

	// Define a test product
	product := config.Product{
		Name: "Raspberry Pi 4",
		URL:  "https://www.microcenter.com/product/621439/raspberry-pi-4-model-b-4gb-ddr4",
	}

	// Check inventory
	result := checker.CheckProductInventory(fmt.Sprintf("%d", cfg.StoreID), product)
	
	// Display results
	if result.Error != nil {
		log.Printf("Error checking inventory: %v", result.Error)
	} else {
		log.Printf("Product: %s", result.ProductName)
		log.Printf("Store ID: %s", result.StoreID)
		log.Printf("Raw Text: %s", result.RawText)
		log.Printf("Count: %d", result.Count)
	}

	// Output: 
	// (No output as this is just an example function)
}
