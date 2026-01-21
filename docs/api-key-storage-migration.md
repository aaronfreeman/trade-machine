# API Key Storage Migration: File to Database

## Overview
This document describes the migration of API key storage from encrypted files to encrypted database storage in the trade-machine application.

## Problem Statement
The application previously stored user-entered API keys in an encrypted file (`~/.trade-machine/settings.enc`). This needed to be changed to store the keys in the database while maintaining encryption.

## Solution

### Architecture
The solution implements a hybrid approach that:
1. **Stores API keys in the database** when a repository connection is available
2. **Falls back to file-based storage** when no database is available
3. **Automatically migrates** existing file-based keys to the database on first run
4. **Maintains strong encryption** using AES-256-GCM with PBKDF2 key derivation

### Key Features
- ✅ AES-256-GCM encryption maintained
- ✅ Automatic file-to-database migration
- ✅ Backward compatible file-based mode
- ✅ 36 comprehensive tests passing
- ✅ No breaking API changes

## Implementation Details

### Database Schema
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY,
    service_name VARCHAR(50) UNIQUE,
    api_key_encrypted BYTEA,       -- Encrypted API key
    api_secret_encrypted BYTEA,    -- Encrypted API secret
    base_url VARCHAR(255),         -- Unencrypted config
    region VARCHAR(50),           
    model_id VARCHAR(100),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Migration Strategy
1. On startup, attempt to load keys from database
2. If database is empty, check for file-based keys
3. If file exists, automatically migrate to database
4. Original file is preserved for safety

### Testing
- 36 tests covering file and database modes
- Mock repository for isolated testing
- Comprehensive coverage of encryption, migration, and error scenarios

## Deployment

### New Installation
```bash
export DATABASE_URL="postgres://..."
goose -dir migrations up
./trade-machine
```

### Existing Installation
```bash
# Keys will auto-migrate on first run
export DATABASE_URL="postgres://..."
goose -dir migrations up  
./trade-machine
# Check migration: verify database has keys
# Optional: backup and remove ~/.trade-machine/settings.enc
```

## Security
- Encryption: AES-256-GCM with PBKDF2 (100k iterations)
- Each key encrypted separately
- Recommend using `SETTINGS_PASSPHRASE` environment variable
- Database connections should use SSL/TLS

## Files Modified
- `migrations/004_add_api_keys_table.sql` - Schema
- `repository/api_keys.go` - Database operations
- `internal/settings/settings.go` - Hybrid storage logic
- Tests updated with comprehensive coverage
