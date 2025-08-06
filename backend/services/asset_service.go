package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"errors"
	"math"
	"strconv"
	"time"
)

type AssetServiceInterface interface {
	GetAllAssets() ([]models.Asset, error)
	GetAssetByID(id uint) (*models.Asset, error)
	CreateAsset(asset *models.Asset) error
	UpdateAsset(asset *models.Asset) error
	DeleteAsset(id uint) error
	GenerateAssetCode(category string) (string, error)
	CalculateDepreciation(asset *models.Asset, asOfDate time.Time) (float64, error)
	GetDepreciationSchedule(asset *models.Asset) ([]DepreciationEntry, error)
	GetAssetsSummary() (*AssetsSummary, error)
	GetAssetsForDepreciationReport() ([]AssetDepreciationReport, error)
}

type AssetService struct {
	assetRepo repositories.AssetRepositoryInterface
}

type DepreciationEntry struct {
	Year             int       `json:"year"`
	Date             time.Time `json:"date"`
	DepreciationCost float64   `json:"depreciation_cost"`
	AccumulatedDepreciation float64 `json:"accumulated_depreciation"`
	BookValue        float64   `json:"book_value"`
}

type AssetsSummary struct {
	TotalAssets      int64   `json:"total_assets"`
	ActiveAssets     int64   `json:"active_assets"`
	TotalValue       float64 `json:"total_value"`
	TotalDepreciation float64 `json:"total_depreciation"`
	NetBookValue     float64 `json:"net_book_value"`
}

type AssetDepreciationReport struct {
	Asset                   models.Asset `json:"asset"`
	AnnualDepreciation      float64      `json:"annual_depreciation"`
	MonthlyDepreciation     float64      `json:"monthly_depreciation"`
	RemainingDepreciation   float64      `json:"remaining_depreciation"`
	RemainingYears          int          `json:"remaining_years"`
	CurrentBookValue        float64      `json:"current_book_value"`
}

func NewAssetService(assetRepo repositories.AssetRepositoryInterface) AssetServiceInterface {
	return &AssetService{
		assetRepo: assetRepo,
	}
}

// GetAllAssets retrieves all assets
func (s *AssetService) GetAllAssets() ([]models.Asset, error) {
	return s.assetRepo.FindAll()
}

// GetAssetByID retrieves asset by ID
func (s *AssetService) GetAssetByID(id uint) (*models.Asset, error) {
	return s.assetRepo.FindByID(id)
}

// CreateAsset creates a new asset with generated code
func (s *AssetService) CreateAsset(asset *models.Asset) error {
	// Generate asset code if not provided
	if asset.Code == "" {
		code, err := s.GenerateAssetCode(asset.Category)
		if err != nil {
			return err
		}
		asset.Code = code
	}

	// Set default status
	if asset.Status == "" {
		asset.Status = models.AssetStatusActive
	}

	// Validate purchase date
	if asset.PurchaseDate.IsZero() {
		return errors.New("purchase date is required")
	}

	// Validate purchase price
	if asset.PurchasePrice <= 0 {
		return errors.New("purchase price must be greater than 0")
	}

	// Validate useful life for depreciable assets
	if asset.UsefulLife <= 0 && asset.DepreciationMethod != "" {
		return errors.New("useful life must be greater than 0 for depreciable assets")
	}

	// Calculate initial depreciation if needed
	if asset.AccumulatedDepreciation == 0 && asset.DepreciationMethod != "" {
		depreciation, err := s.CalculateDepreciation(asset, time.Now())
		if err == nil {
			asset.AccumulatedDepreciation = depreciation
		}
	}

	return s.assetRepo.Create(asset)
}

// UpdateAsset updates an existing asset
func (s *AssetService) UpdateAsset(asset *models.Asset) error {
	// Validate purchase date
	if asset.PurchaseDate.IsZero() {
		return errors.New("purchase date is required")
	}

	// Validate purchase price
	if asset.PurchasePrice <= 0 {
		return errors.New("purchase price must be greater than 0")
	}

	// Recalculate depreciation if depreciation-related fields changed
	if asset.DepreciationMethod != "" && asset.UsefulLife > 0 {
		depreciation, err := s.CalculateDepreciation(asset, time.Now())
		if err == nil {
			asset.AccumulatedDepreciation = depreciation
		}
	}

	return s.assetRepo.Update(asset)
}

