package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"bms/internal/domain"
	"bms/internal/handler"
	"bms/internal/repository"
	"bms/internal/usecase"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	// --- Database Connection ---
	connStr := "user=postgres password=mysecretpassword dbname=billing sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")

	// 1. Initialize Repository (Data Layer)
	loanRepo := repository.NewPostgresRepository(db)

	// 2. Initialize Usecase (Business Logic Layer)
	billingUsecase := usecase.NewBillingUsecase(loanRepo)

	// 3. Initialize Handler (Presentation Layer)
	loanHandler := handler.NewLoanHandler(billingUsecase)

	// --- Simulation ---
	fmt.Println("\n--- 1. Create Loan ---")
	loan, err := loanHandler.CreateLoan("cust_007", 5000000, 0.10, 50, domain.Weekly)
	if err != nil {
		log.Fatalf("Failed to create loan: %v", err)
	}
	fmt.Println("Loan created successfully.")

	fmt.Println("\n--- 2. Simulate 3 weeks passing ---")
	// In a real app, you wouldn't do this. This is just for the simulation.
	if pgRepo, ok := loanRepo.(*repository.PostgresRepository); ok {
		pgRepo.ManuallyAdvanceTime(loan.ID, 3)
		fmt.Println("Advanced time for 3 installments.")
	}

	fmt.Println("\n--- 3. Run Billing Job (should create 3 bills) ---")
	billsCreated, err := loanHandler.CreateBilling()
	if err != nil {
		log.Fatalf("Billing job failed: %v", err)
	}
	fmt.Printf("Billing job completed. Created %d bills.\n", billsCreated)

	fmt.Println("\n--- 4. Get Loan Info (should show delinquent) ---")
	info, err := loanHandler.GetLoanInfo(loan.ID)
	if err != nil {
		log.Fatalf("Failed to get loan info: %v", err)
	}
	fmt.Printf("Loan Info Fetched: Outstanding=%.2f, Delinquent=%v\n", info.OutstandingAmount, info.IsDelinquent)

	fmt.Println("\n--- 5. Make a payment for 2 terms ---")
	paymentAmount := info.TermPayment * 2
	fmt.Printf("Customer pays %.2f...\n", paymentAmount)
	msg, err := loanHandler.MakePayment(loan.ID, paymentAmount)
	if err != nil {
		log.Fatalf("Failed to make payment: %v", err)
	}
	fmt.Println(msg)

	fmt.Println("\n--- 6. Get Loan Info After Payment ---")
	info, err = loanHandler.GetLoanInfo(loan.ID)
	if err != nil {
		log.Fatalf("Failed to get loan info: %v", err)
	}
	fmt.Printf("Loan Info Fetched: Outstanding=%.2f, Delinquent=%v\n", info.OutstandingAmount, info.IsDelinquent)
}