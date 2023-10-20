package main

import (
	"github.com/ilyakaznacheev/cleanenv"

	"github.com/taraslis453/shopify-customer-auth/config"
	"github.com/taraslis453/shopify-customer-auth/pkg/logging"

	"github.com/taraslis453/shopify-customer-auth/internal/app"
)

func main() {
	logger := logging.NewZapLogger("main")

	var cfg config.Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		logger.Fatal("failed to read env", "err", err)
	}
	logger.Info("read config", "config", cfg)

	app.Run(&cfg)
}
