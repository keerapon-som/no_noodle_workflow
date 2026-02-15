package service

import (
	"context"
	"errors"
	"net/http"

	"sync"

	httpCatchup "github.com/keerapon-som/no_noodle_workflow/http"
	"github.com/keerapon-som/no_noodle_workflow/packages/nonoodleclient"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	fiberApp       *fiber.App
	noNoodleClient nonoodleclient.NoNoodleClientInterface
}

func New(noNoodleClient nonoodleclient.NoNoodleClientInterface) *Service {

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
