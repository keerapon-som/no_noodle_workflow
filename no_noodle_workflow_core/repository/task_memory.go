package repository

import "sync"

type WorkflowData struct {
	Task               []TaskStatusData
	ListPublishedStage []string
}

type TaskStatusData struct {
	TaskName   string
	Status     string
	UpdateDate string
}

type TaskMemory struct {
	mu               sync.Mutex
	workflowTaskRepo map[string]WorkflowData
}

func NewTaskMemory() *TaskMemory {
	return &TaskMemory{
		workflowTaskRepo: make(map[string]WorkflowData),
		mu:               sync.Mutex{},
	}
}

func (tm *TaskMemory) AddTask(workflowID string, taskData TaskStatusData) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	workflowData := tm.workflowTaskRepo[workflowID]
	workflowData.Task = append(workflowData.Task, taskData)
	tm.workflowTaskRepo[workflowID] = workflowData
}

func (tm *TaskMemory) GetTasks(workflowID string) []TaskStatusData {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	return tm.workflowTaskRepo[workflowID].Task
}

func (tm *TaskMemory) UpdateTaskStatus(workflowID string, taskName string, newStatus string, updateDate string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tasks := tm.workflowTaskRepo[workflowID]
	for i, task := range tasks.Task {
		if task.TaskName == taskName {
			tasks.Task[i].Status = newStatus
			tasks.Task[i].UpdateDate = updateDate
			break
		}
	}
	tm.workflowTaskRepo[workflowID] = tasks
}

func (tm *TaskMemory) AddPublishedStage(workflowID string, stageName string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	workflowData := tm.workflowTaskRepo[workflowID]
	workflowData.ListPublishedStage = append(workflowData.ListPublishedStage, stageName)
	tm.workflowTaskRepo[workflowID] = workflowData
}

func (tm *TaskMemory) GetPublishedStages(workflowID string) []string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	return tm.workflowTaskRepo[workflowID].ListPublishedStage
}

func (tm *TaskMemory) DeleteWorkflow(workflowID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.workflowTaskRepo, workflowID)
}
