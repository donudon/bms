package repository

import (
	"bms/internal/domain"
	"database/sql"
	"fmt"
	"time"
)

// LoanRepository defines the interface for data operations.
type LoanRepository interface {
	SaveLoan(loan *domain.Loan, schedule []*domain.Installment) error
	FindLoanByID(id string) (*domain.Loan, error)
	FindDueInstallments() ([]*domain.Installment, error)
	GetInstallmentSchedule(loanID string) ([]*domain.Installment, error)
	UpdateDelinquency(loanID string, isDelinquent bool) error
	SaveBilling(billing *domain.Billing, installmentID string) error
	FindUnpaidBillings(loanID string) ([]*domain.Billing, error)
	SavePaymentAndUpdateState(payment *domain.Payment, bill *domain.Billing) error
	UpdateBillingStatus(billing domain.Billing) error
}

// PostgresRepository implements the LoanRepository interface for PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// SaveLoan saves a loan and its entire schedule within a single transaction.
func (r *PostgresRepository) SaveLoan(loan *domain.Loan, schedule []*domain.Installment) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	loanQuery := `INSERT INTO loans (id, customer_id, principal_amount, interest_rate, term, term_type, term_payment, outstanding_amount, is_delinquent, status, creation_date)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err = tx.Exec(loanQuery, loan.ID, loan.CustomerID, loan.PrincipalAmount, loan.InterestRate, loan.Term, loan.TermType, loan.TermPayment, loan.OutstandingAmount, loan.IsDelinquent, loan.Status, loan.CreationDate)
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO installments (id, loan_id, term_number, amount_due, due_date, status) VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, inst := range schedule {
		_, err := stmt.Exec(inst.ID, inst.LoanID, inst.TermNumber, inst.AmountDue, inst.DueDate, inst.Status)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) FindLoanByID(id string) (*domain.Loan, error) {
	loan := &domain.Loan{}
	query := `SELECT id, customer_id, principal_amount, interest_rate, term, term_type, term_payment, outstanding_amount, is_delinquent, status, creation_date FROM loans WHERE id = $1`
	row := r.db.QueryRow(query, id)
	err := row.Scan(&loan.ID, &loan.CustomerID, &loan.PrincipalAmount, &loan.InterestRate, &loan.Term, &loan.TermType, &loan.TermPayment, &loan.OutstandingAmount, &loan.IsDelinquent, &loan.Status, &loan.CreationDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("loan not found")
		}
		return nil, err
	}
	return loan, nil
}

func (r *PostgresRepository) FindDueInstallments() ([]*domain.Installment, error) {
	query := `SELECT id, loan_id, term_number, amount_due, due_date, billed_date, status FROM installments WHERE status = 'PENDING' AND billed_date = $1`
	rows, err := r.db.Query(query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var due []*domain.Installment
	for rows.Next() {
		inst := &domain.Installment{}
		err := rows.Scan(&inst.ID, &inst.LoanID, &inst.TermNumber, &inst.AmountDue, &inst.DueDate, &inst.BilledDate, &inst.Status)
		if err != nil {
			return nil, err
		}
		due = append(due, inst)
	}
	return due, nil
}

func (r *PostgresRepository) GetInstallmentSchedule(loanID string) ([]*domain.Installment, error) {
	query := `SELECT id, loan_id, term_number, amount_due, due_date, status FROM installments WHERE loan_id = $1 ORDER BY term_number ASC`
	rows, err := r.db.Query(query, loanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedule []*domain.Installment
	for rows.Next() {
		inst := &domain.Installment{}
		err := rows.Scan(&inst.ID, &inst.LoanID, &inst.TermNumber, &inst.AmountDue, &inst.DueDate, &inst.Status)
		if err != nil {
			return nil, err
		}
		schedule = append(schedule, inst)
	}
	return schedule, nil
}

func (r *PostgresRepository) UpdateDelinquency(loanID string, isDelinquent bool) error {
	query := `UPDATE loans SET is_delinquent = $1 WHERE id = $2`
	_, err := r.db.Exec(query, isDelinquent, loanID)
	return err
}

func (r *PostgresRepository) SaveBilling(billing *domain.Billing, installmentID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	billingQuery := `INSERT INTO billings (id, installment_id, amount, billing_date, status) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(billingQuery, billing.ID, billing.InstallmentID, billing.Amount, billing.BillingDate, billing.Status)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) UpdateBillingStatus(billing domain.Billing) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	billingQuery := `UPDATE billings SET (billing.ID, billing.status) VALUES ($1, $2)`
	_, err = tx.Exec(billingID, billing.Status)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) FindUnpaidBillings(loanID string) ([]*domain.Billing, error) {
	query := `
		SELECT b.id, b.installment_id, b.amount, b.billing_date, b.status
		FROM billings b
		JOIN installments i ON b.installment_id = i.id
		WHERE i.loan_id = $1 AND b.status != 'PAID'
		ORDER BY i.term_number ASC`

	rows, err := r.db.Query(query, loanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var unpaid []*domain.Billing
	for rows.Next() {
		b := &domain.Billing{}
		err := rows.Scan(&b.ID, &b.InstallmentID, &b.Amount, &b.BillingDate, &b.Status)
		if err != nil {
			return nil, err
		}
		unpaid = append(unpaid, b)
	}
	return unpaid, nil
}

func (r *PostgresRepository) SavePaymentAndUpdateState(payment *domain.Payment, bill *domain.Billing) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// 1. Save Payment
	_, err = tx.Exec(`INSERT INTO payments (id, billing_id, amount_paid, payment_date) VALUES ($1, $2, $3, $4)`,
		payment.ID, payment.BillingID, payment.AmountPaid, payment.PaymentDate)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Update Billing status
	_, err = tx.Exec(`UPDATE billings SET status = 'PAID' WHERE id = $1`, bill.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 3. Update Installment status
	res, err := tx.Exec(`UPDATE installments SET status = 'PAID' WHERE id = $1`, bill.InstallmentID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		tx.Rollback()
		return fmt.Errorf("installment not found for billing update")
	}
	
	// 4. Update Loan outstanding amount
	_, err = tx.Exec(`UPDATE loans SET outstanding_amount = outstanding_amount - $1 WHERE id = (SELECT loan_id FROM installments WHERE id = $2)`,
		bill.Amount, bill.InstallmentID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ManuallyAdvanceTime is a helper for simulation purposes only.
func (r *PostgresRepository) ManuallyAdvanceTime(loanID string, weeks int) {
	query := `UPDATE installments SET due_date = $1 WHERE loan_id = $2 AND term_number <= $3`
	r.db.Exec(query, time.Now().AddDate(0, 0, -1), loanID, weeks)
}