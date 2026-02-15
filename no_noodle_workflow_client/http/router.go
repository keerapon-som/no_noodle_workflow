package http

import (
	"fmt"
	"net/http"
	"no-noodle-workflow-client/api"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

var (
	buildtime, buildcommit, version string
)

type Handler struct {
	noNoodleCore api.NoNoodleClientInterface
}

func NewHTTPRouter(noNoodleCore api.NoNoodleClientInterface) *fiber.App {
	app := fiber.New(fiber.Config{
		Immutable: true,
	})

	app.Use(pprof.New())

	app.Get("/version", getVersion)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "success",
		})
	})

	// app.Delete("/no_noodle_workflow_client/subscribe/:session", func(c *fiber.Ctx) error { // Notify to client that session was deleted --> need to need to do reconnect
	// 	session := c.Params("session")
	// 	fmt.Printf("Session %s was deleted, client should reconnect\n", session)
	// 	return c.JSON(fiber.Map{
	// 		"status": "success",
	// 	})
	// })

	app.Post("/no_noodle_workflow_client/:task/subscribe", func(c *fiber.Ctx) error {
		task := c.Params("task")
		fmt.Println("Received subscribe request for task:", task)
		return c.JSON(fiber.Map{
			"status": "success",
		})
	})

	// h := &Handler{
	// 	noNoodleCore: noNoodleCore,
	// }

	// app.Post("/complete_task", h.CompleteTask)
	// app.Post("/create_workflow", h.CreateWorkflow)
	// app.Post("/deploy_process_config", h.DeployProcessConfig)
	// app.Post("/failed_task", h.FailedTask)
	// app.Post("/subscribe", h.SubscribeTask)

	return app

}

func getVersion(c *fiber.Ctx) error {

	versionInfo := struct {
		BuildCommit string
		BuildTime   string
		Version     string
	}{
		BuildCommit: buildcommit,
		BuildTime:   buildtime,
		Version:     version,
	}

	return c.Status(http.StatusOK).JSON(versionInfo)
}
