-- WarBay Database Initialization
-- This script runs automatically when the postgres container starts for the first time.

-- Enable UUID generation extension (required for uuid_generate_v4())
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Optional: Create additional schemas or users if needed
-- CREATE USER warbay WITH PASSWORD 'warbay';
-- GRANT ALL PRIVILEGES ON DATABASE simple_ecommerce TO warbay;