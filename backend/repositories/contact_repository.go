package repositories

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// ContactRepository provides database operations for contacts
type ContactRepository interface {
	GetAll() ([]models.Contact, error)
	GetByID(id uint) (*models.Contact, error)
	GetByCode(code string) (*models.Contact, error)
	GetByType(contactType string) ([]models.Contact, error)
	Create(contact models.Contact) (*models.Contact, error)
	Update(contact models.Contact) (*models.Contact, error)
	Delete(id uint) error
	CountByType(contactType string) (int64, error)
	BulkCreate(contacts []models.Contact) error
	Search(query string) ([]models.Contact, error)
}

// contactRepository implements ContactRepository
type contactRepository struct {
	db *gorm.DB
}

// NewContactRepository creates a new ContactRepository
func NewContactRepository(db *gorm.DB) ContactRepository {
	return &contactRepository{db: db}
}

// GetAll returns all contacts with their addresses
func (r *contactRepository) GetAll() ([]models.Contact, error) {
	var contacts []models.Contact
	err := r.db.Preload("Addresses").Find(&contacts).Error
	return contacts, err
}

// GetByID returns a contact by ID with its addresses
func (r *contactRepository) GetByID(id uint) (*models.Contact, error) {
	var contact models.Contact
	err := r.db.Preload("Addresses").First(&contact, id).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// GetByCode returns a contact by code with its addresses
func (r *contactRepository) GetByCode(code string) (*models.Contact, error) {
	var contact models.Contact
	err := r.db.Preload("Addresses").Where("code = ?", code).First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// GetByType returns contacts filtered by type
func (r *contactRepository) GetByType(contactType string) ([]models.Contact, error) {
	var contacts []models.Contact
	err := r.db.Preload("Addresses").Where("type = ?", contactType).Find(&contacts).Error
	return contacts, err
}

// Create creates a new contact
func (r *contactRepository) Create(contact models.Contact) (*models.Contact, error) {
	err := r.db.Create(&contact).Error
	if err != nil {
		return nil, err
	}
	
	// Load the contact with addresses
	return r.GetByID(contact.ID)
}

// Update updates an existing contact
func (r *contactRepository) Update(contact models.Contact) (*models.Contact, error) {
	err := r.db.Save(&contact).Error
	if err != nil {
		return nil, err
	}
	
	// Load the contact with addresses
	return r.GetByID(contact.ID)
}

// Delete deletes a contact by ID (soft delete)
func (r *contactRepository) Delete(id uint) error {
	return r.db.Delete(&models.Contact{}, id).Error
}

// CountByType returns the count of contacts by type
func (r *contactRepository) CountByType(contactType string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Contact{}).Where("type = ?", contactType).Count(&count).Error
	return count, err
}

// BulkCreate creates multiple contacts
func (r *contactRepository) BulkCreate(contacts []models.Contact) error {
	return r.db.CreateInBatches(contacts, 100).Error
}

// Search searches contacts by name, email, or phone
func (r *contactRepository) Search(query string) ([]models.Contact, error) {
	var contacts []models.Contact
	searchPattern := "%" + query + "%"
	
	err := r.db.Preload("Addresses").Where(
		"name ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR mobile ILIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	).Find(&contacts).Error
	
	return contacts, err
}
