package services

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/models"
)

// SalesStateMachine manages valid state transitions for sales
type SalesStateMachine struct {
	validTransitions map[string][]string
	transitionRules  map[string]TransitionRule
}

// TransitionRule defines conditions for state transitions
type TransitionRule struct {
	RequiresPayment   bool
	RequiresInventory bool
	RequiresApproval  bool
	AllowsReversal    bool
	Description       string
}

// StateTransitionResult contains the result of a state transition attempt
type StateTransitionResult struct {
	Success     bool   `json:"success"`
	NewStatus   string `json:"new_status"`
	PrevStatus  string `json:"previous_status"`
	Message     string `json:"message"`
	CanReverse  bool   `json:"can_reverse"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewSalesStateMachine creates a new state machine with predefined rules
func NewSalesStateMachine() *SalesStateMachine {
	return &SalesStateMachine{
		validTransitions: map[string][]string{
			models.SaleStatusDraft: {
				models.SaleStatusPending,   // For approval workflow
				models.SaleStatusConfirmed, // Direct confirmation
				models.SaleStatusCancelled, // Cancel draft
			},
			models.SaleStatusPending: {
				models.SaleStatusConfirmed, // Approved
				models.SaleStatusDraft,     // Rejected back to draft
				models.SaleStatusCancelled, // Cancelled during approval
			},
			models.SaleStatusConfirmed: {
				models.SaleStatusInvoiced,  // Create invoice
				models.SaleStatusCompleted, // Direct completion (cash sales)
				models.SaleStatusCancelled, // Cancel confirmed sale
			},
			models.SaleStatusInvoiced: {
				models.SaleStatusPaid,      // Full payment received
				models.SaleStatusOverdue,   // Past due date
				models.SaleStatusCompleted, // Mark as completed
				models.SaleStatusCancelled, // Cancel invoiced sale
			},
			models.SaleStatusOverdue: {
				models.SaleStatusPaid,      // Payment received
				models.SaleStatusCompleted, // Manual completion
				models.SaleStatusCancelled, // Write off
			},
			models.SaleStatusPaid: {
				models.SaleStatusCompleted, // Final completion
			},
			models.SaleStatusCompleted: {
				// Terminal state - only reversal allowed by admin
			},
			models.SaleStatusCancelled: {
				// Terminal state - only reversal allowed by admin
			},
		},
		transitionRules: map[string]TransitionRule{
			"DRAFT->PENDING": {
				RequiresPayment:   false,
				RequiresInventory: false,
				RequiresApproval:  true,
				AllowsReversal:    true,
				Description:       "Submit for approval",
			},
			"DRAFT->CONFIRMED": {
				RequiresPayment:   false,
				RequiresInventory: true,
				RequiresApproval:  false,
				AllowsReversal:    true,
				Description:       "Direct confirmation without approval",
			},
			"PENDING->CONFIRMED": {
				RequiresPayment:   false,
				RequiresInventory: true,
				RequiresApproval:  false,
				AllowsReversal:    true,
				Description:       "Approval granted",
			},
			"CONFIRMED->INVOICED": {
				RequiresPayment:   false,
				RequiresInventory: false,
				RequiresApproval:  false,
				AllowsReversal:    true,
				Description:       "Generate invoice",
			},
			"INVOICED->PAID": {
				RequiresPayment:   true,
				RequiresInventory: false,
				RequiresApproval:  false,
				AllowsReversal:    true,
				Description:       "Payment received",
			},
			"INVOICED->OVERDUE": {
				RequiresPayment:   false,
				RequiresInventory: false,
				RequiresApproval:  false,
				AllowsReversal:    true,
				Description:       "Past due date",
			},
			"OVERDUE->PAID": {
				RequiresPayment:   true,
				RequiresInventory: false,
				RequiresApproval:  false,
				AllowsReversal:    true,
				Description:       "Late payment received",
			},
			"PAID->COMPLETED": {
				RequiresPayment:   false,
				RequiresInventory: false,
				RequiresApproval:  false,
				AllowsReversal:    false,
				Description:       "Final completion",
			},
		},
	}
}

// ValidateTransition checks if a state transition is valid
func (sm *SalesStateMachine) ValidateTransition(fromStatus, toStatus string) error {
	log.Printf("🔍 Validating transition from %s to %s", fromStatus, toStatus)
	
	// Check if from status exists in transitions map
	validNextStates, exists := sm.validTransitions[fromStatus]
	if !exists {
		return fmt.Errorf("unknown status: %s", fromStatus)
	}
	
	// Check if to status is in the list of valid next states
	for _, validState := range validNextStates {
		if validState == toStatus {
			log.Printf("✅ Transition from %s to %s is valid", fromStatus, toStatus)
			return nil
		}
	}
	
	return fmt.Errorf("invalid status transition from %s to %s. Valid transitions: %v", 
		fromStatus, toStatus, validNextStates)
}

// GetValidTransitions returns all valid transitions from a given status
func (sm *SalesStateMachine) GetValidTransitions(fromStatus string) []string {
	if transitions, exists := sm.validTransitions[fromStatus]; exists {
		return transitions
	}
	return []string{}
}

// ExecuteTransition attempts to execute a state transition with validation
func (sm *SalesStateMachine) ExecuteTransition(sale *models.Sale, toStatus string, context TransitionContext) (*StateTransitionResult, error) {
	log.Printf("🔄 Executing transition for sale %d from %s to %s", sale.ID, sale.Status, toStatus)
	
	result := &StateTransitionResult{
		PrevStatus: sale.Status,
		NewStatus:  toStatus,
		Timestamp:  time.Now(),
	}
	
	// Validate transition is allowed
	if err := sm.ValidateTransition(sale.Status, toStatus); err != nil {
		result.Success = false
		result.Message = err.Error()
		return result, err
	}
	
	// Get transition rule
	ruleKey := fmt.Sprintf("%s->%s", sale.Status, toStatus)
	rule, exists := sm.transitionRules[ruleKey]
	if exists {
		result.CanReverse = rule.AllowsReversal
		result.Message = rule.Description
		
		// Check rule conditions
		if err := sm.validateTransitionConditions(sale, rule, context); err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Transition condition failed: %s", err.Error())
			return result, err
		}
	}
	
	// Execute business logic for transition
	if err := sm.executeTransitionLogic(sale, toStatus, context); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Transition execution failed: %s", err.Error())
		return result, err
	}
	
	// Update sale status
	prevStatus := sale.Status
	sale.Status = toStatus
	
	result.Success = true
	result.Message = fmt.Sprintf("Successfully transitioned from %s to %s", prevStatus, toStatus)
	
	log.Printf("✅ Sale %d status transition completed: %s -> %s", sale.ID, prevStatus, toStatus)
	return result, nil
}

// TransitionContext provides context data for state transitions
type TransitionContext struct {
	UserID          uint                   `json:"user_id"`
	Reason          string                 `json:"reason"`
	PaymentReceived float64                `json:"payment_received"`
	InventoryCheck  bool                   `json:"inventory_check"`
	ApprovalLevel   string                 `json:"approval_level"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// validateTransitionConditions checks if all conditions for a transition are met
func (sm *SalesStateMachine) validateTransitionConditions(sale *models.Sale, rule TransitionRule, context TransitionContext) error {
	// Check payment requirement
	if rule.RequiresPayment && context.PaymentReceived <= 0 {
		return fmt.Errorf("payment is required for this transition")
	}
	
	// Check inventory requirement
	if rule.RequiresInventory && !context.InventoryCheck {
		return fmt.Errorf("inventory validation is required for this transition")
	}
	
	// Check approval requirement
	if rule.RequiresApproval && context.ApprovalLevel == "" {
		return fmt.Errorf("approval is required for this transition")
	}
	
	// Validate payment amount for payment-related transitions
	if rule.RequiresPayment {
		if context.PaymentReceived > sale.OutstandingAmount {
			return fmt.Errorf("payment amount %.2f exceeds outstanding amount %.2f", 
				context.PaymentReceived, sale.OutstandingAmount)
		}
	}
	
	return nil
}

// executeTransitionLogic performs business logic specific to each transition
func (sm *SalesStateMachine) executeTransitionLogic(sale *models.Sale, toStatus string, context TransitionContext) error {
	switch toStatus {
	case models.SaleStatusConfirmed:
		return sm.handleConfirmTransition(sale, context)
	case models.SaleStatusInvoiced:
		return sm.handleInvoiceTransition(sale, context)
	case models.SaleStatusPaid:
		return sm.handlePaymentTransition(sale, context)
	case models.SaleStatusOverdue:
		return sm.handleOverdueTransition(sale, context)
	case models.SaleStatusCancelled:
		return sm.handleCancelTransition(sale, context)
	case models.SaleStatusCompleted:
		return sm.handleCompletionTransition(sale, context)
	}
	
	return nil // No specific logic needed
}

// handleConfirmTransition handles business logic for confirmation
func (sm *SalesStateMachine) handleConfirmTransition(sale *models.Sale, context TransitionContext) error {
	log.Printf("📋 Handling confirmation logic for sale %d", sale.ID)
	
	// Validate inventory is available
	if context.InventoryCheck {
		// This would integrate with inventory service
		log.Printf("✅ Inventory check passed for sale %d", sale.ID)
	}
	
	// Set initial outstanding amount if not set
	if sale.OutstandingAmount == 0 {
		sale.OutstandingAmount = sale.TotalAmount
	}
	
	return nil
}

// handleInvoiceTransition handles business logic for invoicing
func (sm *SalesStateMachine) handleInvoiceTransition(sale *models.Sale, context TransitionContext) error {
	log.Printf("🧾 Handling invoice generation for sale %d", sale.ID)
	
	// Generate invoice number if not exists
	if sale.InvoiceNumber == "" {
		sale.InvoiceNumber = fmt.Sprintf("INV-%d-%d", time.Now().Year(), sale.ID)
		log.Printf("📄 Generated invoice number: %s", sale.InvoiceNumber)
	}
	
	return nil
}

// handlePaymentTransition handles business logic for payment
func (sm *SalesStateMachine) handlePaymentTransition(sale *models.Sale, context TransitionContext) error {
	log.Printf("💰 Handling payment logic for sale %d", sale.ID)
	
	// Update payment amounts
	sale.PaidAmount += context.PaymentReceived
	sale.OutstandingAmount -= context.PaymentReceived
	
	// Ensure amounts are consistent
	if sale.OutstandingAmount < 0 {
		sale.OutstandingAmount = 0
	}
	
	log.Printf("💳 Updated amounts for sale %d: paid=%.2f, outstanding=%.2f", 
		sale.ID, sale.PaidAmount, sale.OutstandingAmount)
	
	return nil
}

// handleOverdueTransition handles business logic for overdue status
func (sm *SalesStateMachine) handleOverdueTransition(sale *models.Sale, context TransitionContext) error {
	log.Printf("⏰ Handling overdue logic for sale %d", sale.ID)
	
	// Check if actually overdue
	if !time.Now().After(sale.DueDate) {
		return fmt.Errorf("sale is not yet overdue (due date: %s)", sale.DueDate.Format("2006-01-02"))
	}
	
	// Could add overdue fees here
	log.Printf("📅 Sale %d is overdue (due: %s)", sale.ID, sale.DueDate.Format("2006-01-02"))
	
	return nil
}

// handleCancelTransition handles business logic for cancellation
func (sm *SalesStateMachine) handleCancelTransition(sale *models.Sale, context TransitionContext) error {
	log.Printf("❌ Handling cancellation logic for sale %d", sale.ID)
	
	// Require reason for cancellation
	if context.Reason == "" {
		return fmt.Errorf("cancellation reason is required")
	}
	
	// Reset amounts for cancelled sales
	sale.OutstandingAmount = 0
	
	log.Printf("🚫 Sale %d cancelled with reason: %s", sale.ID, context.Reason)
	
	return nil
}

// handleCompletionTransition handles business logic for completion
func (sm *SalesStateMachine) handleCompletionTransition(sale *models.Sale, context TransitionContext) error {
	log.Printf("🎯 Handling completion logic for sale %d", sale.ID)
	
	// Ensure all amounts are properly set
	if sale.OutstandingAmount > 0.01 { // Allow small tolerance
		return fmt.Errorf("cannot complete sale with outstanding amount %.2f", sale.OutstandingAmount)
	}
	
	sale.OutstandingAmount = 0 // Ensure it's exactly zero
	
	log.Printf("🏁 Sale %d marked as completed", sale.ID)
	
	return nil
}

// IsTerminalState checks if a status is terminal (no further transitions)
func (sm *SalesStateMachine) IsTerminalState(status string) bool {
	transitions := sm.GetValidTransitions(status)
	return len(transitions) == 0
}

// CanTransitionTo checks if a specific transition is possible
func (sm *SalesStateMachine) CanTransitionTo(fromStatus, toStatus string) bool {
	validTransitions := sm.GetValidTransitions(fromStatus)
	for _, valid := range validTransitions {
		if valid == toStatus {
			return true
		}
	}
	return false
}