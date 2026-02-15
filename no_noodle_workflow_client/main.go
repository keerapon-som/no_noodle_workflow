package main

import (
	"context"
	"net/http"
	"no-noodle-workflow-client/api"
	"no-noodle-workflow-client/service"
)

func main() {

	// config := config.GetConfig()

	noNoodleClient := api.NewNoNoodleWorkflowClient("http://localhost:8888", &http.Client{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := service.New(noNoodleClient).Run(ctx)

	if err != nil {
		panic(err)
	}

}
