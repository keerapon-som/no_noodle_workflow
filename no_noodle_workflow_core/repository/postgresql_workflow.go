package repository

import (
	"database/sql"
	"encoding/json"
	"time"
	"workflow_stage/entitites"
	"workflow_stage/util"
)

// func (p *PostgreSQLNoNoodleWorkflow) GetTaskStatusFromWorkflowID(tx *sql.Tx, workflowID string) (map[string]entitites.TaskStatusData, error) {
// 	var workflow entitites.Workflow
// 	var taskStatusJSON []byte

// 	err := tx.QueryRow("SELECT task_status FROM workflow WHERE workflow_id = $1", workflowID).Scan(&taskStatusJSON)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = json.Unmarshal(taskStatusJSON, &workflow.TaskStatus)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return workflow.TaskStatus, nil
// }

func (p *PostgreSQLNoNoodleWorkflow) GetWorkflowByWorkflowID(tx *sql.Tx, workflowID string) (*entitites.Workflow, error) {
	var workflow entitites.Workflow
	var taskStatusJSON []byte
	var publishedStageJSON []byte

	err := tx.QueryRow("SELECT workflow_id, process_id, task_status, published_stage, create_date FROM workflow WHERE workflow_id = $1", workflowID).Scan(&workflow.WorkflowID, &workflow.ProcessID, &taskStatusJSON, &publishedStageJSON, &workflow.CreateDate)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(taskStatusJSON, &workflow.TaskStatus)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(publishedStageJSON, &workflow.PublishedStage)
	if err != nil {
		return nil, err
	}

	return &workflow, nil
}

func (p *PostgreSQLNoNoodleWorkflow) InitializeWorkflow(tx *sql.Tx, workflowID string, processID string, taskStatus map[string]entitites.TaskStatusData, publishedStage map[string]bool) error {

	taskStatusBytes, err := json.Marshal(taskStatus)
	if err != nil {
		return err
	}

	publishedStageBytes, err := json.Marshal(publishedStage)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO workflow (workflow_id, process_id, task_status, published_stage, create_date) VALUES ($1, $2, $3, $4, $5)", workflowID, processID, taskStatusBytes, publishedStageBytes, util.GetCurrentTime())
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgreSQLNoNoodleWorkflow) UpdateTaskStatus(tx *sql.Tx, workflowID string, task string, status string, updateDate time.Time) error {
	// Use PostgreSQL JSONB operators to update specific keys directly
	query := `
		UPDATE workflow 
		SET task_status = jsonb_set(
			jsonb_set(
				task_status, 
				ARRAY[$1, 'status'], 
				to_jsonb($2::text)
			),
			ARRAY[$1, 'update_date'], 
			to_jsonb($3::text)
		)
		WHERE workflow_id = $4
	`

	_, err := tx.Exec(query, task, status, updateDate.Format(time.RFC3339Nano), workflowID)
	return err
}

func (p *PostgreSQLNoNoodleWorkflow) UpdatePublishedStage(tx *sql.Tx, workflowID string, stage string, isPublished bool) error {
	query := `
		UPDATE workflow
		SET published_stage = jsonb_set(
			published_stage,
			ARRAY[$1],
			to_jsonb($2::boolean)
		)
		WHERE workflow_id = $3
	`
	_, err := tx.Exec(query, stage, isPublished, workflowID)
	return err
}
