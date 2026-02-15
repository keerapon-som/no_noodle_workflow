package http

import (
	"fmt"

	"github.com/keerapon-som/no_noodle_workflow/core/entitites"

	"github.com/gofiber/fiber/v2"
)

type CompleteTaskRequest struct {
	WorkflowID string `json:"workflow_id"`
	Task       string `json:"task"`
}

func (h *Handler) CompleteTask(c *fiber.Ctx) error {

	var req CompleteTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.noNoodleCore.CompleteTask(req.WorkflowID, req.Task)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to complete task",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   "Task completed successfully",
	})
}

func (h *Handler) CreateWorkflow(c *fiber.Ctx) error {

	type CreateWorkflowRequest struct {
		ProcessID string `json:"process_id"`
	}

	var req CreateWorkflowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	workflowID, err := h.noNoodleCore.CreateWorkflow(req.ProcessID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create workflow",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   fiber.Map{"workflow_id": workflowID},
	})
}

func (h *Handler) DeployProcessConfig(c *fiber.Ctx) error {

	type DeployProcessConfigRequest struct {
		ProcessID     string              `json:"process_id"`
		MapStageTask  map[string][]string `json:"map_stage_task"`
		MapStageReady map[string][]string `json:"map_stage_ready"`
	}

	var req DeployProcessConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.noNoodleCore.DeployProcessConfig(&entitites.ProcessConfig{
		ProcessID:     req.ProcessID,
		MapStageTask:  req.MapStageTask,
		MapStageReady: req.MapStageReady,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to deploy process config",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   "Process config deployed successfully",
	})
}

func (h *Handler) FailedTask(c *fiber.Ctx) error {

	type FailedTaskRequest struct {
		WorkflowID string `json:"workflow_id"`
		Task       string `json:"task"`
	}

	var req FailedTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.noNoodleCore.FailedTask(req.WorkflowID, req.Task)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark task as failed",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   "Task marked as failed successfully",
	})
}

func (h *Handler) SubscribeTask(c *fiber.Ctx) error {

	type SubscribeRequest struct {
		ProcessID      string `json:"process_id"`
		Task           string `json:"task"`
		HealthCheckURL string `json:"health_check_url"`
		CallbackURL    string `json:"callback_url"`
	}

	var req SubscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	if req.ProcessID == "" || req.Task == "" || req.HealthCheckURL == "" || req.CallbackURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error",
			"error":  "Missing required fields: process_id, task, health_check_url, callback_url",
			"requestData": fiber.Map{
				"process_id":       req.ProcessID,
				"task":             req.Task,
				"health_check_url": req.HealthCheckURL,
				"callback_url":     req.CallbackURL,
			},
		})
	}

	fmt.Println("get Req ", req)

	err := h.noNoodleCore.SubscriberHealthCheck(req.HealthCheckURL)
	if err != nil {
		fmt.Println("GET SUBSCRIBE ERROR ", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"error":   "Subscriber health check failed",
			"details": err.Error(),
		})
	}

	sessionKey, err := h.noNoodleCore.SubscribeTask(req.ProcessID, req.Task, req.HealthCheckURL, req.CallbackURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"error":   "Failed to subscribe to task",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"connection_key": sessionKey,
	})
}
