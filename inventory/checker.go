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
)

// InventoryResult represents the result of a product inventory check
type InventoryResult struct {
	StoreID     string
	ProductName string
	ProductURL  string
	Count       int
	RawText     string
	LastChecked time.Time
	Error       error
}

// Checker handles checking product inventory
type Checker struct {
	config    *config.Config
	service   *selenium.Service
	webDriver selenium.WebDriver
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

	err = wd.AddCookie(&selenium.Cookie{
		Name:   "storeSelected",
		Value:  strconv.Itoa(cfg.StoreID),
		Path:   "/",
		Domain: ".microcenter.com",
		Secure: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set store: %w", err)
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
func (c *Checker) Start(ctx context.Context, resultChan chan<- InventoryResult) {
	ticker := time.NewTicker(c.config.CheckInterval)
	defer ticker.Stop()

	// Do an initial check immediately
	c.checkInventory(resultChan)

	for {
		select {
		case <-ticker.C:
			c.checkInventory(resultChan)
		case <-ctx.Done():
			log.Println("Inventory checker stopping due to context cancellation")
			return
		}
	}
}

// checkInventory checks inventory for all configured products
func (c *Checker) checkInventory(resultChan chan<- InventoryResult) {
	storeID := strconv.Itoa(c.config.StoreID)

	for _, product := range c.config.Products {
		result := c.CheckProductInventory(storeID, product)
		resultChan <- result
	}
}

// CheckProductInventory checks inventory for a single product
func (c *Checker) CheckProductInventory(storeID string, product config.Product) InventoryResult {
	result := InventoryResult{
		StoreID:     storeID,
		ProductName: product.Name,
		ProductURL:  product.URL,
		LastChecked: time.Now(),
	}

	// Navigate to the product page
	if err := c.webDriver.Get(product.URL); err != nil {
		result.Error = fmt.Errorf("failed to load page: %w", err)
		return result
	}

	// FIXME: (willgorman) takes some time before the count loads, how can I check that instead of just waiting?
	time.Sleep(15 * time.Second)

	// Look for inventory information
	// This selector might need adjustment based on MicroCenter's actual HTML structure
	inventoryElement, err := c.webDriver.FindElement(selenium.ByID, "pnlInventory")
	if err != nil {
		result.Error = fmt.Errorf("failed to find inventory element: %w", err)
		return result
	}

	inventoryText, err := inventoryElement.Text()
	if err != nil {
		result.Error = fmt.Errorf("failed to get inventory text: %w", err)
		return result
	}

	result.RawText = inventoryText
	result.Count = parseInventoryCount(inventoryText)

	return result
}

// parseInventoryCount extracts the inventory count from the text
func parseInventoryCount(text string) int {
	// This is a simplified example. Actual parsing logic will depend on MicroCenter's format.
	text = strings.TrimSpace(text)

	// Handle "Out of stock" case
	if strings.Contains(strings.ToLower(text), "sold out") {
		return 0
	}

	// Try to extract a number
	for _, word := range strings.Fields(text) {
		word = strings.Replace(word, "+", "", -1)
		if num, err := strconv.Atoi(word); err == nil {
			return num
		}
	}

	return 0
}
