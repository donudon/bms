package usecase

import (
	"fmt"
	"sort"
	"time"

	"bms/internal/domain"
	"bms/internal/repository"
)

type LoanUsecase struct {
	loanRepo repository.LoanRepository
}

func NewLoanUsecase(loanRepo repository.LoanRepository) *LoanUsecase {
	return &LoanUsecase{loanRepo: loanRepo}
}

func (uc *LoanUsecase) CreateLoan(customerID string, principal, rate float64, term int, termType domain.TermType) (*domain.Loan, error) {
	creationDate := time.Now()
	totalRepayable := principal * (1 + rate)
	termPayment := totalRepayable / float64(term)

	loan := &domain.Loan{
		CustomerID:        customerID,
		PrincipalAmount:   principal,
		InterestRate:      rate,
		Term:              term,
		TermType:          termType,
		TermPayment:       termPayment,
		OutstandingAmount: totalRepayable,
		IsDelinquent:      false,
		Status:            domain.Active,
		CreationDate:      creationDate,
	}

	schedule := make([]*domain.Installment, term)
	lastDueDate := creationDate
	for i := 1; i <= term; i++ {
		lastDueDate = lastDueDate.AddDate(0, 0, 7)
		billedDate := lastDueDate.AddDate(0, 0, -5)
		inst := &domain.Installment{
			LoanID:      loanID,
			TermNumber:  i,
			AmountDue:   termPayment,
			DueDate:     lastDueDate,
			BilledDate:  billedDate,
			Status:      domain.Pending,
		}
		schedule[i-1] = inst
	}

	return loan, uc.loanRepo.SaveLoan(loan, schedule)
}

func (uc *LoanUsecase) CreateBilling() (int, error) {
	pendingInstallments, err := uc.loanRepo.FindDueInstallments()
	if err != nil {
		return 0, err
	}

	billsCreated := 0
	now := time.Now()

	for _, inst := range pendingInstallments {
		// Check for Delinquency
		err := uc.updateDelinquencyStatus(inst.LoanID, inst.TermNumber)
		if err != nil {
			// Log the error but continue processing other bills
			fmt.Printf("Warning: could not update delinquency for loan %s: %v\n", inst.LoanID, err)
		}

		// Create Billing Record
		billing := &domain.Billing{
			ID:            uuid.New().String(),
			InstallmentID: inst.ID,
			Amount:        inst.AmountDue,
			BillingDate:   now,
			Status:        domain.PendingBilling,
		}
		
		err = uc.loanRepo.SaveBilling(billing, inst.ID)
		if err != nil {
			return billsCreated, err
		}

		if err := uc.notificationService.SendBillingNotification(loan, billing); err != nil {
			fmt.Printf("Error sending notification for billing %s: %v\n", billing.ID, err)
			continue
		}

		err = uc.loanRepo.UpdateBillingStatus(billing)
		if err != nil {
			fmt.Printf("Error updating status for billing %s: %v\n", billing.ID, err)
			continue
		}
		billsCreated++
	}
	return billsCreated, nil
}

func (uc *LoanUsecase) updateDelinquencyStatus(loanID string, currentTerm int) error {
	if currentTerm < 2 {
		return nil // Cannot be delinquent before the 2nd term
	}
	
	schedule, err := uc.loanRepo.GetInstallmentSchedule(loanID)
	if err != nil {
		return err
	}
	
	now := time.Now()
	
	// Sort to be safe
	sort.Slice(schedule, func(i, j int) bool {
		return schedule[i].TermNumber < schedule[j].TermNumber
	})

	prevInst1 := schedule[currentTerm-2] // n-1
	if prevInst1.Status != domain.Paid && prevInst1.DueDate.Before(now) {
		if currentTerm > 2 {
			prevInst2 := schedule[currentTerm-3] // n-2
			if prevInst2.Status != domain.Paid && prevInst2.DueDate.Before(now) {
				return uc.loanRepo.UpdateDelinquency(loanID, true)
			}
		}
	}
	return nil
}


func (uc *LoanUsecase) MakePayment(loanID string, paymentAmount float64) (string, error) {
	unpaidBillings, err := uc.loanRepo.FindUnpaidBillings(loanID)
	if err != nil {
		return "", err
	}
	if len(unpaidBillings) == 0 {
		return "No outstanding bills to pay.", nil
	}

	billsPaidCount := 0
	for _, bill := range unpaidBillings {
		if paymentAmount >= bill.Amount {
			paymentAmount -= bill.Amount

			payment := &domain.Payment{
				ID:          uuid.New().String(),
				BillingID:   bill.ID,
				AmountPaid:  bill.Amount,
				PaymentDate: time.Now(),
			}

			err := uc.loanRepo.SavePaymentAndUpdateState(payment, bill)
			if err != nil {
				return "", err // Stop on first error
			}
			billsPaidCount++
		} else {
			break
		}
	}
	
	if billsPaidCount == 0 {
		return "Payment amount insufficient.", nil
	}

	loan, _ := uc.loanRepo.FindLoanByID(loanID)
	return fmt.Sprintf("Payment successful. Paid %d bill(s). New outstanding: %.2f", billsPaidCount, loan.OutstandingAmount), nil
}

func (uc *LoanUsecase) GetLoanInfo(loanID string) (*domain.Loan, error) {
	return uc.loanRepo.FindLoanByID(loanID)
}
