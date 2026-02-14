-- Table 1: process_config
-- Stores process configuration with stage-to-task mappings
CREATE TABLE process_config (
    process_id VARCHAR(255) PRIMARY KEY,
    map_stage_task JSONB NOT NULL,
    map_stage_ready JSONB NOT NULL,
    create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table 2: workflow
-- Stores current workflow status and task tracking
CREATE TABLE workflow (
    workflow_id VARCHAR(255) PRIMARY KEY,
    process_id VARCHAR(255) NOT NULL,
    task_status JSONB NOT NULL,
    published_stage JSONB NOT NULL,
    create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (process_id) REFERENCES process_config(process_id)
);
