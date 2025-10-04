-- Migration: 0012_events_readmodel.down.sql
-- Description: Revert event logging and read model tables

-- Drop triggers (if they exist)
DROP TRIGGER IF EXISTS pipelines_event_trigger ON pipelines;

DROP TRIGGER IF EXISTS clusters_event_trigger ON clusters;

DROP TRIGGER IF EXISTS applications_event_trigger ON applications;

-- Drop functions
DROP FUNCTION IF EXISTS log_event ();

DROP FUNCTION IF EXISTS update_dashboard_counts (UUID);

-- Drop tables
DROP TABLE IF EXISTS dashboard_counts;

DROP TABLE IF EXISTS read_model_projects;

DROP TABLE IF EXISTS event_logs;