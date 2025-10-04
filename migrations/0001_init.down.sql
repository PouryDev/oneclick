-- Drop trigger and function
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column ();

-- Drop users table
DROP TABLE IF EXISTS users;

-- Drop UUID extension (optional, as it might be used by other tables)
-- DROP EXTENSION IF EXISTS "uuid-ossp";