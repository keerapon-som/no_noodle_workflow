package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"no-noodle-workflow-client/entitites"
)

type NoNoodleWorkflowClient struct {
	hosturl    string
	httpClient *http.Client
}

type NoNoodleClientInterface interface {
	DeployProcessConfig(processConfig *entitites.ProcessConfig) error
	CompleteTask(workflowID string, task string) error
	CreateWorkflow(processID string) (string, error)
	FailedTask(workflowID string, task string) error
	SubscribeTask(processID string, task string, healthCheckURL string, callbackURL string, expiration int64) (string, error)
}

func NewNoNoodleWorkflowClient(hosturl string, httpClient *http.Client) NoNoodleClientInterface {
	return &NoNoodleWorkflowClient{
		hosturl:    hosturl,
		httpClient: httpClient,
	}
}

func (c *NoNoodleWorkflowClient) DeployProcessConfig(processConfig *entitites.ProcessConfig) error {
	// Implement the logic to deploy a process configuration using HTTP API or Redis message queue
	return nil
}

func (c *NoNoodleWorkflowClient) CompleteTask(workflowID string, task string) error {

	url := c.hosturl + "/complete_task"

	payload := map[string]string{
		"workflow_id": workflowID,
		"task":        task,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {

		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to complete task, status code: %d, response: %s", res.StatusCode, string(body))
	}

	var createWorkflowResp CreateWorkflowResponse
	err = json.Unmarshal(body, &createWorkflowResp)
	if err != nil {
		return err
	}

	return nil
}

type CreateWorkflowResponse struct {
	WorkflowID string `json:"workflow_id"`
}

func (c *NoNoodleWorkflowClient) CreateWorkflow(processID string) (string, error) {

	url := c.hosturl + "/create_workflow"

	payload := map[string]string{
		"process_id": processID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {

		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create workflow, status code: %d, response: %s", res.StatusCode, string(body))
	}

	var createWorkflowResp CreateWorkflowResponse
	err = json.Unmarshal(body, &createWorkflowResp)
	if err != nil {
		return "", err
	}

	return createWorkflowResp.WorkflowID, nil
}

func (c *NoNoodleWorkflowClient) FailedTask(workflowID string, task string) error {
	// Implement the logic to mark a task as failed in the workflow using HTTP API or Redis message queue
	return nil
}

func (c *NoNoodleWorkflowClient) SubscribeTask(processID string, task string, healthCheckURL string, callbackURL string, expiration int64) (string, error) {

	type SubscribeRequest struct {
		ProcessID      string `json:"process_id"`
		Task           string `json:"task"`
		HealthCheckURL string `json:"health_check_url"`
		CallbackURL    string `json:"callback_url"`
		Expiration     int64  `json:"expiration"`
	}

	payload := SubscribeRequest{
		ProcessID:      processID,
		Task:           task,
		HealthCheckURL: healthCheckURL,
		CallbackURL:    callbackURL,
		Expiration:     expiration,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", callbackURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to notify subscriber, status code: %d", resp.StatusCode)
	}

	type SubscribeResponse struct {
		SessionKey string `json:"connection_key"`
	}

	var subscribeResp SubscribeResponse
	err = json.NewDecoder(resp.Body).Decode(&subscribeResp)

	if err != nil {
		return "", err
	}

	return subscribeResp.SessionKey, nil
}
