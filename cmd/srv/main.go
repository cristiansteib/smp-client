package main

import (
	"context"
	"flag"
	"gama-client/internal"
	"gama-client/internal/appconfig"
	"gama-client/internal/checks/adminprivileges"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func setupLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetOutput(os.Stdout)

	if runtime.GOOS == "windows" {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{}) // JSON para Linux o entornos estructurados
	}

	// Configura el nivel de log
	logger.SetLevel(logrus.InfoLevel)

	return logger
}

type AppExecConfig struct {
	configPath string
}

func flagsAndConfigs() *AppExecConfig {
	defaultConfigPath := "config.json"
	configPathFlag := flag.String("config", "", "Path to the configuration file")
	flag.Parse()
	configPathEnv := os.Getenv("CONFIG_PATH")
	configPath := defaultConfigPath
	if configPathEnv != "" {
		configPath = configPathEnv
	}
	if *configPathFlag != "" {
		configPath = *configPathFlag
	}

	return &AppExecConfig{configPath: configPath}
}

func assertInitialChecks(logger *logrus.Logger) {
	isAdmin, err := adminprivileges.NewAdminPrivileges().Check()
	if err != nil {
		logger.Fatalf("error while asserting initial checks: %v", err)
	}
	if !isAdmin {
		logger.Fatal("Application must run as an admin user.")
	}
	switch runtime.GOOS {
	case "windows":
		logger.Info("Running on Windows")
	case "linux":
		logger.Info("Running on Linux")
	default:
		logger.Fatalf("Os not supported.")
	}

}

func main() {
	logger := setupLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	assertInitialChecks(logger)

	appExecConfig := flagsAndConfigs()
	config, err := appconfig.LoadConfig(appExecConfig.configPath)
	if err != nil {
		logrus.Fatalf("Wrong configuration, '%s', %s", appExecConfig.configPath, err)
	}

	logger.Info("Starting service")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logger.Debugf("Signal received, stopping service [%s].", sig)
		cancel()
	}()

	go internal.Service(ctx, cancel, config)

	<-ctx.Done()
	logger.Info("Service stopped.")
}
