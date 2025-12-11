
-- Drop existing tables if they exist (Clean Start)
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS commands;
DROP TABLE IF EXISTS bot_state;

-- 1. Bot State Table (Singleton)
CREATE TABLE bot_state (
    id INT PRIMARY KEY DEFAULT 1,
    is_running BOOLEAN DEFAULT FALSE,
    last_heartbeat TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    balance NUMERIC,
    active_positions_count INT,
    total_pnl NUMERIC,
    positions_json JSONB DEFAULT '{}',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert the single row we'll use
INSERT INTO bot_state (id, is_running) VALUES (1, FALSE) ON CONFLICT DO NOTHING;

-- 2. Commands Table (Remote Control)
CREATE TABLE commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    command TEXT NOT NULL,          -- 'START', 'STOP', 'Sync'
    params JSONB DEFAULT '{}',
    status TEXT DEFAULT 'PENDING',  -- 'PENDING', 'EXECUTED', 'FAILED'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    executed_at TIMESTAMP WITH TIME ZONE
);

-- 3. Logs Table (Remote Monitoring)
CREATE TABLE logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level TEXT,
    message TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable Realtime for these tables
alter publication supabase_realtime add table bot_state;
alter publication supabase_realtime add table commands;
alter publication supabase_realtime add table logs;
