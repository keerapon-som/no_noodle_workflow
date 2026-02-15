package NoNoodleClient

import "time"

type ProcessConfig struct {
	ProcessID     string              `json:"process_id"`
	MapStageTask  map[string][]string `json:"map_stage_task"`
	MapStageReady map[string][]string `json:"map_stage_ready"`
}

type TaskStatusData struct {
	Status     string    `json:"status"`
	UpdateDate time.Time `json:"update_date"`
}

type Workflow struct {
	WorkflowID     string                    `json:"workflow_id"`
	ProcessID      string                    `json:"process_id"`
	TaskStatus     map[string]TaskStatusData `json:"task_status"`
	PublishedStage map[string]bool           `json:"published_stage"`
	CreateDate     time.Time                 `json:"create_date"`
}

type Job struct {
	ProcessID  string `json:"process_id"`
	TaskID     string `json:"task_id"`
	WorkflowID string `json:"workflow_id"`
}

type JobRegistry struct {
	ProcessID string `json:"process_id"`
	TaskID    string `json:"task_id"`
}
