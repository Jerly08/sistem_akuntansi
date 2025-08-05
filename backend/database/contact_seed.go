package database

import (
	"log"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

// SeedContacts seeds initial contact data
func SeedContacts(db *gorm.DB) {
	log.Println("Seeding contacts...")

	contacts := []models.Contact{
		{
			Code:         "CUST-0001",
			Name:         "PT Maju Jaya",
			Type:         models.ContactTypeCustomer,
			Category:     models.CategoryRetail,
			Email:        "info@majujaya.com",
			Phone:        "+62-21-5551234",
			Mobile:       "+62-812-3456789",
			Website:      "www.majujaya.com",
			TaxNumber:    "01.234.567.8-901.000",
			CreditLimit:  50000000,
			PaymentTerms: 30,
			IsActive:     true,
			Notes:        "Customer utama untuk produk elektronik",
		},
		{
			Code:         "VEND-0001",
			Name:         "CV Sumber Rejeki",
			Type:         models.ContactTypeVendor,
			Category:     models.CategoryWholesale,
			Email:        "sales@sumberrejeki.co.id",
			Phone:        "+62-21-5555678",
			Mobile:       "+62-813-9876543",
			Website:      "www.sumberrejeki.co.id",
			TaxNumber:    "02.345.678.9-012.000",
			CreditLimit:  0,
			PaymentTerms: 14,
			IsActive:     true,
			Notes:        "Supplier utama untuk bahan baku",
		},
		{
			Code:         "EMP-0001",
			Name:         "Ahmad Subandi",
			Type:         models.ContactTypeEmployee,
			Email:        "ahmad.subandi@company.com",
			Phone:        "+62-21-5557890",
			Mobile:       "+62-812-3456789",
			CreditLimit:  0,
			PaymentTerms: 0,
			IsActive:     true,
			Notes:        "Manager Keuangan",
		},
		{
			Code:         "CUST-0002",
			Name:         "PT Global Tech",
			Type:         models.ContactTypeCustomer,
			Category:     models.CategoryWholesale,
			Email:        "contact@globaltech.id",
			Phone:        "+62-21-7771234",
			Mobile:       "+62-815-1234567",
			Website:      "www.globaltech.id",
			TaxNumber:    "03.456.789.0-123.000",
			CreditLimit:  75000000,
			PaymentTerms: 45,
			IsActive:     true,
			Notes:        "Customer wholesale teknologi",
		},
		{
			Code:         "VEND-0002",
			Name:         "Toko Elektronik Sejati",
			Type:         models.ContactTypeVendor,
			Category:     models.CategoryRetail,
			Email:        "admin@elektroniksejati.com",
			Phone:        "+62-21-6661111",
			Mobile:       "+62-814-5678901",
			TaxNumber:    "04.567.890.1-234.000",
			CreditLimit:  0,
			PaymentTerms: 7,
			IsActive:     true,
			Notes:        "Supplier komponen elektronik",
		},
		{
			Code:         "EMP-0002",
			Name:         "Siti Nurhaliza",
			Type:         models.ContactTypeEmployee,
			Email:        "siti.nurhaliza@company.com",
			Phone:        "+62-21-5559999",
			Mobile:       "+62-813-9876543",
			CreditLimit:  0,
			PaymentTerms: 0,
			IsActive:     true,
			Notes:        "Supervisor Inventory",
		},
	}

	// Seed addresses for contacts
	addresses := []models.ContactAddress{
		{
			ContactID:  1, // PT Maju Jaya
			Type:       models.AddressTypeBilling,
			Address1:   "Jl. Sudirman No. 123",
			City:       "Jakarta Pusat",
			State:      "DKI Jakarta",
			PostalCode: "10220",
			Country:    "Indonesia",
			IsDefault:  true,
		},
		{
			ContactID:  1, // PT Maju Jaya
			Type:       models.AddressTypeShipping,
			Address1:   "Jl. Sudirman No. 123",
			Address2:   "Gudang Belakang",
			City:       "Jakarta Pusat",
			State:      "DKI Jakarta",
			PostalCode: "10220",
			Country:    "Indonesia",
			IsDefault:  false,
		},
		{
			ContactID:  2, // CV Sumber Rejeki
			Type:       models.AddressTypeBilling,
			Address1:   "Jl. Gatot Subroto No. 456",
			City:       "Jakarta Selatan",
			State:      "DKI Jakarta",
			PostalCode: "12950",
			Country:    "Indonesia",
			IsDefault:  true,
		},
		{
			ContactID:  3, // Ahmad Subandi
			Type:       models.AddressTypeMailing,
			Address1:   "Jl. Kebon Jeruk No. 789",
			City:       "Jakarta Barat",
			State:      "DKI Jakarta",
			PostalCode: "11530",
			Country:    "Indonesia",
			IsDefault:  true,
		},
		{
			ContactID:  4, // PT Global Tech
			Type:       models.AddressTypeBilling,
			Address1:   "Jl. HR Rasuna Said No. 321",
			City:       "Jakarta Selatan",
			State:      "DKI Jakarta",
			PostalCode: "12940",
			Country:    "Indonesia",
			IsDefault:  true,
		},
		{
			ContactID:  5, // Toko Elektronik Sejati
			Type:       models.AddressTypeBilling,
			Address1:   "Jl. Mangga Besar No. 88",
			City:       "Jakarta Barat",
			State:      "DKI Jakarta",
			PostalCode: "11150",
			Country:    "Indonesia",
			IsDefault:  true,
		},
		{
			ContactID:  6, // Siti Nurhaliza
			Type:       models.AddressTypeMailing,
			Address1:   "Jl. Cempaka Putih No. 55",
			City:       "Jakarta Pusat",
			State:      "DKI Jakarta",
			PostalCode: "10570",
			Country:    "Indonesia",
			IsDefault:  true,
		},
	}

	// Check if contacts already exist
	var existingCount int64
	db.Model(&models.Contact{}).Count(&existingCount)
	
	if existingCount == 0 {
		// Create contacts
		if err := db.Create(&contacts).Error; err != nil {
			log.Printf("Error seeding contacts: %v", err)
		} else {
			log.Printf("Successfully seeded %d contacts", len(contacts))
		}

		// Create addresses
		if err := db.Create(&addresses).Error; err != nil {
			log.Printf("Error seeding contact addresses: %v", err)
		} else {
			log.Printf("Successfully seeded %d contact addresses", len(addresses))
		}
	} else {
		log.Printf("Contacts already exist (%d records), skipping seed", existingCount)
	}
}
