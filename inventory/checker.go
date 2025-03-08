package inventory

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"github.com/willgorman/microcenter-inventory/config"
	"github.com/willgorman/microcenter-inventory/prometheus"
)

// Checker handles checking product inventory
type Checker struct {
	config      *config.Config
	service     *selenium.Service
	webDriver   selenium.WebDriver
}

// NewChecker creates and initializes a new inventory checker
func NewChecker(cfg *config.Config) (*Checker, error) {
	// Setup Selenium
	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(cfg.SeleniumOpts.ChromeDriverPath, 4444, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Chrome driver service: %w", err)
	}

	// Configure Chrome
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	chromeCaps := chrome.Capabilities{
		Args: []string{},
	}

	if cfg.SeleniumOpts.Headless {
		chromeCaps.Args = append(chromeCaps.Args, "--headless")
	}

	caps.AddChrome(chromeCaps)

	// Create WebDriver
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 4444))
	if err != nil {
		service.Stop()
		return nil, fmt.Errorf("failed to create WebDriver: %w", err)
	}

	// Set timeout
	if err := wd.SetImplicitWaitTimeout(cfg.SeleniumOpts.Timeout); err != nil {
		wd.Quit()
		service.Stop()
		return nil, fmt.Errorf("failed to set timeout: %w", err)
	}

	return &Checker{
		config:    cfg,
		service:   service,
		webDriver: wd,
	}, nil
}

// Close cleans up resources
func (c *Checker) Close() {
	if c.webDriver != nil {
		c.webDriver.Quit()
	}
	if c.service != nil {
		c.service.Stop()
	}
}

// Start begins the periodic inventory checking
func (c *Checker) Start(ctx context.Context) {
	ticker := time.NewTicker(c.config.CheckInterval)
	defer ticker.Stop()

	// Do an initial check immediately
	c.checkInventory()

	for {
		select {
		case <-ticker.C:
			c.checkInventory()
		case <-ctx.Done():
			log.Println("Inventory checker stopping due to context cancellation")
			return
		}
	}
}

// checkInventory checks inventory for all configured products
func (c *Checker) checkInventory() {
	storeID := strconv.Itoa(c.config.StoreID)
	
	for _, product := range c.config.Products {
		timer := prometheus.ScrapeDurationSeconds.WithLabelValues(product.URL)
		start := time.Now()
		
		err := c.checkProductInventory(storeID, product)
		
		duration := time.Since(start).Seconds()
		timer.Observe(duration)
		
		if err != nil {
			log.Printf("Error checking inventory for %s: %v", product.Name, err)
			prometheus.ScrapeFailureCounter.WithLabelValues(product.URL, err.Error()).Inc()
		} else {
			prometheus.ScrapeSuccessCounter.WithLabelValues(product.URL).Inc()
		}
	}
}

// checkProductInventory checks inventory for a single product
func (c *Checker) checkProductInventory(storeID string, product config.Product) error {
	// Navigate to the product page
	if err := c.webDriver.Get(product.URL); err != nil {
		return fmt.Errorf("failed to load page: %w", err)
	}

	// Find the store selection dropdown (if needed)
	// In some cases, we might need to select the store first
	
	// Look for inventory information
	// This selector might need adjustment based on MicroCenter's actual HTML structure
	inventoryElement, err := c.webDriver.FindElement(selenium.ByCSSSelector, ".inventory-msg")
	if err != nil {
		return fmt.Errorf("failed to find inventory element: %w", err)
	}

	inventoryText, err := inventoryElement.Text()
	if err != nil {
		return fmt.Errorf("failed to get inventory text: %w", err)
	}

	// Parse the inventory count from text like "5 in stock at [Store]"
	count := parseInventoryCount(inventoryText)
	
	// Update the Prometheus gauge
	prometheus.ProductInventoryGauge.WithLabelValues(
		storeID,
		product.Name,
		product.URL,
	).Set(float64(count))

	log.Printf("Product: %s, Store: %s, Inventory: %d", product.Name, storeID, count)
	return nil
}

// parseInventoryCount extracts the inventory count from the text
func parseInventoryCount(text string) int {
	// This is a simplified example. Actual parsing logic will depend on MicroCenter's format.
	text = strings.TrimSpace(text)
	
	// Handle "Out of stock" case
	if strings.Contains(strings.ToLower(text), "out of stock") {
		return 0
	}
	
	// Try to extract a number
	for _, word := range strings.Fields(text) {
		if num, err := strconv.Atoi(word); err == nil {
			return num
		}
	}
	
	return 0
}
