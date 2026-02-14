package entitites

import "time"

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
