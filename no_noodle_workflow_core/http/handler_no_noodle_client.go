package http

import (
	"workflow_stage/entitites"

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
