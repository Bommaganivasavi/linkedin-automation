package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/sirupsen/logrus"
	"github.com/your-github-username/linkedin-automation/internal/config"
	"github.com/your-github-username/linkedin-automation/internal/storage"
)

// LinkedInBot handles LinkedIn automation
type LinkedInBot struct {
	browser  *rod.Browser
	page     *rod.Page
	email    string
	password string
	logger   *logrus.Logger
	rand     *rand.Rand
	config   *config.Config
	storage  *storage.Storage
}

// NewLinkedInBot creates a new LinkedInBot instance with the provided credentials
func NewLinkedInBot(cfg *config.Config, store *storage.Storage, email, password string) (*LinkedInBot, error) {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Find Chrome browser path
	path, err := launcher.LookPath()
	if err != nil {
		return nil, fmt.Errorf("failed to find Chrome browser: %v", err)
	}

	// Configure launcher
	l := launcher.New()
	l.Headless(cfg.Browser.Headless)
	l.Set("disable-blink-features", "AutomationControlled")
	l.Delete("enable-automation")
	l.Set("disable-web-security")
	l.Set("disable-dev-shm-usage")
	l.Set("no-sandbox")
	l.Bin(path)

	// Launch browser
	url := l.MustLaunch()
	browser := rod.New().
		ControlURL(url).
		Trace(true).
		SlowMotion(2 * time.Second).
		MustConnect()

	// Create stealth page
	page := stealth.MustPage(browser)

	// Set viewport and user agent from config
	page.MustSetViewport(cfg.Browser.Viewport.Width, cfg.Browser.Viewport.Height, 1, false)
	err = proto.EmulationSetUserAgentOverride{
		UserAgent: cfg.Browser.UserAgent,
	}.Call(page)

	if err != nil {
		browser.MustClose()
		return nil, fmt.Errorf("failed to set user agent: %v", err)
	}

	// Initialize random number generator
	randSrc := rand.NewSource(time.Now().UnixNano())

	return &LinkedInBot{
		browser:  browser,
		page:     page,
		email:    email,
		password: password,
		logger:   logger,
		rand:     rand.New(randSrc),
		config:   cfg,
		storage:  store,
	}, nil
}

// Close cleans up the browser resources
func (b *LinkedInBot) Close() {
	b.logger.Info("Closing browser...")
	if b.browser != nil {
		_ = b.browser.Close()
	}
}

// randomDelay pauses execution for a random duration between min and max seconds
func (b *LinkedInBot) randomDelay() {
	min := b.config.Timing.MinDelaySeconds
	max := b.config.Timing.MaxDelaySeconds
	delay := time.Duration(b.rand.Intn(max-min+1)+min) * time.Second
	b.logger.Debugf("Delaying for %v...", delay)
	time.Sleep(delay)
}

// humanType simulates human-like typing with random delays and occasional typos
func (b *LinkedInBot) humanType(element *rod.Element, text string) error {
	if !b.config.Stealth.HumanTyping {
		element.MustInput(text)
		return nil
	}

	minSpeed := b.config.Timing.TypingSpeedMinMs
	maxSpeed := b.config.Timing.TypingSpeedMaxMs

	for _, char := range text {
		// 5% chance of making a typo
		if b.config.Stealth.HumanTyping && b.rand.Float64() < 0.05 {
			// Type a random wrong character (a-z)
			wrongChar := rune('a' + b.rand.Intn(26))
			element.MustInput(string(wrongChar))

			// Wait before backspace
			time.Sleep(time.Duration(100+b.rand.Intn(201)) * time.Millisecond)

			// Send backspace
			element.MustKeyActions().Press('\b').MustDo()
			time.Sleep(time.Duration(50+b.rand.Intn(101)) * time.Millisecond)
		}

		// Type the correct character
		element.MustInput(string(char))
		time.Sleep(time.Duration(minSpeed+b.rand.Intn(maxSpeed-minSpeed+1)) * time.Millisecond)
	}
	return nil
}

// humanScroll simulates human-like scrolling
func (b *LinkedInBot) humanScroll() {
	if !b.config.Stealth.RandomScrolling {
		return
	}

	totalDistance := 200 + b.rand.Intn(401) // 200-600 pixels
	steps := 10 + b.rand.Intn(11)           // 10-20 steps
	stepSize := totalDistance / steps

	// Scroll down
	for i := 0; i < steps; i++ {
		b.page.Mouse.MustScroll(0, stepSize)
		time.Sleep(time.Duration(20+b.rand.Intn(31)) * time.Millisecond)
	}

	// 30% chance to scroll back a bit
	if b.rand.Float64() < 0.3 {
		time.Sleep(time.Duration(200+b.rand.Intn(301)) * time.Millisecond)
		scrollBack := -50 - b.rand.Intn(101) // -50 to -150 pixels
		b.page.Mouse.MustScroll(0, scrollBack)
	}
}

