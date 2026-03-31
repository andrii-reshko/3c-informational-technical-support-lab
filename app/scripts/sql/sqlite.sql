DROP TABLE IF EXISTS metrics;
DROP TABLE IF EXISTS nodes;

-- nodes.sql
CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    cpu_cores REAL DEFAULT 1.0,
    ram_gb REAL DEFAULT 1.0,
    safe_boundary REAL DEFAULT 75.0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS metrics (
    node_id TEXT,
    timestamp DATETIME NOT NULL,
    cpu_usage_percent REAL,
    ram_usage_bytes INTEGER,
    PRIMARY KEY (node_id, timestamp),
    FOREIGN KEY (node_id) REFERENCES nodes(id)
);