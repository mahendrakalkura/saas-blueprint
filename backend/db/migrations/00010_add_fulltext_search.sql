-- Add full-text search capabilities to the database

-- Create a custom text search configuration for better search
CREATE TEXT SEARCH CONFIGURATION english_unaccent (COPY = english);
ALTER TEXT SEARCH CONFIGURATION english_unaccent
  ALTER MAPPING FOR hword, hword_part, word WITH unaccent, english_stem;

-- Add full-text search columns and indexes for users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Generate search vector from name and email
CREATE OR REPLACE FUNCTION users_search_vector_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english_unaccent', COALESCE(NEW.name, '')), 'A') ||
    setweight(to_tsvector('english_unaccent', COALESCE(NEW.email, '')), 'B');
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search vector
DROP TRIGGER IF EXISTS users_search_vector_trigger ON users;
CREATE TRIGGER users_search_vector_trigger
  BEFORE INSERT OR UPDATE OF name, email ON users
  FOR EACH ROW EXECUTE FUNCTION users_search_vector_update();

-- Create GIN index for fast full-text search on users
CREATE INDEX IF NOT EXISTS users_search_vector_idx ON users USING GIN (search_vector);

-- Update existing users' search vectors
UPDATE users SET search_vector =
  setweight(to_tsvector('english_unaccent', COALESCE(name, '')), 'A') ||
  setweight(to_tsvector('english_unaccent', COALESCE(email, '')), 'B');


-- Add full-text search for organizations table
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Generate search vector from name and description
CREATE OR REPLACE FUNCTION organizations_search_vector_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english_unaccent', COALESCE(NEW.name, '')), 'A') ||
    setweight(to_tsvector('english_unaccent', COALESCE(NEW.description, '')), 'C');
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search vector
DROP TRIGGER IF EXISTS organizations_search_vector_trigger ON organizations;
CREATE TRIGGER organizations_search_vector_trigger
  BEFORE INSERT OR UPDATE OF name, description ON organizations
  FOR EACH ROW EXECUTE FUNCTION organizations_search_vector_update();

-- Create GIN index for fast full-text search on organizations
CREATE INDEX IF NOT EXISTS organizations_search_vector_idx ON organizations USING GIN (search_vector);

-- Update existing organizations' search vectors
UPDATE organizations SET search_vector =
  setweight(to_tsvector('english_unaccent', COALESCE(name, '')), 'A') ||
  setweight(to_tsvector('english_unaccent', COALESCE(description, '')), 'C');


-- Add full-text search for notifications table
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Generate search vector from title and message
CREATE OR REPLACE FUNCTION notifications_search_vector_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english_unaccent', COALESCE(NEW.title, '')), 'A') ||
    setweight(to_tsvector('english_unaccent', COALESCE(NEW.message, '')), 'B');
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search vector
DROP TRIGGER IF EXISTS notifications_search_vector_trigger ON notifications;
CREATE TRIGGER notifications_search_vector_trigger
  BEFORE INSERT OR UPDATE OF title, message ON notifications
  FOR EACH ROW EXECUTE FUNCTION notifications_search_vector_update();

-- Create GIN index for fast full-text search on notifications
CREATE INDEX IF NOT EXISTS notifications_search_vector_idx ON notifications USING GIN (search_vector);

-- Update existing notifications' search vectors
UPDATE notifications SET search_vector =
  setweight(to_tsvector('english_unaccent', COALESCE(title, '')), 'A') ||
  setweight(to_tsvector('english_unaccent', COALESCE(message, '')), 'B');


-- Create a global search function that searches across all tables
CREATE OR REPLACE FUNCTION global_search(
  search_query text,
  user_org_id uuid DEFAULT NULL,
  result_limit integer DEFAULT 50
) RETURNS TABLE (
  entity_type text,
  entity_id uuid,
  title text,
  description text,
  rank real,
  created_at timestamptz
) AS $$
BEGIN
  RETURN QUERY

  -- Search users
  SELECT
    'user'::text as entity_type,
    u.id as entity_id,
    u.name as title,
    u.email as description,
    ts_rank(u.search_vector, websearch_to_tsquery('english_unaccent', search_query)) as rank,
    u.created_at
  FROM users u
  WHERE u.search_vector @@ websearch_to_tsquery('english_unaccent', search_query)

  UNION ALL

  -- Search organizations (only if user has access)
  SELECT
    'organization'::text as entity_type,
    o.id as entity_id,
    o.name as title,
    COALESCE(o.description, '') as description,
    ts_rank(o.search_vector, websearch_to_tsquery('english_unaccent', search_query)) as rank,
    o.created_at
  FROM organizations o
  LEFT JOIN organization_members om ON o.id = om.organization_id
  WHERE o.search_vector @@ websearch_to_tsquery('english_unaccent', search_query)
    AND (user_org_id IS NULL OR om.user_id = (SELECT id FROM users WHERE organization_id = user_org_id LIMIT 1))

  UNION ALL

  -- Search notifications (only user's own notifications)
  SELECT
    'notification'::text as entity_type,
    n.id as entity_id,
    n.title as title,
    n.message as description,
    ts_rank(n.search_vector, websearch_to_tsquery('english_unaccent', search_query)) as rank,
    n.created_at
  FROM notifications n
  WHERE n.search_vector @@ websearch_to_tsquery('english_unaccent', search_query)
    AND (user_org_id IS NULL OR n.user_id = (SELECT id FROM users WHERE organization_id = user_org_id LIMIT 1))

  ORDER BY rank DESC, created_at DESC
  LIMIT result_limit;
END;
$$ LANGUAGE plpgsql;

-- Create indexes to support the global search function
CREATE INDEX IF NOT EXISTS users_created_at_idx ON users (created_at DESC);
CREATE INDEX IF NOT EXISTS organizations_created_at_idx ON organizations (created_at DESC);
CREATE INDEX IF NOT EXISTS notifications_created_at_idx ON notifications (created_at DESC);

-- Add helpful search statistics view
CREATE OR REPLACE VIEW search_statistics AS
SELECT
  'users' as table_name,
  COUNT(*) as total_records,
  COUNT(*) FILTER (WHERE search_vector IS NOT NULL) as indexed_records
FROM users
UNION ALL
SELECT
  'organizations' as table_name,
  COUNT(*) as total_records,
  COUNT(*) FILTER (WHERE search_vector IS NOT NULL) as indexed_records
FROM organizations
UNION ALL
SELECT
  'notifications' as table_name,
  COUNT(*) as total_records,
  COUNT(*) FILTER (WHERE search_vector IS NOT NULL) as indexed_records
FROM notifications;
