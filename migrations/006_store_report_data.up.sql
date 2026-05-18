CREATE TABLE report_calculations (
    id SERIAL PRIMARY KEY,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    calculation_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE reports
    ADD COLUMN calculation_id INTEGER REFERENCES report_calculations(id) ON DELETE CASCADE,
    ADD COLUMN onec_report_data JSONB;
