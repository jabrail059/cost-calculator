ALTER TABLE reports
    DROP COLUMN IF EXISTS onec_report_data,
    DROP COLUMN IF EXISTS calculation_id;

DROP TABLE IF EXISTS report_calculations;
