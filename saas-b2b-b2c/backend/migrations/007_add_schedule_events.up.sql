-- Migration: 007_add_schedule_events
-- Creates schedule_events table for calendar/schedule functionality

CREATE TABLE IF NOT EXISTS schedule_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    salon_id UUID NOT NULL,
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) DEFAULT 'task',
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    priority VARCHAR(50) DEFAULT 'normal',
    status VARCHAR(50) DEFAULT 'planned',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_schedule_events_salon ON schedule_events(salon_id);
CREATE INDEX IF NOT EXISTS idx_schedule_events_user ON schedule_events(user_id);
CREATE INDEX IF NOT EXISTS idx_schedule_events_status ON schedule_events(status);
CREATE INDEX IF NOT EXISTS idx_schedule_events_start ON schedule_events(start_time);