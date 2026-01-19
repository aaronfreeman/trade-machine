-- Create test database for running tests
-- This script runs automatically when the PostgreSQL container is first initialized

CREATE DATABASE trademachine_test;

-- Grant all privileges on test database to the main user
GRANT ALL PRIVILEGES ON DATABASE trademachine_test TO trademachine;
