package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/keerapon-som/no_noodle_workflow/packages/api"
	"github.com/keerapon-som/no_noodle_workflow/packages/entitites"
	"github.com/keerapon-som/no_noodle_workflow/service"
)

func handler(noodleJobClient api.NoodleJobClient, job entitites.Job) error {

	fmt.Println("Hello from task handler, job details:", job)

	err := noodleJobClient.CompleteTask(job.WorkflowID, job.TaskID)
	if err != nil {
		fmt.Println("Error completing task:", err)
	}

	return nil
}

func main() {

	// config := config.GetConfig()

	noNoodleClient := api.NewNoNoodleWorkflowClient("http://localhost:8888", &http.Client{}, "http://localhost:1234/health", "http://localhost:1234")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	noNoodleClient.RegisterTask("process1", "task0", handler)
	noNoodleClient.RegisterTask("process1", "task1", handler)
	noNoodleClient.RegisterTask("process1", "task2", handler)
	noNoodleClient.RegisterTask("process1", "task3", handler)
	noNoodleClient.RegisterTask("process1", "task4", handler)
	noNoodleClient.RegisterTask("process1", "task5", handler)
	noNoodleClient.RegisterTask("process1", "task6", handler)

	err := service.New(noNoodleClient).Run(ctx)

	if err != nil {
		panic(err)
	}

}
