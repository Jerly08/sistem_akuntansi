package services

import (
	"encoding/json"
	"fmt"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
)

type NotificationService struct {
	notificationRepo *repositories.NotificationRepository
}

func NewNotificationService(notificationRepo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(notification *models.Notification) error {
	return s.notificationRepo.Create(notification)
}

// GetUserNotifications gets notifications for a user
func (s *NotificationService) GetUserNotifications(userID uint, page, limit int, onlyUnread bool) ([]models.Notification, int64, error) {
	return s.notificationRepo.GetUserNotifications(userID, page, limit, onlyUnread)
}

// GetNotificationsByType gets notifications by type for a user
func (s *NotificationService) GetNotificationsByType(userID uint, notificationType string, page, limit int) ([]models.Notification, int64, error) {
	return s.notificationRepo.GetNotificationsByType(userID, notificationType, page, limit)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID, userID uint) error {
	return s.notificationRepo.MarkAsRead(notificationID, userID)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID uint) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}

// GetUnreadCount gets count of unread notifications
func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	return s.notificationRepo.GetUnreadCount(userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID, userID uint) error {
	return s.notificationRepo.Delete(notificationID, userID)
}

// CreateApprovalNotification creates approval-related notifications
func (s *NotificationService) CreateApprovalNotification(userID uint, notificationType, title, message string, data interface{}) error {
	// Convert data to JSON string
	var dataString string
	if data != nil {
		dataBytes, err := json.Marshal(data)
		if err == nil {
			dataString = string(dataBytes)
		}
	}

	notification := &models.Notification{
		UserID:   userID,
		Type:     notificationType,
		Title:    title,
		Message:  message,
		Data:     dataString,
		Priority: s.getNotificationPriority(notificationType),
	}

	return s.CreateNotification(notification)
}

// CreatePurchaseSubmissionNotification notifies when purchase is submitted for approval
func (s *NotificationService) CreatePurchaseSubmissionNotification(purchase *models.Purchase) error {
	// Get approvers based on purchase amount
	approvers := s.getApproversForPurchase(purchase)

	for _, approverID := range approvers {
		title := "Purchase Approval Required"
		message := fmt.Sprintf("Purchase %s requires your approval (Amount: Rp %,.2f)", 
			purchase.Code, purchase.TotalAmount)
		
		data := map[string]interface{}{
			"purchase_id":   purchase.ID,
			"purchase_code": purchase.Code,
			"vendor_name":   purchase.Vendor.Name,
			"total_amount":  purchase.TotalAmount,
			"action_type":   "approval_required",
		}

		err := s.CreateApprovalNotification(approverID, models.NotificationTypeApprovalPending, title, message, data)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreatePurchaseApprovedNotification notifies when purchase is approved
func (s *NotificationService) CreatePurchaseApprovedNotification(purchase *models.Purchase, approverID uint) error {
	title := "Purchase Approved"
	message := fmt.Sprintf("Your purchase request %s has been approved", purchase.Code)
	
	data := map[string]interface{}{
		"purchase_id":   purchase.ID,
		"purchase_code": purchase.Code,
		"approved_by":   approverID,
		"approved_at":   purchase.ApprovedAt,
		"action_type":   "approved",
	}

	return s.CreateApprovalNotification(purchase.UserID, models.NotificationTypeApprovalApproved, title, message, data)
}

// CreatePurchaseRejectedNotification notifies when purchase is rejected
func (s *NotificationService) CreatePurchaseRejectedNotification(purchase *models.Purchase, approverID uint, reason string) error {
	title := "Purchase Rejected"
	message := fmt.Sprintf("Your purchase request %s has been rejected", purchase.Code)
	if reason != "" {
		message += fmt.Sprintf(". Reason: %s", reason)
	}
	
	data := map[string]interface{}{
		"purchase_id":   purchase.ID,
		"purchase_code": purchase.Code,
		"rejected_by":   approverID,
		"rejected_at":   time.Now(),
		"reason":        reason,
		"action_type":   "rejected",
	}

	return s.CreateApprovalNotification(purchase.UserID, models.NotificationTypeApprovalRejected, title, message, data)
}

// SendBulkNotification sends notification to multiple users
func (s *NotificationService) SendBulkNotification(userIDs []uint, notificationType, title, message string, data interface{}) error {
	for _, userID := range userIDs {
		err := s.CreateApprovalNotification(userID, notificationType, title, message, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// Private helper methods

func (s *NotificationService) getNotificationPriority(notificationType string) string {
	switch notificationType {
	case models.NotificationTypeApprovalPending:
		return models.NotificationPriorityHigh
	case models.NotificationTypeApprovalRejected:
		return models.NotificationPriorityHigh
	case models.NotificationTypeApprovalApproved:
		return models.NotificationPriorityNormal
	default:
		return models.NotificationPriorityNormal
	}
}

func (s *NotificationService) getApproversForPurchase(purchase *models.Purchase) []uint {
	var approvers []uint
	
	// This is a simplified logic - in real implementation, 
	// you would query the approval workflow system
	
	// For demonstration purposes:
	// - Finance approves purchases up to 25M
	// - Director approves purchases above 25M
	
	if purchase.TotalAmount <= 25000000 { // 25M IDR
		// Add finance users (you would query from database)
		approvers = append(approvers, s.getFinanceUserIDs()...)
	} else {
		// Add director users
		approvers = append(approvers, s.getDirectorUserIDs()...)
	}
	
	return approvers
}

func (s *NotificationService) getFinanceUserIDs() []uint {
	// This should query the database for users with finance role
	// For now, return dummy data
	return []uint{2} // Assuming user ID 2 is finance
}

func (s *NotificationService) getDirectorUserIDs() []uint {
	// This should query the database for users with director role
	// For now, return dummy data
	return []uint{3} // Assuming user ID 3 is director
}

// CleanupOldNotifications removes old notifications
func (s *NotificationService) CleanupOldNotifications(daysOld int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)
	return s.notificationRepo.DeleteOlderThan(cutoffDate)
}

// GetNotificationStats gets notification statistics
func (s *NotificationService) GetNotificationStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get total notifications
	total, err := s.notificationRepo.GetTotalCount(userID)
	if err != nil {
		return nil, err
	}
	stats["total_notifications"] = total
	
	// Get unread count
	unread, err := s.GetUnreadCount(userID)
	if err != nil {
		return nil, err
	}
	stats["unread_notifications"] = unread
	
	// Get count by type
	approvalPending, _ := s.notificationRepo.GetCountByType(userID, models.NotificationTypeApprovalPending)
	approvalApproved, _ := s.notificationRepo.GetCountByType(userID, models.NotificationTypeApprovalApproved)
	approvalRejected, _ := s.notificationRepo.GetCountByType(userID, models.NotificationTypeApprovalRejected)
	
	stats["approval_pending"] = approvalPending
	stats["approval_approved"] = approvalApproved
	stats["approval_rejected"] = approvalRejected
	
	return stats, nil
}
