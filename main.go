package main

import (
	"easy_pipeline_engine/api"
	"easy_pipeline_engine/config"
	"easy_pipeline_engine/repository"
	"fmt"
)

func main() {

	configTask := config.WorkflowStageConfig{
		MapStageTask: map[string][]string{
			"start":  {"task0"},
			"stage1": {"task1", "task2"},
			"stage2": {"task3", "task4"},
			"stage3": {"task5", "task6"},
		},
		MapStageReady: map[string][]string{
			"stage1": {"task0"},
			"stage2": {"task1", "task2"},
			"stage3": {"task3", "task4"},
		},
	}

	repo := repository.NewTaskMemory()
	wf := api.NewWorkflowStage(repo, configTask)
	wf.RegisterHandler("task0", func(workflowID string) error {
		fmt.Println("Handle task0")
		return nil
	})
	wf.RegisterHandler("task1", func(workflowID string) error {
		fmt.Println("Handle task1")
		return nil
	})
	wf.RegisterHandler("task2", func(workflowID string) error {
		fmt.Println("Handle task2")
		return nil
	})
	wf.RegisterHandler("task3", func(workflowID string) error {
		fmt.Println("Handle task3")
		return nil
	})
	wf.RegisterHandler("task4", func(workflowID string) error {
		fmt.Println("Handle task4")
		return nil
	})
	wf.RegisterHandler("task5", func(workflowID string) error {
		fmt.Println("Handle task5")
		return nil
	})
	wf.RegisterHandler("task6", func(workflowID string) error {
		fmt.Println("Handle task6")
		return nil
	})

	workflowID, err := wf.CreateProcessInsatnce()
	if err != nil {
		fmt.Println("Error creating process instance:", err)
		return
	}
	fmt.Println("Workflow ID:", workflowID)

}
