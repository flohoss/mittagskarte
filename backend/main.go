package main

import (
	"log"
	"net/http"

	"github.com/flohoss/mittagskarte/internal/mittag"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.NewWithConfig(
		pocketbase.Config{
			DefaultDataDir: "/app/data/pb",
		},
	)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/health", func(re *core.RequestEvent) error {
			return re.String(http.StatusOK, ".")
		})

		return se.Next()
	})

	var mittagService *mittag.Mittag
	var err error

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		mittagService, err = mittag.New(e.App)
		if err != nil {
			return err
		}

		return nil
	})

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		if mittagService != nil {
			mittagService.Close()
		}

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
