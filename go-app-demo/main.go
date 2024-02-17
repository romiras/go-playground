package main

import (
	"github.com/romiras/go-app-demo/initializers"
	"github.com/romiras/go-app-demo/internal/logger"
	"go.uber.org/fx"
)

type AppRegistry struct {
	Logger logger.Logger
}

func main() {
	// https://uber-go.github.io/fx/get-started/
	fx.New(
		fx.Provide(initializers.Dependencies()...),
		// fx.Provide(createHttpServer),
		// fx.Invoke(runHttp),
		// fx.Invoke(runSqsConsumer),
		// fx.Invoke(runKafkaConsumer),
	).Run()
}

// reg := &Registry{Logger: logger}
