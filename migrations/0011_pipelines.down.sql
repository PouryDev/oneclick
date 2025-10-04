-- Migration: 0011_pipelines.down.sql
-- Description: Drop pipelines and pipeline_steps tables

-- Drop triggers first
DROP TRIGGER IF EXISTS update_pipeline_steps_updated_at ON pipeline_steps;
DROP TRIGGER IF EXISTS update_pipelines_updated_at ON pipelines;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (order matters due to foreign key constraints)
DROP TABLE IF EXISTS pipeline_steps;
DROP TABLE IF EXISTS pipelines;
