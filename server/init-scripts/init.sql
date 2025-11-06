-- Initial database setup script
-- This script runs automatically when the database is first created

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Quick Poll System Tables

-- Polls table
CREATE TABLE IF NOT EXISTS polls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    question TEXT NOT NULL CHECK (
        length(question) >= 5
        AND length(question) <= 500
    ),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    total_votes BIGINT DEFAULT 0
);

-- Poll options table
CREATE TABLE IF NOT EXISTS poll_options (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    poll_id UUID NOT NULL REFERENCES polls (id) ON DELETE CASCADE,
    option_text TEXT NOT NULL CHECK (
        length(option_text) >= 1
        AND length(option_text) <= 200
    ),
    vote_count BIGINT DEFAULT 0,
    position INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_poll_position UNIQUE (poll_id, position)
);

-- Votes table (tracks individual votes to prevent duplicate voting)
CREATE TABLE IF NOT EXISTS votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    poll_id UUID NOT NULL REFERENCES polls (id) ON DELETE CASCADE,
    option_id UUID NOT NULL REFERENCES poll_options (id) ON DELETE CASCADE,
    voter_identifier VARCHAR(255) NOT NULL, -- Could be IP, session ID, or user ID
    voted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_voter_per_poll UNIQUE (poll_id, voter_identifier)
);

-- Indexes for performance
CREATE INDEX idx_polls_created_at ON polls (created_at DESC);

CREATE INDEX idx_polls_active ON polls (is_active, expires_at)
WHERE
    is_active = true;

CREATE INDEX idx_poll_options_poll_id ON poll_options (poll_id, position);

CREATE INDEX idx_votes_poll_id ON votes (poll_id);

CREATE INDEX idx_votes_option_id ON votes (option_id);

CREATE INDEX idx_votes_voter ON votes (poll_id, voter_identifier);

-- Function to update poll total votes (trigger)
CREATE OR REPLACE FUNCTION update_poll_total_votes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE polls SET total_votes = total_votes + 1 WHERE id = NEW.poll_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE polls SET total_votes = total_votes - 1 WHERE id = OLD.poll_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to maintain total vote count
CREATE TRIGGER trigger_update_poll_votes
AFTER INSERT OR DELETE ON votes
FOR EACH ROW
EXECUTE FUNCTION update_poll_total_votes();