// moveMouseToElement moves mouse to element with human-like movement
func (b *LinkedInBot) moveMouseToElement(element *rod.Element) error {
	if !b.config.Stealth.MouseMovement {
		element.MustScrollIntoView()
		element.MustWaitVisible()
		return nil
	}

	rect, err := element.Shape()
	if err != nil {
		return fmt.Errorf("failed to get element shape: %v", err)
	}
	element.MustScrollIntoView()
	targetX := int(rect.X + rect.Width/2)
	targetY := int(rect.Y + rect.Height/2)

	// Start from random position near current position
	startX := targetX - 100 + b.rand.Intn(200)
	startY := targetY - 100 + b.rand.Intn(200)

	// Number of steps for smooth movement
	steps := 15 + b.rand.Intn(11) // 15-25 steps

	// Move mouse along Bezier curve
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)

		// Control point (slight curve)
		cx := startX + (targetX-startX)/2 + (b.rand.Intn(100) - 50)
		cy := startY + (targetY-startY)/2 + (b.rand.Intn(100) - 50)

		// Quadratic Bezier curve
		x := int((1-t)*(1-t)*float64(startX) + 2*(1-t)*t*float64(cx) + t*t*float64(targetX))
		y := int((1-t)*(1-t)*float64(startY) + 2*(1-t)*t*float64(cy) + t*t*float64(targetY))

		b.page.Mouse.MustMove(x, y)
		time.Sleep(time.Duration(5+b.rand.Intn(11)) * time.Millisecond)
	}

	// Add slight overshoot
	overX := 10 + b.rand.Intn(11) // 10-20 pixels
	overY := 10 + b.rand.Intn(11)
	b.page.Mouse.MustMove(targetX+overX, targetY+overY)

	// Correct back to target
	time.Sleep(time.Duration(50+b.rand.Intn(51)) * time.Millisecond)
	b.page.Mouse.MustMove(targetX, targetY)

	return nil
}

// randomMouseMovement moves mouse to random position
func (b *LinkedInBot) randomMouseMovement() {
	if !b.config.Stealth.MouseMovement {
		return
	}

	x := b.rand.Intn(b.config.Browser.Viewport.Width)
	y := b.rand.Intn(b.config.Browser.Viewport.Height)
	b.page.Mouse.MustMove(x, y)
	time.Sleep(time.Duration(100+b.rand.Intn(401)) * time.Millisecond)
}

// Login logs into LinkedIn
func (b *LinkedInBot) Login() error {
	// Respect business hours if configured
	if b.config.Stealth.BusinessHoursOnly {
		b.waitForBusinessHours()
	}

	b.logger.Info("Navigating to LinkedIn...")
	_, err := b.page.EnableDomain(proto.NetworkEnable{})
	if err != nil {
		return fmt.Errorf("failed to enable network domain: %v", err)
	}

	// Set navigation timeout from config
	b.page = b.page.Context(
		b.page.GetContext(),
		b.page.Timeout(time.Duration(b.config.Timing.NavigationTimeout)*time.Second),
	)

	// Navigate to login page
	err = rod.Try(func() {
		b.page.Navigate("https://www.linkedin.com/login")
		b.page.WaitLoad()
	})

	if err != nil {
		return fmt.Errorf("failed to navigate to login page: %v", err)
	}

	// Add some human-like delays and movements
	b.randomDelay()
	b.randomMouseMovement()

	// Find and interact with email field
	emailField := b.page.MustElement("#username")
	if emailField == nil {
		return fmt.Errorf("email field not found")
	}

	b.moveMouseToElement(emailField)
	emailField.MustClick()
	b.randomDelay()

	// Type email
	if err := b.humanType(emailField, b.email); err != nil {
		return fmt.Errorf("failed to type email: %v", err)
	}

	b.randomDelay(1, 2)

	// Find and interact with password field
	passwordField := b.page.MustElement("#password")
	if passwordField == nil {
		return fmt.Errorf("password field not found")
	}

	b.moveMouseToElement(passwordField)
	passwordField.MustClick()
	b.randomDelay()

	// Type password
	if err := b.humanType(passwordField, b.password); err != nil {
		return fmt.Errorf("failed to type password: %v", err)
	}

	b.randomDelay()

	// Find and click login button
	b.logger.Info("Clicking login button...")
	loginButton := b.page.MustElement("button[type='submit']")
	if loginButton == nil {
		return fmt.Errorf("login button not found")
	}

	b.moveMouseToElement(loginButton)
	loginButton.MustClick()

	// Wait for navigation
	time.Sleep(5 * time.Second)

	// Check if login was successful
	currentURL := b.page.MustInfo().URL
	if strings.Contains(currentURL, "/feed") || strings.Contains(currentURL, "/in/") {
		b.logger.Info("Login successful!")
		return nil
	} else if strings.Contains(currentURL, "/checkpoint") || strings.Contains(currentURL, "/challenge") {
		return fmt.Errorf("login requires additional verification")
	}

	return fmt.Errorf("login failed, unknown page: %s", currentURL)
}