// DeleteAsset deletes an asset
func (s *AssetService) DeleteAsset(id uint) error {
	return s.assetRepo.Delete(id)
}

// GenerateAssetCode generates a unique asset code based on category
func (s *AssetService) GenerateAssetCode(category string) (string, error) {
	// Get category prefix
	prefix := getCategoryPrefix(category)
	
	// Get current count for this category
	assets, err := s.assetRepo.GetAssetsByCategory(category)
	if err != nil {
		return "", err
	}
	
	// Generate code with sequence number
	sequence := len(assets) + 1
	code := prefix + "-" + time.Now().Format("2006") + "-" + padLeft(strconv.Itoa(sequence), 3, "0")
	
	// Check if code already exists and increment if needed
	for {
		_, err := s.assetRepo.FindByCode(code)
		if err != nil {
			// Code doesn't exist, we can use it
			break
		}
		sequence++
		code = prefix + "-" + time.Now().Format("2006") + "-" + padLeft(strconv.Itoa(sequence), 3, "0")
	}
	
	return code, nil
}

// CalculateDepreciation calculates accumulated depreciation up to a specific date
func (s *AssetService) CalculateDepreciation(asset *models.Asset, asOfDate time.Time) (float64, error) {
	if asset.UsefulLife <= 0 || asset.PurchasePrice <= 0 {
		return 0, nil
	}

	// Calculate months since purchase
	monthsSincePurchase := monthsDifference(asset.PurchaseDate, asOfDate)
	if monthsSincePurchase <= 0 {
		return 0, nil
	}

	depreciableAmount := asset.PurchasePrice - asset.SalvageValue
	
	switch asset.DepreciationMethod {
	case models.DepreciationMethodStraightLine:
		return s.calculateStraightLineDepreciation(depreciableAmount, asset.UsefulLife, monthsSincePurchase), nil
	case models.DepreciationMethodDecliningBalance:
		return s.calculateDecliningBalanceDepreciation(asset.PurchasePrice, asset.SalvageValue, asset.UsefulLife, monthsSincePurchase), nil
	default:
		return s.calculateStraightLineDepreciation(depreciableAmount, asset.UsefulLife, monthsSincePurchase), nil
	}
}

// GetDepreciationSchedule generates depreciation schedule for an asset
func (s *AssetService) GetDepreciationSchedule(asset *models.Asset) ([]DepreciationEntry, error) {
	if asset.UsefulLife <= 0 || asset.PurchasePrice <= 0 {
		return []DepreciationEntry{}, nil
	}

	var schedule []DepreciationEntry
	
	for year := 1; year <= asset.UsefulLife; year++ {
		date := asset.PurchaseDate.AddDate(year-1, 0, 0)
		
		depreciation, _ := s.CalculateDepreciation(asset, date.AddDate(1, 0, -1))
		
		var annualDepreciation float64
		if year == 1 {
			annualDepreciation = depreciation
		} else {
			prevDepreciation, _ := s.CalculateDepreciation(asset, date.AddDate(0, 0, -1))
			annualDepreciation = depreciation - prevDepreciation
		}
		
		bookValue := asset.PurchasePrice - depreciation
		if bookValue < asset.SalvageValue {
			bookValue = asset.SalvageValue
		}
		
		entry := DepreciationEntry{
			Year:                    year,
			Date:                    date,
			DepreciationCost:        annualDepreciation,
			AccumulatedDepreciation: depreciation,
			BookValue:               bookValue,
		}
		
		schedule = append(schedule, entry)
		
		// Stop if we've reached salvage value
		if bookValue <= asset.SalvageValue {
			break
		}
	}
	
	return schedule, nil
}

