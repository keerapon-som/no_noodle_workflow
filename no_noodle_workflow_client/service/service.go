package service

import (
	"context"
	"errors"
	"net/http"
	"no-noodle-workflow-client/packages/api"

	"sync"

	httpCatchup "no-noodle-workflow-client/http"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	fiberApp       *fiber.App
	noNoodleClient api.NoNoodleClientInterface
}

func New(noNoodleClient api.NoNoodleClientInterface) *Service {

	return &Service{
		fiberApp:       httpCatchup.NewHTTPRouter(noNoodleClient),
		noNoodleClient: noNoodleClient,
	}

}

func (s *Service) Run(ctx context.Context) error {

	errgroup, ctx := errgroup.WithContext(ctx)

	wg := sync.WaitGroup{}

	wg.Wait()

	errgroup.Go(func() error {

		err := s.fiberApp.Listen(":" + "1234")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {

			return err
		}

		return nil
	})

	errgroup.Go(func() error {
		<-ctx.Done()

		return s.fiberApp.Shutdown()
	})

	err := s.noNoodleClient.Run()
	if err != nil {
		return err
	}

	return errgroup.Wait()
}
