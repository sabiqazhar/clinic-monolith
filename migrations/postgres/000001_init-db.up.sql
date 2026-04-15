-- PATIENT MODULE
CREATE TABLE IF NOT EXISTS patients (
    id VARCHAR(36) PRIMARY KEY,
    full_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_patients_email ON patients(email);

-- BILLING MODULE
CREATE TABLE IF NOT EXISTS billing_invoices (
    id VARCHAR(36) PRIMARY KEY,
    patient_id VARCHAR(36) NOT NULL, -- NO FK constraint (isolasi ketat)
    amount NUMERIC(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending | paid | cancelled | refunded
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_billing_patient ON billing_invoices(patient_id);
CREATE INDEX IF NOT EXISTS idx_billing_status ON billing_invoices(status);

-- OUTBOX EVENTS (Shared untuk modul PG)
CREATE TABLE IF NOT EXISTS outbox_events (
    id VARCHAR(36) PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending | processed | failed
    attempts INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);
-- Partial index: hanya index row pending agar polling relay cepat
CREATE INDEX IF NOT EXISTS idx_outbox_pending ON outbox_events(created_at) 
WHERE status = 'pending';
