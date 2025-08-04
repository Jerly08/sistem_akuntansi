package database

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// SeedAccounts creates initial chart of accounts
func SeedAccounts(db *gorm.DB) error {
	accounts := []models.Account{
		// ASSETS (1xxx)
		{Code: "1000", Name: "AKTIVA", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 1, IsHeader: true, IsActive: true},
		{Code: "1100", Name: "AKTIVA LANCAR", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: true, IsActive: true},
		{Code: "1110", Name: "Kas", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1120", Name: "Bank", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1130", Name: "Piutang Usaha", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1140", Name: "Piutang Lain-lain", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1150", Name: "Persediaan", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1160", Name: "Biaya Dibayar Dimuka", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		
		{Code: "1200", Name: "AKTIVA TETAP", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 2, IsHeader: true, IsActive: true},
		{Code: "1210", Name: "Tanah", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1220", Name: "Bangunan", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1221", Name: "Akumulasi Penyusutan Bangunan", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 4, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1230", Name: "Peralatan", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1231", Name: "Akumulasi Penyusutan Peralatan", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 4, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1240", Name: "Kendaraan", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1241", Name: "Akumulasi Penyusutan Kendaraan", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 4, IsHeader: false, IsActive: true, Balance: 0},

		// LIABILITIES (2xxx)
		{Code: "2000", Name: "KEWAJIBAN", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 1, IsHeader: true, IsActive: true},
		{Code: "2100", Name: "KEWAJIBAN LANCAR", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 2, IsHeader: true, IsActive: true},
		{Code: "2110", Name: "Hutang Usaha", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "2120", Name: "Hutang Pajak", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "2130", Name: "Hutang Gaji", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "2140", Name: "Hutang Lain-lain", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		
		{Code: "2200", Name: "KEWAJIBAN JANGKA PANJANG", Type: models.AccountTypeLiability, Category: models.CategoryLongTermLiability, Level: 2, IsHeader: true, IsActive: true},
		{Code: "2210", Name: "Hutang Bank Jangka Panjang", Type: models.AccountTypeLiability, Category: models.CategoryLongTermLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

		// EQUITY (3xxx)
		{Code: "3000", Name: "MODAL", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 1, IsHeader: true, IsActive: true},
		{Code: "3100", Name: "Modal Pemilik", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "3200", Name: "Laba Ditahan", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "3300", Name: "Laba Tahun Berjalan", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0},

		// REVENUE (4xxx)
		{Code: "4000", Name: "PENDAPATAN", Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 1, IsHeader: true, IsActive: true},
		{Code: "4100", Name: "PENDAPATAN USAHA", Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 2, IsHeader: true, IsActive: true},
		{Code: "4110", Name: "Penjualan", Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "4120", Name: "Jasa", Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		
		{Code: "4200", Name: "PENDAPATAN LAIN-LAIN", Type: models.AccountTypeRevenue, Category: models.CategoryOtherRevenue, Level: 2, IsHeader: true, IsActive: true},
		{Code: "4210", Name: "Bunga Bank", Type: models.AccountTypeRevenue, Category: models.CategoryOtherRevenue, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "4220", Name: "Lain-lain", Type: models.AccountTypeRevenue, Category: models.CategoryOtherRevenue, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

		// EXPENSES (5xxx)
		{Code: "5000", Name: "BEBAN", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 1, IsHeader: true, IsActive: true},
		{Code: "5100", Name: "HARGA POKOK PENJUALAN", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: true, IsActive: true},
		{Code: "5110", Name: "Pembelian", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5120", Name: "Biaya Angkut Pembelian", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		
		{Code: "5200", Name: "BEBAN OPERASIONAL", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: true, IsActive: true},
		{Code: "5210", Name: "Beban Gaji", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5220", Name: "Beban Listrik", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5230", Name: "Beban Telepon", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5240", Name: "Beban Sewa", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5250", Name: "Beban Penyusutan", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5260", Name: "Beban Administrasi", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5270", Name: "Beban Pemasaran", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		
		{Code: "5300", Name: "BEBAN LAIN-LAIN", Type: models.AccountTypeExpense, Category: models.CategoryOtherExpense, Level: 2, IsHeader: true, IsActive: true},
		{Code: "5310", Name: "Beban Bunga", Type: models.AccountTypeExpense, Category: models.CategoryOtherExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5320", Name: "Beban Lain-lain", Type: models.AccountTypeExpense, Category: models.CategoryOtherExpense, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
	}

	// Set parent relationships
	for i, account := range accounts {
		if account.Level > 1 {
			// Find parent based on code pattern
			parentCode := account.Code[:len(account.Code)-1]
			if account.Level == 3 {
				parentCode = account.Code[:3] + "0"
			} else if account.Level == 4 {
				parentCode = account.Code[:3]
			}
			
			for _, parent := range accounts {
				if parent.Code == parentCode {
					accounts[i].ParentID = &parent.ID
					break
				}
			}
		}
	}

	// Create or update accounts
	for _, account := range accounts {
		var existingAccount models.Account
		result := db.Where("code = ?", account.Code).First(&existingAccount)
		
		if result.Error != nil {
			// Account doesn't exist, create it
			if err := db.Create(&account).Error; err != nil {
				return err
			}
		} else {
			// Account exists, update it
			existingAccount.Name = account.Name
			existingAccount.Type = account.Type
			existingAccount.Category = account.Category
			existingAccount.Level = account.Level
			existingAccount.IsHeader = account.IsHeader
			existingAccount.IsActive = account.IsActive
			existingAccount.Description = account.Description
			
			if err := db.Save(&existingAccount).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
