package main

import (
	"log"

	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/logging"
	"go.uber.org/zap"
)

func main() {
	env, err := env.Parse()
	if err != nil {
		log.Fatal(err)
	}
	zap.ReplaceGlobals(logging.CreateLogger(env.LogLevel))
}
