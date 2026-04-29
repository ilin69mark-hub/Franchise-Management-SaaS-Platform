-- 006_contracts.up.sql
-- Таблица договоров с клиентами

CREATE TABLE IF NOT EXISTS contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    salon_id UUID NOT NULL,
    lead_id UUID,
    manager_id UUID NOT NULL,

    client_name VARCHAR(255) NOT NULL,
    client_phone VARCHAR(50),
    client_email VARCHAR(255),

    total_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    prepaid_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    paid_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    remain_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    margin_percent NUMERIC(5,2) NOT NULL DEFAULT 0,

    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_status VARCHAR(50) NOT NULL DEFAULT 'awaiting_payment',

    payment_date TIMESTAMP,
    deadline_date TIMESTAMP,

    products TEXT,
    description TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contracts_salon_id ON contracts(salon_id);
CREATE INDEX idx_contracts_manager_id ON contracts(manager_id);
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_payment_status ON contracts(payment_status);
CREATE INDEX idx_contracts_payment_date ON contracts(payment_date);