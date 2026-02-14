package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"workflow_stage/entitites"
	"workflow_stage/repository"
	"workflow_stage/util"
)

type NoNoodleWorkflowCorePostgresql struct {
	repo   *repository.PostgreSQLNoNoodleWorkflow
	pubsub *RedisMessageService
}

const (
	TASK_STATUS_WAITING   = "waiting"
	TASK_STATUS_IN_ACTIVE = "active"
	TASK_STATUS_COMPLETED = "completed"
	TASK_STATUS_FAILED    = "failed"
)

type NoNoodleCoreInterface interface {
	DeployProcessConfig(processConfig *entitites.ProcessConfig) error
	CompleteTask(workflowID string, task string) error
	CreateWorkflow(processID string) (string, error)
	FailedTask(workflowID string, taskName string) error
}

func NewNoNoodleWorkflowCorePostgresql(repo *repository.PostgreSQLNoNoodleWorkflow, pubsub *RedisMessageService) NoNoodleCoreInterface {
	return &NoNoodleWorkflowCorePostgresql{
		repo:   repo,
		pubsub: pubsub,
	}
}

func (c *NoNoodleWorkflowCorePostgresql) DeployProcessConfig(processConfig *entitites.ProcessConfig) error {

	// Implement the logic to complete a task in the workflow using the repository
	tx, err := c.repo.GetDB().Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Implement the logic to deploy a process configuration using the repository
	err = c.repo.InsertProcessConfig(tx, processConfig)
	if err != nil {
		return err
	}
	return nil
}

func (c *NoNoodleWorkflowCorePostgresql) CompleteTask(workflowID string, task string) error {
	// Implement the logic to complete a task in the workflow using the repository
	tx, err := c.repo.GetDB().Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = c.repo.UpdateTaskStatus(tx, workflowID, task, TASK_STATUS_COMPLETED, util.GetCurrentTime())
	if err != nil {
		return err
	}

	workflow, err := c.repo.GetWorkflowByWorkflowID(tx, workflowID)
	if err != nil {
		return err
	}

	processConfig, err := c.repo.GetProcessConfigByProcessID(tx, workflow.ProcessID)
	if err != nil {
		return err
	}

	stageToPublish := []string{}

	for stage, taskToValidate := range processConfig.MapStageReady {
		if workflow.PublishedStage[stage] { // ถ้าเคย publish ไปแล้ว ให้ข้ามไปเลย
			continue
		}
		needPublish := false
		for _, validTask := range taskToValidate {
			if workflow.TaskStatus[validTask].Status == "completed" {
				needPublish = true
			} else {
				needPublish = false
				break
			}
		}
		if needPublish {
			stageToPublish = append(stageToPublish, stage)
		}
	}

	for _, stage := range stageToPublish {
		for _, stageTask := range processConfig.MapStageTask[stage] {
			err = c.publishTaskToBroker(tx, workflow.ProcessID, workflowID, stageTask)
			if err != nil {
				return err
			}
		}
		// ws.TaskMemory.AddPublishedStage(workflowID, stage)
		err = c.repo.UpdatePublishedStage(tx, workflowID, stage, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *NoNoodleWorkflowCorePostgresql) CreateWorkflow(processID string) (string, error) {
	// Implement the logic to create a new workflow using the repository

	workflowID := generateWorkflowID()

	tx, err := c.repo.GetDB().Begin()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	processConfig, err := c.repo.GetProcessConfigByProcessID(tx, processID)
	if err != nil {
		return "", err
	}

	taskData := make(map[string]entitites.TaskStatusData)
	for _, tasks := range processConfig.MapStageTask {
		for _, task := range tasks {
			taskData[task] = entitites.TaskStatusData{
				Status:     TASK_STATUS_WAITING,
				UpdateDate: util.GetCurrentTime(),
			}
		}
	}

	for _, task := range processConfig.MapStageTask["start"] {
		taskData[task] = entitites.TaskStatusData{
			Status:     TASK_STATUS_IN_ACTIVE,
			UpdateDate: util.GetCurrentTime(),
		}
	}

	publishedStage := make(map[string]bool)
	for stage := range processConfig.MapStageReady {
		publishedStage[stage] = false
	}

	err = c.repo.InitializeWorkflow(tx, workflowID, processID, taskData, publishedStage)
	if err != nil {
		return "", err
	}

	return workflowID, nil
}

func (c *NoNoodleWorkflowCorePostgresql) FailedTask(workflowID string, taskName string) error {
	// Implement the logic to complete a task in the workflow using the repository
	tx, err := c.repo.GetDB().Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = c.repo.UpdateTaskStatus(tx, workflowID, taskName, TASK_STATUS_FAILED, util.GetCurrentTime())
	if err != nil {
		return err
	}

	return nil
}

func (c *NoNoodleWorkflowCorePostgresql) publishTaskToBroker(tx *sql.Tx, processID string, workflowID string, stageTask string) error {

	payload := map[string]string{
		"workflow_id": workflowID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	channal := "no_noodle_workflow:" + processID + ":" + stageTask

	fmt.Println("Publishing to channal:", channal, " payload:", string(jsonPayload))

	err = c.repo.UpdateTaskStatus(tx, workflowID, stageTask, TASK_STATUS_IN_ACTIVE, util.GetCurrentTime())
	if err != nil {
		return err
	}

	return c.pubsub.SendToMsgChannal(context.Background(), channal, jsonPayload)
}
