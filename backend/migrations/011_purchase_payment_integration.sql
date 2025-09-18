-- Migration 011: Purchase Payment Integration
-- Add payment tracking and create purchase_payments table for cross-reference

BEGIN;

-- Ensure purchase payment fields exist (redundant but safe)
ALTER TABLE purchases 
ADD COLUMN IF NOT EXISTS paid_amount DECIMAL(15,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS outstanding_amount DECIMAL(15,2) DEFAULT 0;

-- Initialize outstanding_amount for existing CREDIT purchases
UPDATE purchases 
SET outstanding_amount = total_amount 
WHERE payment_method = 'CREDIT' AND outstanding_amount = 0;

-- Initialize outstanding_amount to 0 for CASH/TRANSFER purchases (already paid)
UPDATE purchases 
SET outstanding_amount = 0,
    paid_amount = total_amount 
WHERE payment_method IN ('CASH', 'TRANSFER') AND outstanding_amount != 0;

-- Create purchase_payments table for cross-reference tracking
CREATE TABLE IF NOT EXISTS purchase_payments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    purchase_id INT NOT NULL,
    payment_number VARCHAR(50),
    date DATETIME,
    amount DECIMAL(15,2) DEFAULT 0,
    method VARCHAR(20),
    reference VARCHAR(100),
    notes TEXT,
    cash_bank_id INT,
    user_id INT NOT NULL,
    payment_id INT, -- Cross-reference to payments table
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (purchase_id) REFERENCES purchases(id) ON DELETE CASCADE,
    FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL,
    FOREIGN KEY (cash_bank_id) REFERENCES cash_banks(id) ON DELETE SET NULL,
    
    INDEX idx_purchase_payments_purchase_id (purchase_id),
    INDEX idx_purchase_payments_payment_id (payment_id),
    INDEX idx_purchase_payments_date (date)
);

-- Enhance payment_allocations to support purchase bills
ALTER TABLE payment_allocations 
ADD COLUMN IF NOT EXISTS bill_id INT,
ADD INDEX IF NOT EXISTS idx_payment_allocations_bill_id (bill_id);

-- Add foreign key constraint for bill_id
ALTER TABLE payment_allocations 
ADD CONSTRAINT IF NOT EXISTS fk_payment_allocations_bill 
    FOREIGN KEY (bill_id) REFERENCES purchases(id) ON DELETE SET NULL;

-- Comments for documentation
COMMENT ON COLUMN purchases.paid_amount IS 'Total amount paid for this purchase';
COMMENT ON COLUMN purchases.outstanding_amount IS 'Remaining amount to be paid (total_amount - paid_amount)';
COMMENT ON TABLE purchase_payments IS 'Cross-reference table linking purchases to payment management records';
COMMENT ON COLUMN payment_allocations.bill_id IS 'Reference to purchase (bill) for payable payments';

COMMIT;