package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"redislogger/config"
	"redislogger/proxy"
)

func main() {
	// Initialize logger with debug level
	logger, _ := zap.NewDevelopment(zap.IncreaseLevel(zap.DebugLevel))
	defer logger.Sync()

	logger.Debug("Starting Redis proxy")

	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}
	logger.Debug("Configuration loaded", 
		zap.String("listen_addr", cfg.ListenAddr),
		zap.String("redis_addr", cfg.RedisAddr),
	)

	// Create proxy
	p := proxy.New(cfg, logger)
	logger.Debug("Proxy instance created")

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	logger.Debug("Signal handlers registered")

	// Start proxy in a goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Debug("Starting proxy server")
		errChan <- p.Start(ctx)
	}()

	// Wait for either a signal or an error
	select {
	case sig := <-sigChan:
		logger.Info("Received signal", zap.String("signal", sig.String()))
		cancel()
	case err := <-errChan:
		if err != nil {
			logger.Fatal("Proxy error", zap.Error(err))
		}
	}
	logger.Debug("Shutting down")
}
