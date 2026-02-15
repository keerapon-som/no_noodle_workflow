package http

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/keerapon-som/no_noodle_workflow/packages/nonoodleclient"
)

var (
	buildtime, buildcommit, version string
)

type Handler struct {
	noNoodleCore nonoodleclient.NoNoodleClientInterface
}

func NewHTTPRouter(noNoodleCore nonoodleclient.NoNoodleClientInterface) *fiber.App {
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

	noNoodleCore.AddNoNoodleWorkflowHandler(app)

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