// SearchPeople searches for LinkedIn profiles
func (b *LinkedInBot) SearchPeople(jobTitle, location string, maxResults int) ([]string, error) {
	if b.config.Stealth.BusinessHoursOnly && !b.isBusinessHours() {
		b.waitForBusinessHours()
	}

	b.logger.Infof("Searching for '%s' in '%s'...", jobTitle, location)

	// Check daily connection limit
	if b.storage != nil {
		count, err := b.storage.GetTodayConnectionCount()
		if err != nil {
			b.logger.Warnf("Failed to get today's connection count: %v", err)
		} else if count >= b.config.Automation.DailyConnectionLimit {
			return nil, fmt.Errorf("daily connection limit reached (%d/%d)",
				count, b.config.Automation.DailyConnectionLimit)
		}
	}

	// Build search URL
	searchQuery := fmt.Sprintf("%s %s", jobTitle, location)
	searchURL := fmt.Sprintf("https://www.linkedin.com/search/results/people/?keywords=%s", searchQuery)

	// Navigate to search results
	err := rod.Try(func() {
		b.page.Navigate(searchURL)
		b.page.WaitLoad()
	})

	if err != nil {
		return nil, fmt.Errorf("failed to navigate to search page: %v", err)
	}

	// Wait for page load and add some delay
	b.page.WaitLoad()
	b.randomDelay()

	// Scroll to load more results
	for i := 0; i < 3; i++ {
		b.humanScroll()
		b.randomDelay()

		// 50% chance of random mouse movement
		if b.config.Stealth.MouseMovement && b.rand.Float64() < 0.5 {
			b.randomMouseMovement()
		}
	}

	// Find profile links
	var profiles []string
	seen := make(map[string]bool)

	elements, err := b.page.Elements("a.app-aware-link")
	if err != nil {
		return nil, fmt.Errorf("failed to find profile links: %v", err)
	}

	for _, element := range elements {
		if len(profiles) >= maxResults {
			break
		}

		href, err := element.Property("href")
		if err != nil {
			continue
		}

		url := href.String()
		if strings.Contains(url, "/in/") && !strings.Contains(url, "/search/") {
			if !seen[url] {
				// Check if we've already sent a connection request
				if b.storage != nil {
					hasSent, err := b.storage.HasSentConnectionRequest(url)
					if err != nil {
						b.logger.Warnf("Failed to check connection status for %s: %v", url, err)
						continue
					}
					if hasSent {
						b.logger.Debugf("Already sent connection to %s", url)
						continue
					}
				}

				profiles = append(profiles, url)
				seen[url] = true
			}
		}
	}

	b.logger.Infof("Found %d profiles", len(profiles))
	return profiles, nil
}

// isBusinessHours checks if current time is within business hours (Mon-Fri, 9-17)
func (b *LinkedInBot) isBusinessHours() bool {
	if !b.config.Stealth.BusinessHoursOnly {
		return true
	}

	now := time.Now()
	hour := now.Hour()
	weekday := now.Weekday()

	// Not business hours if weekend
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// Business hours are 9 AM to 5 PM
	return hour >= 9 && hour < 17
}

// waitForBusinessHours waits until business hours
func (b *LinkedInBot) waitForBusinessHours() {
	if !b.config.Stealth.BusinessHoursOnly {
		return
	}

	for !b.isBusinessHours() {
		now := time.Now()
		nextHour := now.Truncate(time.Hour).Add(time.Hour)
		sleepTime := time.Until(nextHour)

		b.logger.Infof("Outside business hours, waiting until %s...", nextHour.Format("15:04"))
		time.Sleep(sleepTime)
	}
}
