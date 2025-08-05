package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"strconv"
	"fmt"
)

// ContactService provides business logic for contacts
type ContactService interface {
	GetAllContacts() ([]models.Contact, error)
	GetContactByID(id string) (*models.Contact, error)
	CreateContact(contact models.Contact) (*models.Contact, error)
	UpdateContact(id string, contact models.Contact) (*models.Contact, error)
	DeleteContact(id string) error
	GetContactsByType(contactType string) ([]models.Contact, error)
	ImportContacts(contacts []models.Contact) error
	ExportContacts() ([]models.Contact, error)
}

// contactService implements ContactService
type contactService struct {
	repo repositories.ContactRepository
}

// NewContactService creates a new ContactService
func NewContactService(repo repositories.ContactRepository) ContactService {
	return &contactService{repo: repo}
}

// GetAllContacts returns all contacts
func (s *contactService) GetAllContacts() ([]models.Contact, error) {
	return s.repo.GetAll()
}

// GetContactByID returns a contact by ID
func (s *contactService) GetContactByID(id string) (*models.Contact, error) {
	contactID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid contact ID: %v", err)
	}
	return s.repo.GetByID(uint(contactID))
}

// CreateContact creates a new contact
func (s *contactService) CreateContact(contact models.Contact) (*models.Contact, error) {
	// Generate contact code if not provided
	if contact.Code == "" {
		code, err := s.generateContactCode(contact.Type)
		if err != nil {
			return nil, err
		}
		contact.Code = code
	}

	// Validate contact type
	if !isValidContactType(contact.Type) {
		return nil, fmt.Errorf("invalid contact type: %s", contact.Type)
	}

	return s.repo.Create(contact)
}

// UpdateContact updates an existing contact
func (s *contactService) UpdateContact(id string, contact models.Contact) (*models.Contact, error) {
	contactID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid contact ID: %v", err)
	}

	// Validate contact type
	if !isValidContactType(contact.Type) {
		return nil, fmt.Errorf("invalid contact type: %s", contact.Type)
	}

	contact.ID = uint(contactID)
	return s.repo.Update(contact)
}

// DeleteContact deletes a contact by ID
func (s *contactService) DeleteContact(id string) error {
	contactID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid contact ID: %v", err)
	}
	return s.repo.Delete(uint(contactID))
}

// GetContactsByType returns contacts filtered by type
func (s *contactService) GetContactsByType(contactType string) ([]models.Contact, error) {
	if !isValidContactType(contactType) {
		return nil, fmt.Errorf("invalid contact type: %s", contactType)
	}
	return s.repo.GetByType(contactType)
}

// ImportContacts imports multiple contacts
func (s *contactService) ImportContacts(contacts []models.Contact) error {
	for i, contact := range contacts {
		// Generate contact code if not provided
		if contact.Code == "" {
			code, err := s.generateContactCode(contact.Type)
			if err != nil {
				return fmt.Errorf("error generating code for contact %d: %v", i+1, err)
			}
			contacts[i].Code = code
		}

		// Validate contact type
		if !isValidContactType(contact.Type) {
			return fmt.Errorf("invalid contact type for contact %d: %s", i+1, contact.Type)
		}
	}

	return s.repo.BulkCreate(contacts)
}

// ExportContacts exports all contacts
func (s *contactService) ExportContacts() ([]models.Contact, error) {
	return s.repo.GetAll()
}

// generateContactCode generates a unique contact code based on type
func (s *contactService) generateContactCode(contactType string) (string, error) {
	var prefix string
	switch contactType {
	case models.ContactTypeCustomer:
		prefix = "CUST"
	case models.ContactTypeVendor:
		prefix = "VEND"
	case models.ContactTypeEmployee:
		prefix = "EMP"
	default:
		return "", fmt.Errorf("invalid contact type: %s", contactType)
	}

	// Get the count of contacts of this type
	count, err := s.repo.CountByType(contactType)
	if err != nil {
		return "", err
	}

	// Generate code with format: PREFIX-YYYYMMDD-XXX
	return fmt.Sprintf("%s-%04d", prefix, count+1), nil
}

// isValidContactType validates contact type
func isValidContactType(contactType string) bool {
	validTypes := []string{
		models.ContactTypeCustomer,
		models.ContactTypeVendor,
		models.ContactTypeEmployee,
	}

	for _, validType := range validTypes {
		if contactType == validType {
			return true
		}
	}
	return false
}
