package api

import (
	"easy_pipeline_engine/config"
	"easy_pipeline_engine/repository"

	"github.com/google/uuid"
)

type WorkflowStage struct {
	TaskMemory          *repository.TaskMemory
	workflowStageConfig config.WorkflowStageConfig
	taskRegisteredFunc  map[string]func(workflowID string) error
}

func isInSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func uniqueValue(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func NewWorkflowStage(taskMemory *repository.TaskMemory, config config.WorkflowStageConfig) *WorkflowStage {
	return &WorkflowStage{
		TaskMemory:          taskMemory,
		workflowStageConfig: config,
		taskRegisteredFunc:  make(map[string]func(workflowID string) error),
	}
}

func (ws *WorkflowStage) CompleteTask(workflowID string, task string) error {
	ws.TaskMemory.UpdateTaskStatus(workflowID, task, "completed")
	tasksStatus := ws.TaskMemory.GetTasks(workflowID)
	stageToPublish := []string{}

	publishedStage := ws.TaskMemory.GetPublishedStages(workflowID)
	for stage, taskToValidate := range ws.workflowStageConfig.MapStageReady {
		if isInSlice(publishedStage, stage) {
			continue
		}
		needPublish := false
		for _, task := range taskToValidate {
			for _, taskStatus := range tasksStatus {
				if taskStatus.TaskName == task {
					if taskStatus.Status == "completed" {
						needPublish = true
					} else {
						needPublish = false
						break
					}
				}
			}
			if needPublish {
				stageToPublish = append(stageToPublish, stage)
			}
		}
	}

	for _, stage := range stageToPublish {
		ws.TaskMemory.AddPublishedStage(workflowID, stage)
		for _, task := range ws.workflowStageConfig.MapStageTask[stage] {
			ws.publishTask(workflowID, task)
		}
	}

	return nil
}

func (ws *WorkflowStage) CreateProcessInsatnce() (string, error) {
	// 1. generate workflowID
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	uuidString := uuid.String()

	listAllTasks := []string{}
	for _, tasks := range ws.workflowStageConfig.MapStageTask {
		listAllTasks = append(listAllTasks, tasks...)
	}

	uniqueTasks := uniqueValue(listAllTasks)

	for _, task := range uniqueTasks {
		ws.TaskMemory.AddTask(uuidString, repository.TaskStatusData{
			TaskName: task,
			Status:   "initial",
		})
	}

	stageTasks := ws.workflowStageConfig.MapStageTask["start"]
	ws.TaskMemory.AddPublishedStage(uuidString, "start")
	for _, task := range stageTasks {
		ws.CompleteTask(uuidString, task)
	}

	return uuidString, nil
}

func (ws *WorkflowStage) publishTask(workflowID string, task string) error {
	ws.TaskMemory.UpdateTaskStatus(workflowID, task, "waiting")
	return ws.handleTask(workflowID, task)
}

func (ws *WorkflowStage) RegisterHandler(task string, handler func(workflowID string) error) error {
	ws.taskRegisteredFunc[task] = handler
	return nil
}

func (ws *WorkflowStage) handleTask(workflowID string, task string) error {

	handler := ws.taskRegisteredFunc[task]
	err := handler(workflowID)
	if err != nil {
		return err
	}

	return ws.CompleteTask(workflowID, task)
}