// GetAssetsSummary returns summary statistics for all assets
func (s *AssetService) GetAssetsSummary() (*AssetsSummary, error) {
	totalAssets, err := s.assetRepo.Count()
	if err != nil {
		return nil, err
	}

	activeAssets, err := s.assetRepo.GetActiveAssets()
	if err != nil {
		return nil, err
	}

	totalValue, err := s.assetRepo.GetTotalValue()
	if err != nil {
		return nil, err
	}

	// Calculate total depreciation and net book value
	var totalDepreciation, netBookValue float64
	for _, asset := range activeAssets {
		totalDepreciation += asset.AccumulatedDepreciation
		netBookValue += (asset.PurchasePrice - asset.AccumulatedDepreciation)
	}

	return &AssetsSummary{
		TotalAssets:       totalAssets,
		ActiveAssets:      int64(len(activeAssets)),
		TotalValue:        totalValue,
		TotalDepreciation: totalDepreciation,
		NetBookValue:      netBookValue,
	}, nil
}

// GetAssetsForDepreciationReport returns depreciation report for all assets
func (s *AssetService) GetAssetsForDepreciationReport() ([]AssetDepreciationReport, error) {
	assets, err := s.assetRepo.GetAssetsForDepreciation()
	if err != nil {
		return nil, err
	}

	var reports []AssetDepreciationReport
	
	for _, asset := range assets {
		// Calculate annual depreciation
		depreciableAmount := asset.PurchasePrice - asset.SalvageValue
		var annualDepreciation float64
		
		if asset.DepreciationMethod == models.DepreciationMethodStraightLine {
			annualDepreciation = depreciableAmount / float64(asset.UsefulLife)
		} else {
			// For declining balance, use first year depreciation as estimate
			annualDepreciation = depreciableAmount * 0.2 // 20% declining balance
		}
		
		monthlyDepreciation := annualDepreciation / 12
		currentBookValue := asset.PurchasePrice - asset.AccumulatedDepreciation
		remainingDepreciation := math.Max(0, currentBookValue - asset.SalvageValue)
		
		var remainingYears int
		if annualDepreciation > 0 {
			remainingYears = int(math.Ceil(remainingDepreciation / annualDepreciation))
		}
		
		report := AssetDepreciationReport{
			Asset:                 asset,
			AnnualDepreciation:    annualDepreciation,
			MonthlyDepreciation:   monthlyDepreciation,
			RemainingDepreciation: remainingDepreciation,
			RemainingYears:        remainingYears,
			CurrentBookValue:      currentBookValue,
		}
		
		reports = append(reports, report)
	}
	
	return reports, nil
}

// Helper functions

func getCategoryPrefix(category string) string {
	switch category {
	case "Real Estate":
		return "RE"
	case "Computer Equipment":
		return "CE"
	case "Vehicle":
		return "VH"
	case "Office Equipment":
		return "OE"
	case "Furniture":
		return "FR"
	case "IT Infrastructure":
		return "IT"
	case "Machinery":
		return "MC"
	default:
		return "AS" // Generic Asset
	}
}

func padLeft(str string, length int, pad string) string {
	for len(str) < length {
		str = pad + str
	}
	return str
}

func monthsDifference(startDate, endDate time.Time) int {
	if endDate.Before(startDate) {
		return 0
	}
	
	months := int(endDate.Month()) - int(startDate.Month())
	months += (endDate.Year() - startDate.Year()) * 12
	
	// Add partial month if end date day is >= start date day
	if endDate.Day() >= startDate.Day() {
		months++
	}
	
	return months
}

func (s *AssetService) calculateStraightLineDepreciation(depreciableAmount float64, usefulLifeYears int, monthsSincePurchase int) float64 {
	monthlyDepreciation := depreciableAmount / float64(usefulLifeYears*12)
	return monthlyDepreciation * float64(monthsSincePurchase)
}

func (s *AssetService) calculateDecliningBalanceDepreciation(purchasePrice, salvageValue float64, usefulLifeYears int, monthsSincePurchase int) float64 {
	rate := 2.0 / float64(usefulLifeYears) // Double declining balance
	monthlyRate := rate / 12
	
	accumulated := 0.0
	bookValue := purchasePrice
	
	for month := 0; month < monthsSincePurchase; month++ {
		monthlyDepreciation := bookValue * monthlyRate
		
		// Don't depreciate below salvage value
		if bookValue-monthlyDepreciation < salvageValue {
			monthlyDepreciation = bookValue - salvageValue
		}
		
		accumulated += monthlyDepreciation
		bookValue -= monthlyDepreciation
		
		// Stop if we reach salvage value
		if bookValue <= salvageValue {
			break
		}
	}
	
	return accumulated
}
