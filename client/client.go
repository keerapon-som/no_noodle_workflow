package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type ProcessRegistry struct {
	listTaskRegistry map[string]func(jobRegistry Job) error
}

type NoNoodleWorkflowClient struct {
	hosturl              string
	httpClient           *http.Client
	ProcessRegistry      *ProcessRegistry
	fiberApp             *fiber.App
	clientHealthCheckUrl string
	clientBaseUrl        string
}

type NoodleJobClient struct {
	CompleteTask func(workflowID string, task string) error
	FailedTask   func(workflowID string, task string) error
}

type NoNoodleClientInterface interface {
	DeployProcessConfig(processConfig *ProcessConfig) error
	CompleteTask(workflowID string, task string) error
	CreateWorkflow(processID string) (string, error)
	FailedTask(workflowID string, task string) error
	AddNoNoodleWorkflowHandler(fiberApp *fiber.App)
	RegisterTask(processID string, task string, handler func(noodleJobClient NoodleJobClient, job Job) error)
	Run() error
}

func NewNoNoodleWorkflowClient(hosturl string, httpClient *http.Client, clientHealthCheckUrl string, clientBaseUrl string) NoNoodleClientInterface {

	return &NoNoodleWorkflowClient{
		hosturl:    hosturl,
		httpClient: httpClient,
		ProcessRegistry: &ProcessRegistry{
			listTaskRegistry: make(map[string]func(job Job) error),
		},
		clientHealthCheckUrl: clientHealthCheckUrl,
		clientBaseUrl:        clientBaseUrl,
	}
}

func (nn *NoNoodleWorkflowClient) RegisterTask(processID string, task string, handler func(noodleJobClient NoodleJobClient, job Job) error) {

	nn.ProcessRegistry.listTaskRegistry[processID+"_"+task] = func(job Job) error {
		return handler(NoodleJobClient{
			CompleteTask: nn.CompleteTask,
			FailedTask:   nn.FailedTask,
		}, job)
	}
}

func (nn *NoNoodleWorkflowClient) healthCheck() error {

	req, err := http.NewRequest("GET", nn.clientHealthCheckUrl, nil)
	if err != nil {
		return err
	}

	resp, err := nn.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status code: %d", resp.StatusCode)
	}
	return nil
}

func (pr *ProcessRegistry) getTaskHandler(processID string, task string) (func(job Job) error, bool) {
	handler, exists := pr.listTaskRegistry[processID+"_"+task]
	return handler, exists
}

func (nn *NoNoodleWorkflowClient) reSubscribeAllTasks() {

	err := nn.healthCheck()
	if err != nil {
		fmt.Printf("Health check failed, cannot subscribe to tasks, error: %v\n", err)
		return
	}

	for key := range nn.ProcessRegistry.listTaskRegistry {
		splitTaskAndProcess := bytes.Split([]byte(key), []byte("_"))
		processID := string(splitTaskAndProcess[0])
		task := string(splitTaskAndProcess[1])

		callbackUrl := fmt.Sprintf("%s/no_noodle_workflow_client/subscribe", nn.clientBaseUrl)

		_, err := nn.subscribeTask(processID, task, nn.clientHealthCheckUrl, callbackUrl)
		if err != nil {
			fmt.Printf("Error re-subscribing to task: %s of process: %s, error: %v\n", task, processID, err)
		}
	}
}

func (nn *NoNoodleWorkflowClient) Run() error {
	if len(nn.ProcessRegistry.listTaskRegistry) == 0 {
		return fmt.Errorf("no task handlers registered, please register at least one task handler before running the client")
	}

	nn.fiberApp.Post("/no_noodle_workflow_client/subscribe", func(c *fiber.Ctx) error {

		var jsonPayloads Job
		if err := c.BodyParser(&jsonPayloads); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"error":   "Invalid request body",
				"details": err.Error(),
			})
		}

		if jsonPayloads.ProcessID == "" || jsonPayloads.TaskID == "" || jsonPayloads.WorkflowID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "error",
				"error":  "Missing required fields in request body",
				"data": fiber.Map{
					"process_id":  jsonPayloads.ProcessID,
					"task_id":     jsonPayloads.TaskID,
					"workflow_id": jsonPayloads.WorkflowID,
				},
			})
		}

		go nn.taskHandler(jsonPayloads)

		return c.JSON(fiber.Map{
			"status": "success",
		})
	})

	nn.reSubscribeAllTasks()

	// Implement the logic to run the client, such as starting an HTTP server to listen for incoming task requests or connecting to a message queue
	return nil
}
func (nn *NoNoodleWorkflowClient) taskHandler(job Job) {
	handler, exists := nn.ProcessRegistry.getTaskHandler(job.ProcessID, job.TaskID)
	if exists {
		handler(job)
	}
}

// func (nn *NoNoodleWorkflowClient) HandleTask()

func (nn *NoNoodleWorkflowClient) AddNoNoodleWorkflowHandler(fiberApp *fiber.App) {
	nn.fiberApp = fiberApp
}

func (nn *NoNoodleWorkflowClient) DeployProcessConfig(processConfig *ProcessConfig) error {
	// Implement the logic to deploy a process configuration using HTTP API or Redis message queue
	return nil
}

func (nn *NoNoodleWorkflowClient) CompleteTask(workflowID string, task string) error {

	url := nn.hosturl + "/complete_task"

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

	res, err := nn.httpClient.Do(req)
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

func (nn *NoNoodleWorkflowClient) CreateWorkflow(processID string) (string, error) {

	url := nn.hosturl + "/create_workflow"

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

	res, err := nn.httpClient.Do(req)
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

func (nn *NoNoodleWorkflowClient) FailedTask(workflowID string, task string) error {
	// Implement the logic to mark a task as failed in the workflow using HTTP API or Redis message queue
	return nil
}

func (nn *NoNoodleWorkflowClient) subscribeTask(processID string, task string, healthCheckURL string, callbackURL string) (string, error) {

	type SubscribeRequest struct {
		ProcessID      string `json:"process_id"`
		Task           string `json:"task"`
		HealthCheckURL string `json:"health_check_url"`
		CallbackURL    string `json:"callback_url"`
	}

	payload := SubscribeRequest{
		ProcessID:      processID,
		Task:           task,
		HealthCheckURL: healthCheckURL,
		CallbackURL:    callbackURL,
	}

	fmt.Println("Subscribing to task with payload:", payload)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", nn.hosturl+"/subscribe", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := nn.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to subscribe process %s task %s, status code: %d", processID, task, resp.StatusCode)
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
