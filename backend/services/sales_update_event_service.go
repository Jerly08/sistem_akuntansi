package services

import (
	"fmt"
	"log"
	"sync"
	"time"
	"encoding/json"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

// Event types
const (
	EventPaymentCreated    = "payment.created"
	EventPaymentUpdated    = "payment.updated"
	EventPaymentDeleted    = "payment.deleted"
	EventAllocationCreated = "allocation.created"
	EventAllocationUpdated = "allocation.updated"
	EventAllocationDeleted = "allocation.deleted"
	EventSaleStatusChanged = "sale.status.changed"
	EventPurchaseStatusChanged = "purchase.status.changed"
)

// Event data structures
type PaymentEvent struct {
	Type      string          `json:"type"`
	PaymentID uint            `json:"payment_id"`
	Payment   models.Payment  `json:"payment"`
	Timestamp time.Time       `json:"timestamp"`
	UserID    uint            `json:"user_id"`
	Changes   json.RawMessage `json:"changes,omitempty"`
}

type AllocationEvent struct {
	Type       string                   `json:"type"`
	Allocation models.PaymentAllocation `json:"allocation"`
	Timestamp  time.Time                `json:"timestamp"`
	UserID     uint                     `json:"user_id"`
	Changes    json.RawMessage          `json:"changes,omitempty"`
}

type SalesStatusEvent struct {
	Type         string      `json:"type"`
	SaleID       uint        `json:"sale_id"`
	PurchaseID   uint        `json:"purchase_id,omitempty"`
	OldStatus    string      `json:"old_status"`
	NewStatus    string      `json:"new_status"`
	OldOutstanding float64   `json:"old_outstanding"`
	NewOutstanding float64   `json:"new_outstanding"`
	Timestamp    time.Time   `json:"timestamp"`
	UserID       uint        `json:"user_id"`
}

// Event handler interface
type EventHandler interface {
	Handle(event interface{}) error
}

// Sales Update Event Service
type SalesUpdateEventService struct {
	db               *gorm.DB
	salesRepo        *repositories.SalesRepository
	purchaseRepo     *repositories.PurchaseRepository
	paymentRepo      *repositories.PaymentRepository
	eventQueue       chan interface{}
	handlers         map[string][]EventHandler
	workerPool       int
	isRunning        bool
	stopChan         chan bool
	wg               sync.WaitGroup
	mu               sync.RWMutex
}

func NewSalesUpdateEventService(
	db *gorm.DB,
	salesRepo *repositories.SalesRepository,
	purchaseRepo *repositories.PurchaseRepository,
	paymentRepo *repositories.PaymentRepository,
	workerPoolSize int,
) *SalesUpdateEventService {
	if workerPoolSize <= 0 {
		workerPoolSize = 5 // Default worker pool size
	}

	service := &SalesUpdateEventService{
		db:           db,
		salesRepo:    salesRepo,
		purchaseRepo: purchaseRepo,
		paymentRepo:  paymentRepo,
		eventQueue:   make(chan interface{}, 1000), // Buffered channel
		handlers:     make(map[string][]EventHandler),
		workerPool:   workerPoolSize,
		stopChan:     make(chan bool),
	}

	// Register default handlers
	service.registerDefaultHandlers()

	return service
}

// 🚀 Start the event processing service
func (s *SalesUpdateEventService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return
	}

	s.isRunning = true
	log.Printf("Starting Sales Update Event Service with %d workers", s.workerPool)

	// Start worker goroutines
	for i := 0; i < s.workerPool; i++ {
		s.wg.Add(1)
		go s.eventWorker(i)
	}
}

// 🛑 Stop the event processing service
func (s *SalesUpdateEventService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	log.Println("Stopping Sales Update Event Service...")
	s.isRunning = false

	// Signal all workers to stop
	close(s.stopChan)
	
	// Wait for all workers to finish
	s.wg.Wait()

	// Close the event queue
	close(s.eventQueue)
	
	log.Println("Sales Update Event Service stopped")
}

