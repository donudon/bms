package handler

import (
	"bms/internal/domain"
	"bms/internal/usecase"
)

type LoanHandler struct {
	billingUsecase *usecase.BillingUsecase
}

func NewLoanHandler(billingUsecase *usecase.BillingUsecase) *LoanHandler {
	return &LoanHandler{billingUsecase: billingUsecase}
}

func (h *LoanHandler) CreateLoan(customerID string, principal, rate float64, term int, termType domain.TermType) (*domain.Loan, error) {
	return h.billingUsecase.CreateLoan(customerID, principal, rate, term, termType)
}

func (h *LoanHandler) CreateBilling() (int, error) {
	return h.billingUsecase.CreateBilling()
}

func (h *LoanHandler) MakePayment(loanID string, paymentAmount float64) (string, error) {
	return h.billingUsecase.MakePayment(loanID, paymentAmount)
}

func (h *LoanHandler) GetLoanInfo(loanID string) (*domain.Loan, error) {
	return h.billingUsecase.GetLoanInfo(loanID)
}
