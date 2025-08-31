package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "github.com/mattn/go-sqlite3"

	"pinjol/internal"
	"pinjol/pkg/domain"
)

var scenarioCmd = &cobra.Command{
	Use:   "scenario",
	Short: "Run loan payment scenarios",
	Long:  `Run various loan payment scenarios for testing.`,
}

var ontimeCmd = &cobra.Command{
	Use:   "ontime",
	Short: "Run on-time payment scenario",
	Run: func(cmd *cobra.Command, args []string) {
		runOntimeScenario()
	},
}

var skip2Cmd = &cobra.Command{
	Use:   "skip2",
	Short: "Run skip 2 weeks scenario",
	Run: func(cmd *cobra.Command, args []string) {
		runSkip2Scenario()
	},
}

var fullpayCmd = &cobra.Command{
	Use:   "fullpay",
	Short: "Run full payment scenario",
	Run: func(cmd *cobra.Command, args []string) {
		runFullPayScenario()
	},
}

func init() {
	scenarioCmd.AddCommand(ontimeCmd)
	scenarioCmd.AddCommand(skip2Cmd)
	scenarioCmd.AddCommand(fullpayCmd)

	// Common flags
	scenarioCmd.PersistentFlags().Int64("principal", 5_000_000, "Loan principal amount")
	scenarioCmd.PersistentFlags().Float64("rate", 0.10, "Annual interest rate")
	scenarioCmd.PersistentFlags().String("start", "2025-08-15", "Start date (YYYY-MM-DD)")
	scenarioCmd.PersistentFlags().String("now", "", "Current date override (YYYY-MM-DD)")
	scenarioCmd.PersistentFlags().StringP("db-path", "d", "./pinjol.db", "Database path")
	scenarioCmd.PersistentFlags().Bool("verbose", false, "Verbose output")

	// Specific flags
	ontimeCmd.Flags().Int("repeat", 1, "Number of payments to make")

	viper.BindPFlags(scenarioCmd.PersistentFlags())
	viper.BindPFlags(ontimeCmd.Flags())
}

func initScenarioDB() domain.LoanRepository {
	dbPath := viper.GetString("db-path")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := internal.InitDatabase(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	repo := internal.NewSQLiteLoanRepository(db)

	principal := viper.GetInt64("principal")
	rate := viper.GetFloat64("rate")
	startDate := viper.GetString("start")

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		log.Fatalf("Invalid start date: %v", err)
	}

	loan, err := domain.NewLoan("cli-test", principal, rate, start)
	if err != nil {
		log.Fatalf("Failed to create loan: %v", err)
	}

	if err := repo.Create(loan); err != nil {
		log.Fatalf("Failed to save loan to database: %v", err)
	}

	return repo
}

func runOntimeScenario() {
	fmt.Println("=== On-time Payment Scenario ===")

	repo := initScenarioDB()
	repeat := viper.GetInt("repeat")
	verbose := viper.GetBool("verbose")
	nowStr := viper.GetString("now")

	startTime := time.Now()
	if nowStr != "" {
		var err error
		startTime, err = time.Parse("2006-01-02", nowStr)
		if err != nil {
			log.Fatalf("Invalid now date: %v", err)
		}
	}

	for i := 0; i < repeat && i < 50; i++ {
		loan, err := repo.GetByID("cli-test")
		if err != nil {
			log.Printf("Failed to get loan for payment %d: %v", i+1, err)
			break
		}

		paymentTime := startTime.Add(time.Duration(i*7) * 24 * time.Hour)

		err = loan.MakePayment(loan.WeeklyDue, paymentTime)
		if err != nil {
			log.Printf("Payment %d failed: %v", i+1, err)
			break
		}

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

	finalOutput()
}

func runSkip2Scenario() {
	fmt.Println("=== Skip 2 Weeks Scenario ===")

	repo := initScenarioDB()
	startDate := viper.GetString("start")
	verbose := viper.GetBool("verbose")

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		log.Fatalf("Invalid start date: %v", err)
	}

	checkTime := start.Add(14 * 24 * time.Hour)

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

	fmt.Println("Making catch-up payments...")
	for i := 0; i < 2; i++ {
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

		if err := repo.Update(loan); err != nil {
			log.Printf("Failed to save catch-up payment %d: %v", i+1, err)
			break
		}

		if verbose {
			fmt.Printf("Catch-up payment %d completed\n", i+1)
		}
	}

	loan, err = repo.GetByID("cli-test")
	if err != nil {
		log.Fatalf("Failed to get final loan state: %v", err)
	}

	delinquent, streak, observedWeek = loan.IsDelinquent(checkTime)
	fmt.Printf("After catch-up: delinquent=%v, streak=%d, observed_week=%d\n",
		delinquent, streak, observedWeek)

	finalOutput()
}

func runFullPayScenario() {
	fmt.Println("=== Full Payment Scenario ===")

	repo := initScenarioDB()
	verbose := viper.GetBool("verbose")
	nowStr := viper.GetString("now")

	currentTime := time.Now()
	if nowStr != "" {
		var err error
		currentTime, err = time.Parse("2006-01-02", nowStr)
		if err != nil {
			log.Fatalf("Invalid now date: %v", err)
		}
	}

	for i := 0; i < 50; i++ {
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

		if err := repo.Update(loan); err != nil {
			log.Printf("Failed to save payment %d: %v", i+1, err)
			break
		}

		if verbose && (i+1)%10 == 0 {
			fmt.Printf("Completed %d payments\n", i+1)
		}
	}

	loan, err := repo.GetByID("cli-test")
	if err != nil {
		log.Fatalf("Failed to get final loan state: %v", err)
	}

	outstanding := loan.GetOutstanding()
	fmt.Printf("After 50 payments: outstanding=%d\n", outstanding)

	if outstanding != 0 {
		log.Fatal("Expected outstanding to be 0")
	}

	fmt.Println("Attempting extra payment...")
	err = loan.MakePayment(loan.WeeklyDue, currentTime)
	if err != nil {
		fmt.Printf("Extra payment correctly rejected: %v\n", err)
	} else {
		log.Fatal("Extra payment should have been rejected")
	}

	finalOutput()
}

func finalOutput() {
	dbPath := viper.GetString("db-path")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := internal.NewSQLiteLoanRepository(db)
	finalLoan, err := repo.GetByID("cli-test")
	if err != nil {
		log.Fatalf("Failed to get final loan state: %v", err)
	}

	nowStr := viper.GetString("now")
	currentTime := time.Now()
	if nowStr != "" {
		currentTime, _ = time.Parse("2006-01-02", nowStr)
	}

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
