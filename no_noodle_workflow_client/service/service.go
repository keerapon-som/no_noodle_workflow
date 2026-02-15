package service

import (
	"context"
	"errors"
	"net/http"
	"no-noodle-workflow-client/api"
	"no-noodle-workflow-client/config"
	"sync"

	httpCatchup "no-noodle-workflow-client/http"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	fiberApp *fiber.App
}

func New(noNoodleClient api.NoNoodleClientInterface) *Service {

	return &Service{
		fiberApp: httpCatchup.NewHTTPRouter(noNoodleClient),
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
