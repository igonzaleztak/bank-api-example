package bootstrap

import (
	"bank_test/internal/conf"
	"bank_test/internal/db"
	"bank_test/internal/helpers"
	"bank_test/internal/transport"
	"log"
)

func Run() error {
	// Setup the configuration
	if err := conf.SetupConfig(); err != nil {
		return err
	}

	// Setup the logger
	logger, err := NewZapLogger()
	if err != nil {
		return err
	}

	logger.Info("starting the bank API")
	logger.Debugf("starting with config: %s", helpers.PrettyPrintStructResponse(conf.GlobalConfig))

	logger.Debugf("setting up database connection")
	db := db.NewDatabaseAdapter(logger)
	logger.Debugf("database connection established")

	// Setup the transport layer and start the server
	server := transport.NewTransporter(logger, db)

	go func() {
		if err := server.HealthCheck(); err != nil {
			logger.Error(err)
			log.Fatal(err)
		}
	}()

	if err := server.Serve(); err != nil {
		return err
	}

	return nil
}
