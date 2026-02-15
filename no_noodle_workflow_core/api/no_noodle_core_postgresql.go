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
	httpClient *http.Client
	repo       *repository.PostgreSQLNoNoodleWorkflow
	pubsub     *RedisMessageService
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
	FailedTask(workflowID string, task string) error
	SubscribeTask(processID string, task string, healthCheckURL string, callbackURL string) (string, error)
	SubscriberHealthCheck(callbackURL string) error
}

func NewNoNoodleWorkflowCorePostgresql(repo *repository.PostgreSQLNoNoodleWorkflow, pubsub *RedisMessageService) NoNoodleCoreInterface {

	noNoodleCore := &NoNoodleWorkflowCorePostgresql{
		httpClient: &http.Client{},
		repo:       repo,
		pubsub:     pubsub,
	}

	err := noNoodleCore.ReSubscribeTask()
	if err != nil {
		// Handle the error appropriately, e.g., log it or return it
		fmt.Println("Error re-subscribing tasks:", err)
	}

	return noNoodleCore
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

func (c *NoNoodleWorkflowCorePostgresql) FailedTask(workflowID string, task string) error {
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

	err = c.repo.UpdateTaskStatus(tx, workflowID, task, TASK_STATUS_FAILED, util.GetCurrentTime())
	if err != nil {
		return err
	}

	return nil
}

func (c *NoNoodleWorkflowCorePostgresql) publishTaskToBroker(tx *sql.Tx, processID string, workflowID string, stageTask string) error {

	type PublishedPayload struct {
		ProcessID  string `json:"process_id"`
		TaskID     string `json:"task_id"`
		WorkflowID string `json:"workflow_id"`
	}

	payload := PublishedPayload{
		ProcessID:  processID,
		TaskID:     stageTask,
		WorkflowID: workflowID,
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

func (c *NoNoodleWorkflowCorePostgresql) subscribeChannel(sessionKey string, task string, processID string, healthCheckURL string, callbackURL string) (string, error) {

	ctx, cancel := context.WithCancel(context.Background())
	// Periodic health check, stops on context cancellation or failure
	go func() {
		healthCheckMaxFailures := 10
		healthCheckFailures := 0
		healthCheckInterval := 5 * time.Second
		ticker := time.NewTicker(healthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Stopping health checks for subscriber with session key %s: %v\n", sessionKey, ctx.Err())
				return
			case <-ticker.C:
				go func() {
					err := c.websocketHealthCheck(healthCheckURL)
					if err != nil {
						healthCheckFailures++
						fmt.Println("Healtch check failed for callback URL:", callbackURL, " error:", err, " failure count:", healthCheckFailures)
						if healthCheckFailures >= healthCheckMaxFailures {
							fmt.Printf("Health check failed at max retries %d for callback URL %s: %v\n", healthCheckMaxFailures, callbackURL, err)
							if removeErr := c.repo.RemoveSubscriber(sessionKey); removeErr != nil {
								fmt.Printf("Failed to remove subscriber with session key %s: %v\n", sessionKey, removeErr)
							}
							cancel()
							return
						}
					} else {
						healthCheckFailures = 0 // Reset failure count on successful health check
					}
				}()
				go func() {
					subscriber, err := c.repo.GetSubscriberBySessionKey(sessionKey)
					if subscriber == nil {
						fmt.Printf("Subscriber with session key %s not found: %v\n", sessionKey, "Do Eject Subscriber")
						cancel()
						return
					}
					if err != nil {
						fmt.Printf("Error retrieving subscriber with session key %s: %v\n", sessionKey, err)
						cancel()
						return
					}
				}()
			}
		}
	}()

	channel := "no_noodle_workflow:" + processID + ":" + task

	go c.pubsub.SubscribeChannal(ctx, callbackURL, channel, c.websocketNotify)

	return sessionKey, nil
}

func (c *NoNoodleWorkflowCorePostgresql) SubscribeTask(processID string, task string, healthCheckURL string, callbackURL string) (string, error) {

	// channal := "no_noodle_workflow:" + processID + ":" + task

	sessionKey := generateSessionKey()
	if sessionKey == "" {
		return "", fmt.Errorf("failed to generate session key")
	}

	// Register subscriber before starting background goroutines
	if err := c.repo.SaveSubscriber(sessionKey, healthCheckURL, task, processID, callbackURL); err != nil {
		return "", err
	}

	return c.subscribeChannel(sessionKey, task, processID, healthCheckURL, callbackURL)
}

func (c *NoNoodleWorkflowCorePostgresql) websocketNotify(callbackURL string, payload []byte) error {

	req, err := http.NewRequest("POST", callbackURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

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

func (c *NoNoodleWorkflowCorePostgresql) ReSubscribeTask() error {

	allSubscribers, err := c.repo.GetAllSubscribers()
	if err != nil {
		return fmt.Errorf("failed to get all channel infos: %v", err)
	}

	// Process all live channel infos
	for _, subscriber := range *allSubscribers {
		// Implement your logic to re-subscribe tasks based on channelInfo
		_, err := c.subscribeChannel(
			subscriber.SessionKey,
			subscriber.Task,
			subscriber.ProcessID,
			subscriber.HealthCheckURL,
			subscriber.CallbackURL,
		)
		if err != nil {
			fmt.Printf("Failed to re-subscribe to task %s with session key %s: %v\n", subscriber.Task, subscriber.SessionKey, err)
			panic(err)
		}
	}

	return nil
}
