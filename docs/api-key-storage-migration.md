# API Key Storage Migration: File to Database

## Overview
This document describes the migration of API key storage from encrypted files to encrypted database storage in the trade-machine application.

## Problem Statement
The application previously stored user-entered API keys in an encrypted file (`~/.trade-machine/settings.enc`). This has been changed to store the keys in the database while maintaining encryption. **Database storage is now required** - the application will not fall back to file-based storage.

## Solution

### Architecture
The solution implements database-only storage with one-time file migration:
1. **Stores API keys in the database** (repository is required)
2. **Automatically migrates** existing file-based keys to the database on first run
3. **Maintains strong encryption** using AES-256-GCM with PBKDF2 key derivation
4. **No file-based fallback** - if database is not available, the application will fail

### Key Features
- ✅ AES-256-GCM encryption maintained
- ✅ Automatic one-time file-to-database migration
- ✅ Database storage required (no fallback)
- ✅ 37 comprehensive tests passing
- ⚠️ Breaking change: Repository parameter now required

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
3. If file exists, automatically migrate to database (one-time)
4. Original file is preserved for safety

### Testing
- 37 tests covering database mode
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

### BREAKING CHANGE
**Repository is now required.** The application will fail to start without a database connection. The file-based fallback has been removed.

## Security
- Encryption: AES-256-GCM with PBKDF2 (100k iterations)
- Each key encrypted separately
- Recommend using `SETTINGS_PASSPHRASE` environment variable
- Database connections should use SSL/TLS

## Files Modified
- `migrations/004_add_api_keys_table.sql` - Schema
- `repository/api_keys.go` - Database operations
- `internal/settings/settings.go` - Database-only storage logic
- Tests updated with comprehensive coverage
