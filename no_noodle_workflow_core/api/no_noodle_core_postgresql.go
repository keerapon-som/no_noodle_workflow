package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"no-noodle-workflow-core/entitites"
	"no-noodle-workflow-core/repository"
	"no-noodle-workflow-core/util"
	"time"
)

type NoNoodleWorkflowCorePostgresql struct {
	httpClient             *http.Client
	repo                   *repository.PostgreSQLNoNoodleWorkflow
	pubsub                 *RedisMessageService
	taskSubscriberRegistry *repository.RedisTaskSubscriberRegistry
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
	SubscribeTask(processID string, taskName string, healthCheckURL string, callbackURL string, expiration int64) (string, error)
	SubscriberHealthCheck(callbackURL string) error
}

func NewNoNoodleWorkflowCorePostgresql(repo *repository.PostgreSQLNoNoodleWorkflow, pubsub *RedisMessageService, taskSubscriberRegistry *repository.RedisTaskSubscriberRegistry) NoNoodleCoreInterface {
	return &NoNoodleWorkflowCorePostgresql{
		httpClient:             &http.Client{},
		repo:                   repo,
		pubsub:                 pubsub,
		taskSubscriberRegistry: taskSubscriberRegistry,
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

		publishTaskToBrokerErr := c.publishTaskToBroker(tx, processID, workflowID, task)
		if publishTaskToBrokerErr != nil {
			return "", publishTaskToBrokerErr
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

func (c *NoNoodleWorkflowCorePostgresql) SubscriberHealthCheck(callbackURL string) error {

	err := c.websocketHealthCheck(callbackURL)
	if err != nil {
		return fmt.Errorf("subscriber health check failed: %v", err)
	}

	return nil
}

func (c *NoNoodleWorkflowCorePostgresql) SubscribeTask(processID string, taskName string, healthCheckURL string, callbackURL string, expiration int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(expiration)*time.Second)

	// Initial health check before registering subscriber
	if err := c.websocketHealthCheck(healthCheckURL); err != nil {
		cancel()
		return "", fmt.Errorf("subscriber health check failed: %v", err)
	}

	sessionKey := generateSessionKey()
	if sessionKey == "" {
		cancel()
		return "", fmt.Errorf("failed to generate session key")
	}

	channal := "no_noodle_workflow:" + processID + ":" + taskName

	// Register subscriber before starting background goroutines
	if err := c.taskSubscriberRegistry.AddSubscriber(sessionKey, channal, callbackURL, time.Duration(expiration)*time.Second); err != nil {
		cancel()
		return "", err
	}

	// Periodic health check, stops on context cancellation or failure
	go func() {
		healthCheckInterval := 10 * time.Second
		ticker := time.NewTicker(healthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Stopping health checks for subscriber with session key %s: %v\n", sessionKey, ctx.Err())
				return
			case <-ticker.C:
				go func() {
					if err := c.websocketHealthCheck(healthCheckURL); err != nil {
						fmt.Printf("Health check failed for subscriber with callback URL %s: %v\n", callbackURL, err)
						if removeErr := c.taskSubscriberRegistry.RemoveSubscriber(sessionKey); removeErr != nil {
							fmt.Printf("Failed to remove subscriber with session key %s: %v\n", sessionKey, removeErr)
						}
						cancel()
						return
					}
				}()
				go func() {
					sessionKey, err := c.taskSubscriberRegistry.GetChannelInfoBySessionKey(sessionKey)
					if err != nil {
						fmt.Printf("Notfound channelInfo for session key %s: %v\n", sessionKey, "Do Eject Subscriber")
						cancel()
						return
					}
				}()
			}
		}
	}()

	// Start consuming messages; this will stop when ctx is cancelled
	go c.pubsub.SubscribeChannal(ctx, callbackURL, channal, c.websocketNotify)

	return sessionKey, nil
}

func (c *NoNoodleWorkflowCorePostgresql) websocketNotify(callbackURL string, payload []byte) error {

	req, err := http.NewRequest("POST", callbackURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to notify subscriber, status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *NoNoodleWorkflowCorePostgresql) websocketHealthCheck(callbackURL string) error {

	req, err := http.NewRequest("GET", callbackURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("subscriber health check failed, status code: %d", resp.StatusCode)
	}

	return nil
}
