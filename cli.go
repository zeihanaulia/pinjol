package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"pinjol/pkg/domain"
	_ "github.com/mattn/go-sqlite3"
)

func runCLI() {
	// Skip the "cli" argument
	args := flag.NewFlagSet("cli", flag.ExitOnError)
	var (
		scenario  = args.String("scenario", "", "Scenario to run: ontime, skip2, fullpay")
		principal = args.Int64("principal", 5_000_000, "Loan principal amount")
		rate      = args.Float64("rate", 0.10, "Annual interest rate")
		startDate = args.String("start", "2025-08-15", "Start date (YYYY-MM-DD)")
		now       = args.String("now", "", "Current date override (YYYY-MM-DD)")
		repeat    = args.Int("repeat", 1, "Number of payments to make")
		verbose   = args.Bool("verbose", false, "Verbose output")
		dbPath    = args.String("db", "./pinjol.db", "Database path")
	)
	if err := args.Parse(os.Args[2:]); err != nil { // Skip "program" and "cli"
		log.Fatalf("failed to parse flags: %v", err)
	}

	if *scenario == "" {
		log.Fatal("Please specify a scenario: ontime, skip2, or fullpay")
	}

	start, err := time.Parse("2006-01-02", *startDate)
	if err != nil {
		log.Fatalf("Invalid start date: %v", err)
	}

	currentTime := time.Now()
	if *now != "" {
		currentTime, err = time.Parse("2006-01-02", *now)
		if err != nil {
			log.Fatalf("Invalid now date: %v", err)
		}
	}

	// Initialize database for CLI
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	if err := InitDatabase(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create repository
	repo := NewSQLiteLoanRepository(db)

	// Create loan
	loan, err := domain.NewLoan("cli-test", *principal, *rate, start)
	if err != nil {
		log.Fatalf("Failed to create loan: %v", err)
	}

	// Save loan to database
	if err := repo.Create(loan); err != nil {
		log.Fatalf("Failed to save loan to database: %v", err)
	}

	switch *scenario {
	case "ontime":
		runOntimeScenario(repo, currentTime, *repeat, *verbose)
	case "skip2":
		runSkip2Scenario(repo, start, *verbose)
	case "fullpay":
		runFullPayScenario(repo, currentTime, *verbose)
	default:
		log.Fatalf("Unknown scenario: %s", *scenario)
	}

	// Get final state from database
	finalLoan, err := repo.GetByID("cli-test")
	if err != nil {
		log.Fatalf("Failed to get final loan state: %v", err)
	}

	// Output final state
	delinquent, streak, observedWeek := finalLoan.IsDelinquent(currentTime)
	finalState := map[string]interface{}{
		"loan":          finalLoan,
		"delinquent":    delinquent,
		"streak":        streak,
		"observed_week": observedWeek,
	}

	output, err := json.MarshalIndent(finalState, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal final state: %v", err)
	}
	fmt.Println(string(output))
}

func runOntimeScenario(repo LoanRepository, startTime time.Time, repeat int, verbose bool) {
	fmt.Println("=== On-time Payment Scenario ===")

	for i := 0; i < repeat && i < 50; i++ {
		// Get current loan state from database
		loan, err := repo.GetByID("cli-test")
		if err != nil {
			log.Printf("Failed to get loan for payment %d: %v", i+1, err)
			break
		}

		// Simulate payment at the right time
		paymentTime := startTime.Add(time.Duration(i*7) * 24 * time.Hour)

		err = loan.MakePayment(loan.WeeklyDue, paymentTime)
		if err != nil {
			log.Printf("Payment %d failed: %v", i+1, err)
			break
		}

		// Save updated loan to database
		if err := repo.Update(loan); err != nil {
			log.Printf("Failed to save payment %d: %v", i+1, err)
			break
		}

		outstanding := loan.GetOutstanding()
		delinquent, streak, observedWeek := loan.IsDelinquent(paymentTime)

		if verbose {
			fmt.Printf("Payment %d: outstanding=%d, delinquent=%v, streak=%d, observed_week=%d\n",
				i+1, outstanding, delinquent, streak, observedWeek)
		}
	}
}

func runSkip2Scenario(repo LoanRepository, startDate time.Time, verbose bool) {
	fmt.Println("=== Skip 2 Weeks Scenario ===")

	// Simulate being 14 days after start (week 3)
	checkTime := startDate.Add(14 * 24 * time.Hour)

	loan, err := repo.GetByID("cli-test")
	if err != nil {
		log.Fatalf("Failed to get loan: %v", err)
	}

	delinquent, streak, observedWeek := loan.IsDelinquent(checkTime)
	fmt.Printf("Before payments: delinquent=%v, streak=%d, observed_week=%d\n",
		delinquent, streak, observedWeek)

	if !delinquent {
		log.Fatal("Expected to be delinquent, but was not")
	}

	// Make catch-up payments
	fmt.Println("Making catch-up payments...")
	for i := 0; i < 2; i++ {
		// Get fresh loan state
		loan, err := repo.GetByID("cli-test")
		if err != nil {
			log.Printf("Failed to get loan for catch-up payment %d: %v", i+1, err)
			break
		}

		err = loan.MakePayment(loan.WeeklyDue, checkTime)
		if err != nil {
			log.Printf("Catch-up payment %d failed: %v", i+1, err)
			break
		}

		// Save updated loan
		if err := repo.Update(loan); err != nil {
			log.Printf("Failed to save catch-up payment %d: %v", i+1, err)
			break
		}

		if verbose {
			fmt.Printf("Catch-up payment %d completed\n", i+1)
		}
	}

	// Check delinquency again
	loan, err = repo.GetByID("cli-test")
	if err != nil {
		log.Fatalf("Failed to get final loan state: %v", err)
	}

	delinquent, streak, observedWeek = loan.IsDelinquent(checkTime)
	fmt.Printf("After catch-up: delinquent=%v, streak=%d, observed_week=%d\n",
		delinquent, streak, observedWeek)
}

func runFullPayScenario(repo LoanRepository, currentTime time.Time, verbose bool) {
	fmt.Println("=== Full Payment Scenario ===")

	// Pay all 50 weeks
	for i := 0; i < 50; i++ {
		// Get current loan state
		loan, err := repo.GetByID("cli-test")
		if err != nil {
			log.Printf("Failed to get loan for payment %d: %v", i+1, err)
			break
		}

		err = loan.MakePayment(loan.WeeklyDue, currentTime)
		if err != nil {
			log.Printf("Payment %d failed: %v", i+1, err)
			break
		}

		// Save updated loan
		if err := repo.Update(loan); err != nil {
			log.Printf("Failed to save payment %d: %v", i+1, err)
			break
		}

		if verbose && (i+1)%10 == 0 {
			fmt.Printf("Completed %d payments\n", i+1)
		}
	}

	// Get final state
	loan, err := repo.GetByID("cli-test")
	if err != nil {
		log.Fatalf("Failed to get final loan state: %v", err)
	}

	outstanding := loan.GetOutstanding()
	fmt.Printf("After 50 payments: outstanding=%d\n", outstanding)

	if outstanding != 0 {
		log.Fatal("Expected outstanding to be 0")
	}

	// Try to make an extra payment
	fmt.Println("Attempting extra payment...")
	err = loan.MakePayment(loan.WeeklyDue, currentTime)
	if err != nil {
		fmt.Printf("Extra payment correctly rejected: %v\n", err)
	} else {
		log.Fatal("Extra payment should have been rejected")
	}
}
