-- Add performance-enhancing indexes for common queries

-- Users table indexes
CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));
CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users (organization_id) WHERE organization_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users (email_verified) WHERE email_verified = true;
CREATE INDEX IF NOT EXISTS idx_users_created_at_desc ON users (created_at DESC);

-- Sessions table indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions (refresh_token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at) WHERE expires_at > NOW();
CREATE INDEX IF NOT EXISTS idx_sessions_user_expires ON sessions (user_id, expires_at DESC);

-- Organizations table indexes
CREATE INDEX IF NOT EXISTS idx_organizations_owner_id ON organizations (owner_id);
CREATE INDEX IF NOT EXISTS idx_organizations_created_at_desc ON organizations (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_organizations_name_lower ON organizations (LOWER(name));

-- Organization members table indexes
CREATE INDEX IF NOT EXISTS idx_org_members_org_id ON organization_members (organization_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON organization_members (user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_role ON organization_members (organization_id, role);
CREATE INDEX IF NOT EXISTS idx_org_members_composite ON organization_members (organization_id, user_id);

-- Files table indexes
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files (user_id);
CREATE INDEX IF NOT EXISTS idx_files_user_created ON files (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_files_storage_key ON files (storage_key);
CREATE INDEX IF NOT EXISTS idx_files_mime_type ON files (mime_type);

-- Subscriptions table indexes
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_id ON subscriptions (stripe_subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions (status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON subscriptions (user_id, status) WHERE status IN ('active', 'trialing');
CREATE INDEX IF NOT EXISTS idx_subscriptions_ends_at ON subscriptions (current_period_end) WHERE current_period_end > NOW();

-- Audit logs table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs (action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at_desc ON audit_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_composite ON audit_logs (user_id, created_at DESC) WHERE user_id IS NOT NULL;

-- Notifications table indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_read ON notifications (user_id, read);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications (user_id, created_at DESC) WHERE read = false;
CREATE INDEX IF NOT EXISTS idx_notifications_created_at_desc ON notifications (created_at DESC);

-- OAuth providers table indexes
CREATE INDEX IF NOT EXISTS idx_oauth_user_id ON oauth_providers (user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_provider_user_id ON oauth_providers (provider, provider_user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_composite ON oauth_providers (provider, user_id);

-- Add composite indexes for common join queries
CREATE INDEX IF NOT EXISTS idx_users_org_email ON users (organization_id, email) WHERE organization_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_files_user_type ON files (user_id, mime_type);

-- Add partial indexes for common filter queries
CREATE INDEX IF NOT EXISTS idx_subscriptions_active_users ON subscriptions (user_id)
    WHERE status IN ('active', 'trialing', 'past_due');

CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions (user_id, refresh_token)
    WHERE expires_at > NOW();

CREATE INDEX IF NOT EXISTS idx_users_active ON users (id, email)
    WHERE email_verified = true;

-- Statistics and monitoring views
CREATE OR REPLACE VIEW performance_statistics AS
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

CREATE OR REPLACE VIEW table_statistics AS
SELECT
    schemaname,
    tablename,
    seq_scan as sequential_scans,
    seq_tup_read as sequential_tuples_read,
    idx_scan as index_scans,
    idx_tup_fetch as index_tuples_fetched,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size
FROM pg_stat_user_tables
ORDER BY seq_scan DESC;

-- Add table comments for documentation
COMMENT ON INDEX idx_users_email_lower IS 'Case-insensitive email lookup for login';
COMMENT ON INDEX idx_sessions_user_expires IS 'Composite index for session cleanup and user session queries';
COMMENT ON INDEX idx_org_members_composite IS 'Composite index for organization membership checks';
COMMENT ON INDEX idx_subscriptions_active IS 'Index for active subscription queries';
COMMENT ON INDEX idx_notifications_user_unread IS 'Partial index for unread notification queries';

-- Analyze tables to update statistics
ANALYZE users;
ANALYZE sessions;
ANALYZE organizations;
ANALYZE organization_members;
ANALYZE files;
ANALYZE subscriptions;
ANALYZE notifications;
ANALYZE oauth_providers;