// 👷 Event worker goroutine
func (s *SalesUpdateEventService) eventWorker(workerID int) {
	defer s.wg.Done()
	
	log.Printf("Event worker %d started", workerID)

	for {
		select {
		case event, ok := <-s.eventQueue:
			if !ok {
				log.Printf("Event worker %d: Queue closed, stopping", workerID)
				return
			}
			
			if err := s.processEvent(event); err != nil {
				log.Printf("Event worker %d: Error processing event: %v", workerID, err)
			}

		case <-s.stopChan:
			log.Printf("Event worker %d: Received stop signal", workerID)
			return
		}
	}
}

// 📤 Emit an event
func (s *SalesUpdateEventService) EmitEvent(event interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isRunning {
		return fmt.Errorf("event service is not running")
	}

	select {
	case s.eventQueue <- event:
		return nil
	default:
		return fmt.Errorf("event queue is full")
	}
}

// 🔄 Process a single event
func (s *SalesUpdateEventService) processEvent(event interface{}) error {
	var eventType string

	switch e := event.(type) {
	case PaymentEvent:
		eventType = e.Type
	case AllocationEvent:
		eventType = e.Type
	case SalesStatusEvent:
		eventType = e.Type
	default:
		return fmt.Errorf("unknown event type: %T", event)
	}

	handlers, exists := s.handlers[eventType]
	if !exists {
		log.Printf("No handlers registered for event type: %s", eventType)
		return nil
	}

	var lastError error
	for _, handler := range handlers {
		if err := handler.Handle(event); err != nil {
			log.Printf("Handler error for event %s: %v", eventType, err)
			lastError = err
		}
	}

	return lastError
}

// 📋 Register an event handler
func (s *SalesUpdateEventService) RegisterHandler(eventType string, handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handlers[eventType] = append(s.handlers[eventType], handler)
}

// 🔧 Register default event handlers
func (s *SalesUpdateEventService) registerDefaultHandlers() {
	// Payment created - update outstanding amounts
	s.RegisterHandler(EventPaymentCreated, &PaymentCreatedHandler{
		salesRepo:    s.salesRepo,
		purchaseRepo: s.purchaseRepo,
		eventService: s,
	})

	// Allocation created - update specific invoice/bill
	s.RegisterHandler(EventAllocationCreated, &AllocationCreatedHandler{
		salesRepo:    s.salesRepo,
		purchaseRepo: s.purchaseRepo,
		eventService: s,
	})

	// Sales status changed - trigger downstream updates
	s.RegisterHandler(EventSaleStatusChanged, &SalesStatusChangedHandler{
		db: s.db,
	})

	// Purchase status changed - trigger downstream updates  
	s.RegisterHandler(EventPurchaseStatusChanged, &PurchaseStatusChangedHandler{
		db: s.db,
	})
}

// 🎯 PAYMENT CREATED HANDLER
type PaymentCreatedHandler struct {
	salesRepo    *repositories.SalesRepository
	purchaseRepo *repositories.PurchaseRepository
	eventService *SalesUpdateEventService
}

func (h *PaymentCreatedHandler) Handle(event interface{}) error {
	paymentEvent, ok := event.(PaymentEvent)
	if !ok {
		return fmt.Errorf("invalid event type for PaymentCreatedHandler")
	}

	log.Printf("Processing payment created: %s (ID: %d)", paymentEvent.Payment.Code, paymentEvent.PaymentID)

	// Get all allocations for this payment
	var allocations []models.PaymentAllocation
	if err := h.eventService.db.Where("payment_id = ?", paymentEvent.PaymentID).Find(&allocations).Error; err != nil {
		return fmt.Errorf("failed to get payment allocations: %v", err)
	}

	// Process each allocation
	for _, allocation := range allocations {
		allocationEvent := AllocationEvent{
			Type:       EventAllocationCreated,
			Allocation: allocation,
			Timestamp:  time.Now(),
			UserID:     paymentEvent.UserID,
		}

		if err := h.eventService.EmitEvent(allocationEvent); err != nil {
			log.Printf("Failed to emit allocation event: %v", err)
		}
	}

	return nil
}

// 🎯 ALLOCATION CREATED HANDLER  
type AllocationCreatedHandler struct {
	salesRepo    *repositories.SalesRepository
	purchaseRepo *repositories.PurchaseRepository
	eventService *SalesUpdateEventService
}

