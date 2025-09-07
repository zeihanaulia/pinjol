package main

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := InitDatabase(db); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	return db
}

func TestSQLiteLoanRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteLoanRepository(db)

	startDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	loan, err := NewLoan("test-loan-1", 1000000, 0.1, startDate)
	if err != nil {
		t.Fatalf("Failed to create loan: %v", err)
	}

	err = repo.Create(loan)
	if err != nil {
		t.Fatalf("Failed to create loan in repository: %v", err)
	}

	// Verify loan was created
	retrieved, err := repo.GetByID("test-loan-1")
	if err != nil {
		t.Fatalf("Failed to retrieve created loan: %v", err)
	}

	if retrieved.ID != loan.ID {
		t.Errorf("Expected ID %s, got %s", loan.ID, retrieved.ID)
	}
	if retrieved.Principal != loan.Principal {
		t.Errorf("Expected principal %d, got %d", loan.Principal, retrieved.Principal)
	}
}

func TestSQLiteLoanRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteLoanRepository(db)

	startDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	loan, err := NewLoan("test-loan-2", 2000000, 0.15, startDate)
	if err != nil {
		t.Fatalf("Failed to create loan: %v", err)
	}

	err = repo.Create(loan)
	if err != nil {
		t.Fatalf("Failed to create loan: %v", err)
	}

	// Test successful retrieval
	retrieved, err := repo.GetByID("test-loan-2")
	if err != nil {
		t.Fatalf("Failed to get loan: %v", err)
	}

	if retrieved.ID != "test-loan-2" {
		t.Errorf("Expected ID test-loan-2, got %s", retrieved.ID)
	}

	// Test non-existent loan
	_, err = repo.GetByID("non-existent")
	if err != ErrLoanNotFound {
		t.Errorf("Expected ErrLoanNotFound, got %v", err)
	}
}

func TestSQLiteLoanRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteLoanRepository(db)

	startDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	loan, err := NewLoan("test-loan-3", 1500000, 0.12, startDate)
	if err != nil {
		t.Fatalf("Failed to create loan: %v", err)
	}

	err = repo.Create(loan)
	if err != nil {
		t.Fatalf("Failed to create loan: %v", err)
	}

	// Make a payment
	now := time.Now()
	err = loan.MakePayment(loan.WeeklyDue, now)
	if err != nil {
		t.Fatalf("Failed to make payment: %v", err)
	}

	// Update the loan
	err = repo.Update(loan)
	if err != nil {
		t.Fatalf("Failed to update loan: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID("test-loan-3")
	if err != nil {
		t.Fatalf("Failed to retrieve updated loan: %v", err)
	}

	if retrieved.PaidCount != 1 {
		t.Errorf("Expected PaidCount 1, got %d", retrieved.PaidCount)
	}
	if retrieved.Schedule[0].Paid != true {
		t.Errorf("Expected first week to be paid")
	}
}

func TestSQLiteLoanRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteLoanRepository(db)

	startDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)

	// Create multiple loans
	for i := 1; i <= 3; i++ {
		loan, err := NewLoan(fmt.Sprintf("test-loan-%d", i), int64(i*1000000), 0.1, startDate)
		if err != nil {
			t.Fatalf("Failed to create loan %d: %v", i, err)
		}
		err = repo.Create(loan)
		if err != nil {
			t.Fatalf("Failed to create loan %d in repo: %v", i, err)
		}
	}

	loans, err := repo.List()
	if err != nil {
		t.Fatalf("Failed to list loans: %v", err)
	}

	if len(loans) != 3 {
		t.Errorf("Expected 3 loans, got %d", len(loans))
	}
}

func TestSQLiteLoanRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteLoanRepository(db)

	startDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	loan, err := NewLoan("test-loan-delete", 1000000, 0.1, startDate)
	if err != nil {
		t.Fatalf("Failed to create loan: %v", err)
	}

	err = repo.Create(loan)
	if err != nil {
		t.Fatalf("Failed to create loan: %v", err)
	}

	// Delete the loan
	err = repo.Delete("test-loan-delete")
	if err != nil {
		t.Fatalf("Failed to delete loan: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID("test-loan-delete")
	if err != ErrLoanNotFound {
		t.Errorf("Expected ErrLoanNotFound after deletion, got %v", err)
	}

	// Test deleting non-existent loan
	err = repo.Delete("non-existent")
	if err != ErrLoanNotFound {
		t.Errorf("Expected ErrLoanNotFound for non-existent loan, got %v", err)
	}
}
