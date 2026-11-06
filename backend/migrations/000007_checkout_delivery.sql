ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS delivery_method VARCHAR(20) NOT NULL DEFAULT 'PICKUP',
    ADD COLUMN IF NOT EXISTS delivery_address TEXT,
    ADD COLUMN IF NOT EXISTS delivery_fee NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    ADD COLUMN IF NOT EXISTS delivery_fee_status VARCHAR(30) NOT NULL DEFAULT 'NOT_REQUIRED';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'valid_order_delivery_method'
    ) THEN
        ALTER TABLE orders
            ADD CONSTRAINT valid_order_delivery_method
            CHECK (delivery_method IN ('PICKUP', 'DELIVERY'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'valid_order_delivery_fee_status'
    ) THEN
        ALTER TABLE orders
            ADD CONSTRAINT valid_order_delivery_fee_status
            CHECK (delivery_fee_status IN ('NOT_REQUIRED', 'TO_CONFIRM', 'CONFIRMED'));
    END IF;
END $$;
