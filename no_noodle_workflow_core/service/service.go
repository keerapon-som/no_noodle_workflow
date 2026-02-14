package service

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"workflow_stage/api"
	"workflow_stage/config"
	httpCatchup "workflow_stage/http"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	fiberApp *fiber.App
}

func New(noNoodleCore api.NoNoodleCoreInterface) *Service {

	return &Service{
		fiberApp: httpCatchup.NewHTTPRouter(noNoodleCore),
	}

}

func (s *Service) Run(ctx context.Context) error {

	errgroup, ctx := errgroup.WithContext(ctx)

	wg := sync.WaitGroup{}

	wg.Wait()

	errgroup.Go(func() error {

		err := s.fiberApp.Listen(":" + config.GetConfig().ServerConfig.HTTP.Port)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {

			return err
		}

		return nil
	})

	errgroup.Go(func() error {
		<-ctx.Done()

		return s.fiberApp.Shutdown()
	})

	return errgroup.Wait()
}