func (h *AllocationCreatedHandler) Handle(event interface{}) error {
	allocationEvent, ok := event.(AllocationEvent)
	if !ok {
		return fmt.Errorf("invalid event type for AllocationCreatedHandler")
	}

	allocation := allocationEvent.Allocation
	
	log.Printf("Processing allocation created: Payment %d -> Amount %.2f", 
		allocation.PaymentID, allocation.AllocatedAmount)

	// Update invoice outstanding if allocated to invoice
	if allocation.InvoiceID != nil {
		if err := h.updateInvoiceOutstanding(*allocation.InvoiceID, allocationEvent.UserID); err != nil {
			return fmt.Errorf("failed to update invoice outstanding: %v", err)
		}
	}

	// Update bill outstanding if allocated to bill
	if allocation.BillID != nil {
		if err := h.updateBillOutstanding(*allocation.BillID, allocationEvent.UserID); err != nil {
			return fmt.Errorf("failed to update bill outstanding: %v", err)
		}
	}

	return nil
}

func (h *AllocationCreatedHandler) updateInvoiceOutstanding(invoiceID uint, userID uint) error {
	// Start transaction
	tx := h.eventService.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get invoice with lock
	var sale models.Sale
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&sale, invoiceID).Error; err != nil {
		tx.Rollback()
		return err
	}

	oldStatus := sale.Status
	oldOutstanding := sale.OutstandingAmount

	// Calculate new outstanding amount
	var totalAllocated float64
	tx.Model(&models.PaymentAllocation{}).
		Where("invoice_id = ?", invoiceID).
		Select("COALESCE(SUM(allocated_amount), 0)").
		Scan(&totalAllocated)

	newOutstanding := sale.TotalAmount - totalAllocated
	if newOutstanding < 0 {
		newOutstanding = 0
	}

	// Determine new status
	newStatus := models.SaleStatusPending
	if newOutstanding <= 0.01 {
		newStatus = models.SaleStatusPaid
		newOutstanding = 0
	} else if totalAllocated > 0 {
		newStatus = models.SaleStatusCompleted // Partial payment
	}

	// Update sale
	if err := tx.Model(&sale).Updates(map[string]interface{}{
		"outstanding_amount": newOutstanding,
		"status":            newStatus,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Emit status change event if status changed
	if oldStatus != newStatus || oldOutstanding != newOutstanding {
		statusEvent := SalesStatusEvent{
			Type:           EventSaleStatusChanged,
			SaleID:         invoiceID,
			OldStatus:      oldStatus,
			NewStatus:      newStatus,
			OldOutstanding: oldOutstanding,
			NewOutstanding: newOutstanding,
			Timestamp:      time.Now(),
			UserID:         userID,
		}

		h.eventService.EmitEvent(statusEvent)
	}

	log.Printf("Invoice %d outstanding updated: %.2f -> %.2f (Status: %s -> %s)",
		invoiceID, oldOutstanding, newOutstanding, oldStatus, newStatus)

	return nil
}

func (h *AllocationCreatedHandler) updateBillOutstanding(billID uint, userID uint) error {
	// Start transaction
	tx := h.eventService.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get purchase with lock
	var purchase models.Purchase
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&purchase, billID).Error; err != nil {
		tx.Rollback()
		return err
	}

	oldStatus := purchase.Status
	oldOutstanding := purchase.OutstandingAmount

	// Calculate new outstanding amount
	var totalAllocated float64
	tx.Model(&models.PaymentAllocation{}).
		Where("bill_id = ?", billID).
		Select("COALESCE(SUM(allocated_amount), 0)").
		Scan(&totalAllocated)

	newOutstanding := purchase.TotalAmount - totalAllocated
	if newOutstanding < 0 {
		newOutstanding = 0
	}

	// Determine new status
	newStatus := models.PurchaseStatusPending
	if newOutstanding <= 0.01 {
		newStatus = models.PurchaseStatusPaid
		newOutstanding = 0
	} else if totalAllocated > 0 {
		newStatus = models.PurchaseStatusCompleted // Partial payment
	}

	// Update purchase
	if err := tx.Model(&purchase).Updates(map[string]interface{}{
		"outstanding_amount": newOutstanding,
		"status":            newStatus,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Emit status change event if status changed
	if oldStatus != newStatus || oldOutstanding != newOutstanding {
		statusEvent := SalesStatusEvent{
			Type:           EventPurchaseStatusChanged,
			PurchaseID:     billID,
			OldStatus:      oldStatus,
			NewStatus:      newStatus,
			OldOutstanding: oldOutstanding,
			NewOutstanding: newOutstanding,
			Timestamp:      time.Now(),
			UserID:         userID,
		}

		h.eventService.EmitEvent(statusEvent)
	}

	log.Printf("Bill %d outstanding updated: %.2f -> %.2f (Status: %s -> %s)",
		billID, oldOutstanding, newOutstanding, oldStatus, newStatus)

	return nil
}

// 🎯 SALES STATUS CHANGED HANDLER
type SalesStatusChangedHandler struct {
	db *gorm.DB
}

func (h *SalesStatusChangedHandler) Handle(event interface{}) error {
	statusEvent, ok := event.(SalesStatusEvent)
	if !ok {
		return fmt.Errorf("invalid event type for SalesStatusChangedHandler")
	}

	log.Printf("Sales status changed: Sale %d from %s to %s (Outstanding: %.2f -> %.2f)",
		statusEvent.SaleID, statusEvent.OldStatus, statusEvent.NewStatus,
		statusEvent.OldOutstanding, statusEvent.NewOutstanding)

	// Here you can add additional logic like:
	// - Send notifications
	// - Update related reports
	// - Trigger integrations
	// - Log audit trail

	return nil
}

// 🎯 PURCHASE STATUS CHANGED HANDLER
type PurchaseStatusChangedHandler struct {
	db *gorm.DB
}

func (h *PurchaseStatusChangedHandler) Handle(event interface{}) error {
	statusEvent, ok := event.(SalesStatusEvent)
	if !ok {
		return fmt.Errorf("invalid event type for PurchaseStatusChangedHandler")
	}

	log.Printf("Purchase status changed: Purchase %d from %s to %s (Outstanding: %.2f -> %.2f)",
		statusEvent.PurchaseID, statusEvent.OldStatus, statusEvent.NewStatus,
		statusEvent.OldOutstanding, statusEvent.NewOutstanding)

	// Here you can add additional logic like:
	// - Send notifications
	// - Update related reports
	// - Trigger integrations
	// - Log audit trail

	return nil
}

// 🛠️ HELPER METHODS

// Emit payment created event
func (s *SalesUpdateEventService) EmitPaymentCreated(payment models.Payment, userID uint) error {
	event := PaymentEvent{
		Type:      EventPaymentCreated,
		PaymentID: payment.ID,
		Payment:   payment,
		Timestamp: time.Now(),
		UserID:    userID,
	}
	return s.EmitEvent(event)
}

// Emit payment updated event
func (s *SalesUpdateEventService) EmitPaymentUpdated(payment models.Payment, userID uint, changes json.RawMessage) error {
	event := PaymentEvent{
		Type:      EventPaymentUpdated,
		PaymentID: payment.ID,
		Payment:   payment,
		Timestamp: time.Now(),
		UserID:    userID,
		Changes:   changes,
	}
	return s.EmitEvent(event)
}

// Emit allocation created event
func (s *SalesUpdateEventService) EmitAllocationCreated(allocation models.PaymentAllocation, userID uint) error {
	event := AllocationEvent{
		Type:       EventAllocationCreated,
		Allocation: allocation,
		Timestamp:  time.Now(),
		UserID:     userID,
	}
	return s.EmitEvent(event)
}

// Get event queue status
func (s *SalesUpdateEventService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"is_running":         s.isRunning,
		"worker_pool_size":   s.workerPool,
		"queue_length":       len(s.eventQueue),
		"queue_capacity":     cap(s.eventQueue),
		"registered_handlers": len(s.handlers),
	}
}

// Health check
func (s *SalesUpdateEventService) HealthCheck() error {
	if !s.isRunning {
		return fmt.Errorf("event service is not running")
	}

	queueUsage := float64(len(s.eventQueue)) / float64(cap(s.eventQueue))
	if queueUsage > 0.8 {
		return fmt.Errorf("event queue is %.1f%% full", queueUsage*100)
	}

	return nil
}