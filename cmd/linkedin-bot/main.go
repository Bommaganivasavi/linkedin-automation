package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type LinkedInBot struct {
	browser *rod.Browser
	page    *rod.Page
	logger  *logrus.Logger
	rand    *rand.Rand
}

func NewLinkedInBot() *LinkedInBot {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	return &LinkedInBot{
		logger: logger,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *LinkedInBot) Start() error {
	// Initialize browser
	launch := launcher.New().
		Headless(false).
		Set("disable-blink-features", "AutomationControlled").
		Set("--disable-web-security", "").
		Set("--disable-dev-shm-usage", "").
		Set("--no-sandbox", "").
		Leakless(false)

	url := launch.MustLaunch()
	b.browser = rod.New().ControlURL(url).MustConnect()
	b.page = stealth.MustPage(b.browser)

	// Set viewport
	b.page.MustSetViewport(1920, 1080, 1, false)

	// Set user agent
	b.page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	})

	return nil
}

func (b *LinkedInBot) Login(email, password string) error {
	b.logger.Info("Navigating to LinkedIn login page...")
	b.page.MustNavigate("https://www.linkedin.com/login")
	b.page.MustWaitLoad()

	b.logger.Info("Logging in...")
	b.page.MustElement("#username").MustInput(email)
	b.page.MustElement("#password").MustInput(password)
	b.page.MustElement("button[type='submit']").MustClick()
	b.page.MustWaitLoad()

	// Check for 2FA
	if b.page.MustHas("#input__phone_verification_pin") {
		b.logger.Info("2FA detected. Please enter the code in the browser.")
		b.page.MustElement("#input__phone_verification_pin").MustWaitVisible()
		b.page.MustWaitNavigation()
	}

	b.logger.Info("Successfully logged in")
	return nil
}

func (b *LinkedInBot) SearchPeople(jobTitle, location string, maxResults int) ([]string, error) {
	b.logger.Infof("Searching for '%s' in '%s'...", jobTitle, location)

	// Build search URL
	searchQuery := jobTitle
	if location != "" {
		searchQuery += " " + location
	}
	searchURL := "https://www.linkedin.com/search/results/people/?keywords=" + searchQuery

	// Navigate to search results
	b.page.MustNavigate(searchURL)
	b.page.MustWaitLoad()
	b.randomDelay(3, 5)

	// Find profile links
	var profiles []string
	elements := b.page.MustElements(".entity-result__title-text a.app-aware-link")
	for _, el := range elements {
		if len(profiles) >= maxResults {
			break
		}
		href := el.MustAttribute("href")
		if href != nil && *href != "" {
			profiles = append(profiles, *href)
		}
	}

	return profiles, nil
}

func (b *LinkedInBot) SendConnectionRequest(profileURL, message string) error {
	b.logger.Infof("Navigating to profile: %s", profileURL)
	b.page.MustNavigate(profileURL)
	b.page.MustWaitLoad()
	b.randomDelay(3, 5)

	// Click the "Connect" button
	connectBtn := b.page.MustElement("button:has-text('Connect')")
	connectBtn.MustClick()
	b.randomDelay(1, 2)

	// Add a note if message is provided
	if message != "" {
		addNoteBtn, err := b.page.Element("button:has-text('Add a note')")
		if err == nil {
			addNoteBtn.MustClick()
			b.randomDelay(1, 2)

			// Type the message
			noteTextarea, err := b.page.Element("textarea[name='message']")
			if err == nil {
				b.humanType(noteTextarea, message)
				b.randomDelay(1, 2)
			}
		}
	}

	// Send the connection request
	sendBtn := b.page.MustElement("button:has-text('Send')")
	sendBtn.MustClick()
	b.randomDelay(2, 3)

	b.logger.Info("Connection request sent successfully")
	return nil
}

func (b *LinkedInBot) humanType(element *rod.Element, text string) {
	for _, char := range text {
		element.MustInput(string(char))
		time.Sleep(time.Duration(50+b.rand.Intn(100)) * time.Millisecond)
	}
}

func (b *LinkedInBot) randomDelay(minMs, maxMs int) {
	delay := time.Duration(minMs+b.rand.Intn(maxMs-minMs+1)) * time.Millisecond
	time.Sleep(delay)
}

func (b *LinkedInBot) Close() {
	if b.page != nil {
		b.page.MustClose()
	}
	if b.browser != nil {
		b.browser.MustClose()
	}
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Initialize bot
	bot := NewLinkedInBot()
	defer bot.Close()

	// Start the browser
	if err := bot.Start(); err != nil {
		logger.Fatalf("Failed to start browser: %v", err)
	}

	// Login
	email := os.Getenv("LINKEDIN_EMAIL")
	password := os.Getenv("LINKEDIN_PASSWORD")
	if err := bot.Login(email, password); err != nil {
		logger.Fatalf("Login failed: %v", err)
	}

	// Search for people and send connection requests
	searchTerm := "Software Engineer"
	location := "India"
	maxResults := 3 // Start with a small number for testing
	message := "Hi, I came across your profile and would like to connect."

	logger.Infof("Searching for %s in %s...", searchTerm, location)
	profiles, err := bot.SearchPeople(searchTerm, location, maxResults)
	if err != nil {
		logger.Fatalf("Failed to search people: %v", err)
	}

	logger.Infof("Found %d profiles", len(profiles))

	// Send connection requests
	for i, profile := range profiles {
		logger.Infof("Sending connection request to profile %d/%d", i+1, len(profiles))
		if err := bot.SendConnectionRequest(profile, message); err != nil {
			logger.Errorf("Failed to send connection request: %v", err)
			continue
		}
		// Random delay between requests (30-90 seconds)
		delay := 30 + rand.Intn(61)
		logger.Infof("Waiting %d seconds before next request...", delay)
		time.Sleep(time.Duration(delay) * time.Second)
	}

	logger.Info("Bot has completed all tasks. Keeping browser open for 5 minutes...")
	time.Sleep(5 * time.Minute)
}
