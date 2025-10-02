# Database Migration Debugging Guide for Railway

## Quick Diagnosis Steps

### Step 1: Check Railway Deployment Logs

1. Go to [Railway Dashboard](https://railway.app)
2. Select your backend service
3. Go to **Deployments** → Click latest deployment
4. Check both **Build Logs** and **Deploy Logs**

Look for errors containing:
- `ошибка при чтении директории миграций` (error reading migrations directory)
- `ошибка при выполнении миграции` (error executing migration)
- `Не удалось подключиться к БД` (failed to connect to DB)

### Step 2: Check Migrations Directory in Container

In Railway dashboard, go to your service and open the **Settings** tab:

```bash
# If Railway allows shell access, run:
ls -la /app/migrations
```

Expected output:
```
001_init_schema.sql
002_add_missing_columns.sql
003_add_chat_tables.sql
004_add_profile_photo_url.sql
005_add_sample_specializations.sql
006_add_dietologist_specializations.sql
007_remove_communication_method.sql
008_sync_chat_sessions_schema.sql
009_add_trainer_specializations.sql
```

### Step 3: Check Database Connection

In Railway dashboard → PostgreSQL service → **Connect**, verify:
- `POSTGRES_HOST`
- `POSTGRES_PORT`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `POSTGRES_DATABASE`

Make sure your backend has access to these variables.

### Step 4: Common Issues & Solutions

#### Issue 1: "Error reading migrations directory"
**Cause:** Migrations folder not copied or wrong path

**Solution:** Verify Dockerfile line 25:
```dockerfile
COPY --from=builder /app/migrations ./migrations
```

**Fix:** The code uses `./migrations` which should work. Check if files exist:
```go
// In main.go line 82, try absolute path:
if err := database.RunMigrations(db, "/app/migrations", logger); err != nil {
```

#### Issue 2: "Permission denied"
**Cause:** Non-root user can't read files

**Solution:** Update Dockerfile ownership:
```dockerfile
COPY --from=builder /app/migrations ./migrations
RUN chown -R appuser:appuser /app  # This line already exists at line 30
```

#### Issue 3: "pq: SSL not supported"
**Cause:** Wrong SSL mode for Railway

**Solution:** In Railway Variables, ensure:
```
POSTGRES_SSL_MODE=require
```

#### Issue 4: Specific Migration Fails
**Cause:** Migration has SQL errors or duplicate operations

**Solution:** Check which migration failed in logs, then:
1. Connect to Railway database directly
2. Check `migrations` table:
```sql
SELECT * FROM migrations ORDER BY version;
```
3. See which migrations have been applied
4. Fix the failing migration

### Step 5: Manual Database Check

Connect to your Railway PostgreSQL database:

```bash
# Get connection string from Railway dashboard
psql postgresql://user:password@host:port/database

# Check migrations table
\d migrations

# See applied migrations
SELECT * FROM migrations ORDER BY version;

# Check if tables exist
\dt
```

## Enhanced Logging Solution

To get better error messages, let's add more verbose logging:

### Option A: Add Debug Logging (Quick Fix)

Add this before running migrations in `main.go`:

```go
// Line 81, before running migrations
logger.Info("Checking migrations directory")
if _, err := os.Stat("./migrations"); os.IsNotExist(err) {
    logger.Error("Migrations directory does not exist", zap.Error(err))
    
    // Try absolute path
    if _, err := os.Stat("/app/migrations"); os.IsNotExist(err) {
        logger.Error("Migrations directory also not found at /app/migrations", zap.Error(err))
    } else {
        logger.Info("Found migrations at /app/migrations")
        // Use absolute path
        if err := database.RunMigrations(db, "/app/migrations", logger); err != nil {
            logger.Fatal("Ошибка при выполнении миграций", zap.Error(err))
        }
    }
} else {
    logger.Info("Found migrations at ./migrations")
    if err := database.RunMigrations(db, "./migrations", logger); err != nil {
        logger.Fatal("Ошибка при выполнении миграций", zap.Error(err))
    }
}
```

### Option B: Enhanced Migration Function

Update `pkg/database/migration.go` to add more logging:

```go
// Add at line 56, after reading directory
logger.Info("checking migrations directory", 
    zap.String("path", migrationsDir),
    zap.Int("files_found", len(files)),
)

if len(files) == 0 {
    logger.Warn("no files found in migrations directory")
}

// Log each file found
for _, file := range files {
    logger.Debug("found file", 
        zap.String("name", file.Name()),
        zap.Bool("is_dir", file.IsDir()),
    )
}
```

## Testing Locally with Docker

Build and test locally first:

```bash
cd /Users/user/Desktop/laps/backend

# Build the image
docker build -t laps-backend .

# Run with local database
docker run --rm \
  -e POSTGRES_HOST=host.docker.internal \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USERNAME=postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DBNAME=laps \
  -e POSTGRES_SSL_MODE=disable \
  -e HTTP_PORT=8080 \
  -e JWT_SIGNING_KEY=test-key \
  -p 8080:8080 \
  laps-backend

# Check if migrations run successfully
```

## Rollback Strategy

If you need to rollback migrations:

```sql
-- Connect to Railway database
-- Check migrations table
SELECT * FROM migrations ORDER BY version DESC;

-- Delete the last migration record (if needed)
DELETE FROM migrations WHERE version = '009';

-- Then manually revert the schema changes that migration made
-- (Check the migration file to see what needs to be reverted)
```

## Next Steps

1. **First**, check the Railway logs to see the exact error
2. **Then**, try one of the solutions above based on the error
3. **If still failing**, share the specific error message for more help

## Prevention

For future deployments:
1. Always test migrations locally with Docker first
2. Use Railway's preview deployments for testing
3. Consider using a migration tool like `golang-migrate` or `goose` for better control
4. Add health check endpoint that verifies database connectivity

