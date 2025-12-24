LinkedIn Automation Tool

Hey there! 👋 This is a LinkedIn automation tool I built using Go and the Rod browser automation library. It's designed to demonstrate advanced browser automation techniques while maintaining a low profile.⚠️ Heads Up Before We Start
Just a quick note - this is purely for educational purposes. LinkedIn doesn't allow automation, so please don't use this on real accounts (you could get banned). I built this to learn about browser automation and anti-detection techniques.

✨ What Can It Do?
- Log into LinkedIn (handles 2FA too!)
- Search for profiles based on job titles and locations
- Send connection requests (with personalized messages)
- Mimic human behavior to avoid detection
- Keep track of sent requests

 🛠️ Getting Started

What You'll Need
- Go 1.20 or later
- Chrome or Chromium browser

Let's Get It Running
1. First, clone the repo:
   ```bash
   git clone https://github.com/Bommaganivasavi/linkedin-automation.git
   cd linkedin-automation
   ```

2. Grab the dependencies:
   ```bash
   go mod download
   ```

3. Set up your [.env](cci:7://file:///c:/Users/DELL/linkedin-automation/.env:0:0-0:0) file:
   - Copy `.env.example` to [.env](cci:7://file:///c:/Users/DELL/linkedin-automation/.env:0:0-0:0)
   - Add your LinkedIn email and password

4. Run it!
   ```bash
   go run cmd/linkedin-bot/main.go
   ```
 🏗️ How It's Built

### Main Components
- **`cmd/linkedin-bot/`** - Where the magic starts
- **`internal/config/`** - Handles all the settings
- **`internal/storage/`** - Keeps track of connections

Cool Tech Stuff
- Uses Rod for browser automation (like Puppeteer but for Go)
- Implements human-like delays and behaviors
- Has built-in rate limiting

 🤔 Why Did I Build This?
I wanted to learn more about:
- Browser automation
- Anti-bot detection techniques
- Building CLI tools in Go
- Handling authentication flows

## 📝 Notes
- This is a work in progress
- Use responsibly and ethically
- Feel free to contribute or fork!

## 📜 License
MIT - Do whatever you want with it, but no guarantees!

---

Built with ❤️ by [Vasavi Bommagani] - Hope you find this useful for learning!
