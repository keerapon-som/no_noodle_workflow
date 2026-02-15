package main

import (
	"context"
	"fmt"
	"no-noodle-workflow-core/api"
	"no-noodle-workflow-core/config"
	"no-noodle-workflow-core/msgbroker"
	"no-noodle-workflow-core/repository"
	"no-noodle-workflow-core/service"
	"no-noodle-workflow-core/util"
	"time"
)

func main() {

	config := config.GetConfig()

	pgDB, err := util.NewPostgresql(
		config.PostgresqlRepoConfig.Host,
		config.PostgresqlRepoConfig.Port,
		config.PostgresqlRepoConfig.User,
		config.PostgresqlRepoConfig.Password,
		config.PostgresqlRepoConfig.Dbname,
		config.PostgresqlRepoConfig.SSLMode,
	)
	if err != nil {
		fmt.Println("Error connecting to PostgreSQL:", err)
		return
	}

	repo := repository.NewPostgreSQLNoNoodleWorkflow(pgDB)

	redisBroker, err := msgbroker.NewRedisMessageBroker(
		config.RedisMessageBrokerConfig.Addr, config.RedisMessageBrokerConfig.Password, config.RedisMessageBrokerConfig.DB,
	)
	if err != nil {
		panic(err)
	}
	defer redisBroker.Close()

	msgService := api.NewRedisMessageService(redisBroker)

	taskSubscriberRegistry, err := repository.NewRedisTaskSubscriberRegistry(
		config.RedisMessageBrokerConfig.Addr,
		config.RedisMessageBrokerConfig.Password,
		config.RedisMessageBrokerConfig.DB,
		30*time.Minute, // default expiration for subscribers
	)
	if err != nil {
		panic(err)
	}
	noNoodleCoreService := api.NewNoNoodleWorkflowCorePostgresql(repo, msgService, taskSubscriberRegistry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = service.New(noNoodleCoreService).Run(ctx)

	if err != nil {
		panic(err)
	}

	// Start subscriber first (in a goroutine so main can continue)
	// go msgService.SubscribeChannal(ctx, "no_noodle_workflow:process1:task0")

	// time.Sleep(500 * time.Millisecond) // small delay to ensure subscription

	// time.Sleep(2 * time.Second)
	// err = noNoodleCoreService.CompleteTask(workflowID, "task0")
	// if err != nil {
	// 	fmt.Println("Error completing task:", err)
	// 	return
	// }
	// time.Sleep(2 * time.Second)

	// repo := repository.NewTaskMemory()
	// wf := service.NewWorkflowStage(repo, configTask)
	// wf.RegisterHandler("task0", func(workflowID string) error {
	// 	fmt.Println("Handle task0")
	// 	return nil
	// })
	// wf.RegisterHandler("task1", func(workflowID string) error {
	// 	fmt.Println("Handle task1")
	// 	return nil
	// })
	// wf.RegisterHandler("task2", func(workflowID string) error {
	// 	fmt.Println("Handle task2")
	// 	return nil
	// })
	// wf.RegisterHandler("task3", func(workflowID string) error {
	// 	fmt.Println("Handle task3")
	// 	return nil
	// })
	// wf.RegisterHandler("task4", func(workflowID string) error {
	// 	fmt.Println("Handle task4")
	// 	return nil
	// })
	// wf.RegisterHandler("task5", func(workflowID string) error {
	// 	fmt.Println("Handle task5")
	// 	return nil
	// })
	// wf.RegisterHandler("task6", func(workflowID string) error {
	// 	fmt.Println("Handle task6")
	// 	return nil
	// })

	// workflowID, err := wf.CreateProcessInsatnce()
	// if err != nil {
	// 	fmt.Println("Error creating process instance:", err)
	// 	return
	// }
	// fmt.Println("Workflow ID:", workflowID)

}
