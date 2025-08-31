package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// User represents a simulated user
type User struct {
	ID       int
	Name     string
	Email    string
	Loans    []Loan
	Activity []string
}

// Loan represents a loan created by user
type Loan struct {
	ID        string
	Amount    int64
	WeeklyDue int64
	Status    string
}

// SimulationConfig holds simulation configuration
type SimulationConfig struct {
	Duration    time.Duration
	UserCount   int
	BaseURL     string
	MaxRequests int
}

// CreateLoanRequest represents loan creation request
type CreateLoanRequest struct {
	Principal  int64   `json:"principal"`
	AnnualRate float64 `json:"annual_rate"`
	StartDate  string  `json:"start_date"`
}

// PaymentRequest represents payment request
type PaymentRequest struct {
	Amount int64 `json:"amount"`
}

// Fake names for simulation
var firstNames = []string{"Ahmad", "Budi", "Citra", "Dewi", "Eko", "Fani", "Gilang", "Hana", "Iwan", "Joko"}
var lastNames = []string{"Putra", "Wijaya", "Sari", "Kusuma", "Rahman", "Lestari", "Pratama", "Nugroho", "Susanti", "Hartono"}

func main() {
	config := parseConfig()

	fmt.Printf("🚀 Starting Pinjol Smoke Test Simulation\n")
	fmt.Printf("⏱️  Duration: %v\n", config.Duration)
	fmt.Printf("👥 Users: %d\n", config.UserCount)
	fmt.Printf("🔗 Base URL: %s\n", config.BaseURL)
	fmt.Printf("📊 Max requests per user: %d\n", config.MaxRequests)
	fmt.Println("========================================")

	// Create users
	users := createUsers(config.UserCount)

	// Start simulation
	runSimulation(config, users)
}

func parseConfig() SimulationConfig {
	durationStr := getEnv("SIMULATION_DURATION", "30m")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Fatalf("Invalid duration format: %v", err)
	}

	userCount, _ := strconv.Atoi(getEnv("SIMULATION_USERS", "5"))
	maxRequests, _ := strconv.Atoi(getEnv("SIMULATION_MAX_REQUESTS", "10"))
	baseURL := getEnv("PINJOL_URL", "http://localhost:8081")

	return SimulationConfig{
		Duration:    duration,
		UserCount:   userCount,
		BaseURL:     baseURL,
		MaxRequests: maxRequests,
	}
}

func createUsers(count int) []*User {
	users := make([]*User, count)
	for i := 0; i < count; i++ {
		users[i] = &User{
			ID:    i + 1,
			Name:  fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))]),
			Email: fmt.Sprintf("user%d@example.com", i+1),
			Loans: []Loan{},
			Activity: []string{},
		}
	}
	return users
}

func runSimulation(config SimulationConfig, users []*User) {
	var wg sync.WaitGroup
	done := make(chan bool)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start users
	for _, user := range users {
		wg.Add(1)
		go func(u *User) {
			defer wg.Done()
			simulateUser(config, u, done)
		}(user)
	}

	// Timer for simulation duration
	timer := time.NewTimer(config.Duration)

	select {
	case <-timer.C:
		fmt.Println("\n⏰ Simulation duration reached")
	case <-sigChan:
		fmt.Println("\n🛑 Simulation interrupted by user")
	}

	// Stop all users
	close(done)

	// Wait for all users to finish
	wg.Wait()

	// Print summary
	printSummary(users)
}

func simulateUser(config SimulationConfig, user *User, done <-chan bool) {
	requestCount := 0

	for requestCount < config.MaxRequests {
		select {
		case <-done:
			return
		default:
			// Random delay between requests (1-10 seconds)
			time.Sleep(time.Duration(rand.Intn(9)+1) * time.Second)

			// Random action
			action := rand.Intn(4)
			switch action {
			case 0:
				createLoan(config, user)
			case 1:
				if len(user.Loans) > 0 {
					makePayment(config, user)
				} else {
					createLoan(config, user)
				}
			case 2:
				checkLoanStatus(config, user)
			case 3:
				checkHealth(config, user)
			}

			requestCount++
		}
	}
}

