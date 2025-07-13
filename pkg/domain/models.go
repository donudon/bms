package domain

import "time"

// --- ENUMS & CONSTANTS ---

type LoanStatus string
const (
	Active LoanStatus = "ACTIVE"
	Closed LoanStatus = "CLOSED"
)

type TermType string
const (
	Weekly TermType = "WEEKLY"
)

type InstallmentStatus string
const (
	Pending   InstallmentStatus = "PENDING"
	Paid      InstallmentStatus = "PAID"
)

type BillingStatus string
const (
	PendingBilling BillingStatus = "PENDING"
	BilledBilling  BillingStatus = "BILLED"
	PaidBilling    BillingStatus = "PAID"
)

// --- DATA MODELS (Structs) ---

type Loan struct {
	ID                string
	CustomerID        string
	PrincipalAmount   float64
	InterestRate      float64
	Term              int
	TermType          TermType
	TermPayment       float64
	OutstandingAmount float64 // Stored state
	IsDelinquent      bool    // Stored state
	Status            LoanStatus
	CreationDate      time.Time
}

type Installment struct {
	ID          string
	LoanID      string
	TermNumber  int
	AmountDue   float64
	DueDate     time.Time
	BilledDate  time.Time
	Status      InstallmentStatus
}

type Billing struct {
	ID            string
	InstallmentID string
	Amount        float64
	BillingDate   time.Time
	Status        BillingStatus
}

type Payment struct {
	ID          string
	BillingID   string
	AmountPaid  float64
	PaymentDate time.Time
}