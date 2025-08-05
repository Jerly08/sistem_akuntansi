package controllers

import (
	"net/http"
	"encoding/json"
	"io"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
)

// ContactController handles HTTP requests for Contacts
type ContactController struct {
	contactService services.ContactService
}

// NewContactController creates a new ContactController
func NewContactController(contactService services.ContactService) *ContactController {
	return &ContactController{contactService: contactService}
}

// GetContacts returns a list of contacts
func (cc *ContactController) GetContacts(c *gin.Context) {
	contacts, err := cc.contactService.GetAllContacts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

// GetContact returns a contact by ID
func (cc *ContactController) GetContact(c *gin.Context) {
	id := c.Param("id")
	contact, err := cc.contactService.GetContactByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contact)
}

// CreateContact creates a new contact
func (cc *ContactController) CreateContact(c *gin.Context) {
	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newContact, err := cc.contactService.CreateContact(contact)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newContact)
}

// UpdateContact updates an existing contact by ID
func (cc *ContactController) UpdateContact(c *gin.Context) {
	id := c.Param("id")
	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedContact, err := cc.contactService.UpdateContact(id, contact)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedContact)
}

// DeleteContact deletes a contact by ID
func (cc *ContactController) DeleteContact(c *gin.Context) {
	id := c.Param("id")
	if err := cc.contactService.DeleteContact(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// GetContactsByType returns contacts filtered by type
func (cc *ContactController) GetContactsByType(c *gin.Context) {
	contactType := c.Param("type")
	contacts, err := cc.contactService.GetContactsByType(contactType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

// SearchContacts searches contacts by query
func (cc *ContactController) SearchContacts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}
	
	// For now, get all contacts and filter in service layer
	// In production, this should be handled in repository
	contacts, err := cc.contactService.GetAllContacts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

// ImportContacts imports contacts from JSON
func (cc *ContactController) ImportContacts(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	
	var contacts []models.Contact
	if err := json.Unmarshal(body, &contacts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	
	if err := cc.contactService.ImportContacts(contacts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Contacts imported successfully", "count": len(contacts)})
}

// ExportContacts exports all contacts as JSON
func (cc *ContactController) ExportContacts(c *gin.Context) {
	contacts, err := cc.contactService.ExportContacts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=contacts.json")
	c.JSON(http.StatusOK, contacts)
}