func createLoan(config SimulationConfig, user *User) {
	// Use good principal and rate combinations that result in integral weekly amounts
	goodLoans := []struct {
		principal int64
		rate      float64
	}{
		{5_000_000, 0.10}, // 5,500,000 / 50 = 110,000
		{3_000_000, 0.10}, // 3,300,000 / 50 = 66,000
		{2_000_000, 0.10}, // 2,200,000 / 50 = 44,000
		{4_000_000, 0.10}, // 4,400,000 / 50 = 88,000
		{5_000_000, 0.08}, // 5,400,000 / 50 = 108,000
		{3_000_000, 0.08}, // 3,240,000 / 50 = 64,800
		{1_000_000, 0.10}, // 1,100,000 / 50 = 22,000
	}
	
	selected := goodLoans[rand.Intn(len(goodLoans))]
	principal := selected.principal
	rate := selected.rate

	// Start date: today or tomorrow
	startDate := time.Now().AddDate(0, 0, rand.Intn(2)).Format("2006-01-02")

	request := CreateLoanRequest{
		Principal:  principal,
		AnnualRate: rate,
		StartDate:  startDate,
	}

	jsonData, _ := json.Marshal(request)

	url := fmt.Sprintf("%s/loans", config.BaseURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logActivity(user, fmt.Sprintf("❌ Failed to create loan: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		// Parse response to get loan ID
		var loanResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&loanResp)

		loanID := ""
		weeklyDue := int64(0)
		if id, ok := loanResp["ID"].(string); ok {
			loanID = id
		}
		if wd, ok := loanResp["WeeklyDue"].(float64); ok {
			weeklyDue = int64(wd)
		}

		loan := Loan{
			ID:        loanID,
			Amount:    principal,
			WeeklyDue: weeklyDue,
			Status:    "active",
		}
		user.Loans = append(user.Loans, loan)

		logActivity(user, fmt.Sprintf("✅ Created loan ID:%s Amount:Rp%d Rate:%.1f%%", loanID, principal, rate*100))
	} else {
		// Read response body for error details
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		logActivity(user, fmt.Sprintf("❌ Failed to create loan: HTTP %d - %s", resp.StatusCode, bodyString))
	}
}

func makePayment(config SimulationConfig, user *User) {
	if len(user.Loans) == 0 {
		return
	}

	// Pick random loan
	loan := user.Loans[rand.Intn(len(user.Loans))]

	// Use the actual WeeklyDue amount from loan
	paymentAmount := loan.WeeklyDue

	request := PaymentRequest{
		Amount: paymentAmount,
	}

	jsonData, _ := json.Marshal(request)

	url := fmt.Sprintf("%s/loans/%s/pay", config.BaseURL, loan.ID)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logActivity(user, fmt.Sprintf("❌ Failed to make payment for loan %s: %v", loan.ID, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logActivity(user, fmt.Sprintf("💰 Made payment for loan %s: Rp%d", loan.ID, paymentAmount))
	} else {
		// Read response body for error details
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		logActivity(user, fmt.Sprintf("❌ Failed to make payment for loan %s: HTTP %d - %s", loan.ID, resp.StatusCode, bodyString))
	}
}

func checkLoanStatus(config SimulationConfig, user *User) {
	if len(user.Loans) == 0 {
		return
	}

	// Pick random loan
	loan := user.Loans[rand.Intn(len(user.Loans))]

	url := fmt.Sprintf("%s/loans/%s", config.BaseURL, loan.ID)
	resp, err := http.Get(url)
	if err != nil {
		logActivity(user, fmt.Sprintf("❌ Failed to check loan %s status: %v", loan.ID, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logActivity(user, fmt.Sprintf("📊 Checked status for loan %s", loan.ID))
	} else {
		logActivity(user, fmt.Sprintf("❌ Failed to check loan %s status: HTTP %d", loan.ID, resp.StatusCode))
	}
}

func checkHealth(config SimulationConfig, user *User) {
	url := fmt.Sprintf("%s/healthz", config.BaseURL)
	resp, err := http.Get(url)
	if err != nil {
		logActivity(user, fmt.Sprintf("❌ Health check failed: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logActivity(user, "🏥 Health check passed")
	} else {
		logActivity(user, fmt.Sprintf("❌ Health check failed: HTTP %d", resp.StatusCode))
	}
}

func logActivity(user *User, activity string) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf("[%s] 👤 User %d (%s): %s", timestamp, user.ID, user.Name, activity)
	fmt.Println(message)

	user.Activity = append(user.Activity, message)
}

func printSummary(users []*User) {
	fmt.Println("\n========================================")
	fmt.Println("📊 SIMULATION SUMMARY")
	fmt.Println("========================================")

	totalLoans := 0
	totalPayments := 0
	totalActivities := 0

	for _, user := range users {
		totalLoans += len(user.Loans)
		totalPayments += countPayments(user)
		totalActivities += len(user.Activity)

		fmt.Printf("\n👤 User %d (%s):\n", user.ID, user.Name)
		fmt.Printf("   📋 Loans created: %d\n", len(user.Loans))
		fmt.Printf("   💰 Payments made: %d\n", countPayments(user))
		fmt.Printf("   📝 Total activities: %d\n", len(user.Activity))
	}

	fmt.Println("\n========================================")
	fmt.Printf("🎯 OVERALL STATISTICS:\n")
	fmt.Printf("   👥 Total users: %d\n", len(users))
	fmt.Printf("   📋 Total loans created: %d\n", totalLoans)
	fmt.Printf("   💰 Total payments made: %d\n", totalPayments)
	fmt.Printf("   📝 Total activities: %d\n", totalActivities)
	fmt.Println("========================================")
}

func countPayments(user *User) int {
	count := 0
	for _, activity := range user.Activity {
		if contains(activity, "Made payment") {
			count++
		}
	}
	return count
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
