package service

import (
	"billing-engine/internal/domain"
	"fmt"
	"time"
)

// NotificationService defines the interface for sending notifications.
type NotificationService interface {
	SendBillingNotification(loan *domain.Loan, billing *domain.Billing) error
}

// mockNotificationService is a simple implementation for simulation.
type mockNotificationService struct{}

func NewMockNotificationService() NotificationService {
	return &mockNotificationService{}
}

// SendBillingNotification simulates sending an email.
func (s *mockNotificationService) SendBillingNotification(loan *domain.Loan, billing *domain.Billing) error {
	fmt.Printf("--- NOTIFICATION SENT ---\n")
	fmt.Printf("To: Customer %s\n", loan.CustomerID)
	fmt.Printf("Subject: Your bill #%s for Rp %.2f is ready\n", billing.ID[:8], billing.Amount)
	fmt.Printf("Sent at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("-------------------------\n")
	// In a real implementation, this would involve an API call to an email service.
	// We'll assume it always succeeds for this mock.
	return nil
}