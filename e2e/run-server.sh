#!/bin/bash
set -e

export E2E_DATABASE_URL="postgres://trademachine_test:test_password@localhost:5433/trademachine_test?sslmode=disable"
exec go run ./cmd/e2e-server
