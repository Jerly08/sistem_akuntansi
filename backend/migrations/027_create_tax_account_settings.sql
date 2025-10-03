-- Migration: Create tax_account_settings table
-- Description: Create table for storing tax account configuration settings
-- Version: 027
-- Created: 2024-10-03

CREATE TABLE IF NOT EXISTS tax_account_settings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,

    -- Sales Account Configuration (required)
    sales_receivable_account_id BIGINT UNSIGNED NOT NULL,
    sales_cash_account_id BIGINT UNSIGNED NOT NULL,
    sales_bank_account_id BIGINT UNSIGNED NOT NULL,
    sales_revenue_account_id BIGINT UNSIGNED NOT NULL,
    sales_output_vat_account_id BIGINT UNSIGNED NOT NULL,

    -- Purchase Account Configuration (required)
    purchase_payable_account_id BIGINT UNSIGNED NOT NULL,
    purchase_cash_account_id BIGINT UNSIGNED NOT NULL,
    purchase_bank_account_id BIGINT UNSIGNED NOT NULL,
    purchase_input_vat_account_id BIGINT UNSIGNED NOT NULL,
    purchase_expense_account_id BIGINT UNSIGNED NOT NULL,

    -- Other Tax Accounts (optional)
    withholding_tax21_account_id BIGINT UNSIGNED NULL DEFAULT NULL,
    withholding_tax23_account_id BIGINT UNSIGNED NULL DEFAULT NULL,
    withholding_tax25_account_id BIGINT UNSIGNED NULL DEFAULT NULL,
    tax_payable_account_id BIGINT UNSIGNED NULL DEFAULT NULL,

    -- Inventory Account (optional)
    inventory_account_id BIGINT UNSIGNED NULL DEFAULT NULL,
    cogs_account_id BIGINT UNSIGNED NULL DEFAULT NULL,

    -- Configuration flags
    is_active BOOLEAN DEFAULT TRUE,
    apply_to_all_companies BOOLEAN DEFAULT TRUE,

    -- Metadata
    updated_by BIGINT UNSIGNED NOT NULL,
    notes TEXT NULL,

    -- Indexes
    INDEX idx_tax_account_settings_deleted_at (deleted_at),
    INDEX idx_tax_account_settings_is_active (is_active),
    INDEX idx_tax_account_settings_updated_by (updated_by),

    -- Foreign key constraints for Sales accounts
    INDEX idx_tax_settings_sales_receivable (sales_receivable_account_id),
    INDEX idx_tax_settings_sales_cash (sales_cash_account_id),
    INDEX idx_tax_settings_sales_bank (sales_bank_account_id),
    INDEX idx_tax_settings_sales_revenue (sales_revenue_account_id),
    INDEX idx_tax_settings_sales_output_vat (sales_output_vat_account_id),

    -- Foreign key constraints for Purchase accounts
    INDEX idx_tax_settings_purchase_payable (purchase_payable_account_id),
    INDEX idx_tax_settings_purchase_cash (purchase_cash_account_id),
    INDEX idx_tax_settings_purchase_bank (purchase_bank_account_id),
    INDEX idx_tax_settings_purchase_input_vat (purchase_input_vat_account_id),
    INDEX idx_tax_settings_purchase_expense (purchase_expense_account_id),

    -- Foreign key constraints for optional accounts
    INDEX idx_tax_settings_withholding_tax21 (withholding_tax21_account_id),
    INDEX idx_tax_settings_withholding_tax23 (withholding_tax23_account_id),
    INDEX idx_tax_settings_withholding_tax25 (withholding_tax25_account_id),
    INDEX idx_tax_settings_tax_payable (tax_payable_account_id),
    INDEX idx_tax_settings_inventory (inventory_account_id),
    INDEX idx_tax_settings_cogs (cogs_account_id),

    -- Foreign key constraints
    CONSTRAINT fk_tax_settings_sales_receivable 
        FOREIGN KEY (sales_receivable_account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_tax_settings_sales_cash 
        FOREIGN KEY (sales_cash_account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_tax_settings_sales_bank 
        FOREIGN KEY (sales_bank_account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_tax_settings_sales_revenue 
        FOREIGN KEY (sales_revenue_account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_tax_settings_sales_output_vat 
        FOREIGN KEY (sales_output_vat_account_id) REFERENCES accounts(id) ON DELETE RESTRICT,

    CONSTRAINT fk_tax_settings_purchase_payable 
        FOREIGN KEY (purchase_payable_account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_tax_settings_purchase_cash 
        FOREIGN KEY (purchase_cash_account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_tax_settings_purchase_bank 
        FOREIGN KEY (purchase_bank_account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_tax_settings_purchase_input_vat 
        FOREIGN KEY (purchase_input_vat_account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_tax_settings_purchase_expense 
        FOREIGN KEY (purchase_expense_account_id) REFERENCES accounts(id) ON DELETE RESTRICT,

    CONSTRAINT fk_tax_settings_withholding_tax21 
        FOREIGN KEY (withholding_tax21_account_id) REFERENCES accounts(id) ON DELETE SET NULL,
    CONSTRAINT fk_tax_settings_withholding_tax23 
        FOREIGN KEY (withholding_tax23_account_id) REFERENCES accounts(id) ON DELETE SET NULL,
    CONSTRAINT fk_tax_settings_withholding_tax25 
        FOREIGN KEY (withholding_tax25_account_id) REFERENCES accounts(id) ON DELETE SET NULL,
    CONSTRAINT fk_tax_settings_tax_payable 
        FOREIGN KEY (tax_payable_account_id) REFERENCES accounts(id) ON DELETE SET NULL,
    CONSTRAINT fk_tax_settings_inventory 
        FOREIGN KEY (inventory_account_id) REFERENCES accounts(id) ON DELETE SET NULL,
    CONSTRAINT fk_tax_settings_cogs 
        FOREIGN KEY (cogs_account_id) REFERENCES accounts(id) ON DELETE SET NULL,
    CONSTRAINT fk_tax_settings_updated_by 
        FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default configuration based on current hardcoded values
INSERT INTO tax_account_settings (
    sales_receivable_account_id,
    sales_cash_account_id,
    sales_bank_account_id,
    sales_revenue_account_id,
    sales_output_vat_account_id,
    purchase_payable_account_id,
    purchase_cash_account_id,
    purchase_bank_account_id,
    purchase_input_vat_account_id,
    purchase_expense_account_id,
    is_active,
    apply_to_all_companies,
    updated_by,
    notes
) 
SELECT 
    -- Sales accounts (based on hardcoded values in services)
    COALESCE((SELECT id FROM accounts WHERE code = '1201' AND is_active = 1 LIMIT 1), 1) as sales_receivable_account_id,
    COALESCE((SELECT id FROM accounts WHERE code = '1101' AND is_active = 1 LIMIT 1), 1) as sales_cash_account_id,
    COALESCE((SELECT id FROM accounts WHERE code = '1102' AND is_active = 1 LIMIT 1), 1) as sales_bank_account_id,
    COALESCE((SELECT id FROM accounts WHERE code = '4101' AND is_active = 1 LIMIT 1), 1) as sales_revenue_account_id,
    COALESCE((SELECT id FROM accounts WHERE code = '2103' AND is_active = 1 LIMIT 1), 1) as sales_output_vat_account_id,
    
    -- Purchase accounts (based on hardcoded values in services)
    COALESCE((SELECT id FROM accounts WHERE code = '2001' AND is_active = 1 LIMIT 1), 1) as purchase_payable_account_id,
    COALESCE((SELECT id FROM accounts WHERE code = '1101' AND is_active = 1 LIMIT 1), 1) as purchase_cash_account_id,
    COALESCE((SELECT id FROM accounts WHERE code = '1102' AND is_active = 1 LIMIT 1), 1) as purchase_bank_account_id,
    COALESCE((SELECT id FROM accounts WHERE code = '1105' AND is_active = 1 LIMIT 1), 1) as purchase_input_vat_account_id,
    COALESCE((SELECT id FROM accounts WHERE code = '6001' AND is_active = 1 LIMIT 1), 1) as purchase_expense_account_id,
    
    -- Configuration
    TRUE as is_active,
    TRUE as apply_to_all_companies,
    1 as updated_by, -- System user
    'Default configuration based on existing hardcoded values' as notes
WHERE NOT EXISTS (SELECT 1 FROM tax_account_settings WHERE is_active = 1);

-- Add comment to table
ALTER TABLE tax_account_settings COMMENT = 'Configuration table for tax account mappings used in sales and purchase transactions';

-- Log migration
INSERT INTO migrations (migration, batch, executed_at) 
VALUES ('027_create_tax_account_settings.sql', 27, NOW())
ON DUPLICATE KEY UPDATE executed_at = NOW();