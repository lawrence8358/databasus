package launcher

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"databasus-backend/internal/launcher/externaldb"
	"databasus-backend/internal/launcher/metadatadb"
)

// metadataSource is the metadata database the backend uses: the bundled PostgreSQL the launcher
// manages, or an external server the launcher only hands to the backend.
type metadataSource struct {
	dsn      string
	bundled  *metadatadb.Database
	exited   <-chan error
	stopFunc func() error
}

func (s *metadataSource) getDsn() string {
	return s.dsn
}

func (s *metadataSource) stop() error {
	return s.stopFunc()
}

func startMetadataDatabase(
	ctx context.Context,
	config Config,
	logger *slog.Logger,
	childOutput io.Writer,
) (*metadataSource, error) {
	if config.ExternalMetadataDsnFile != "" {
		return useExternalMetadataDatabase(ctx, config.ExternalMetadataDsnFile, logger)
	}

	metadataBinFolder, err := findMetadataBinFolder()
	if err != nil {
		return nil, err
	}

	bundled, err := metadatadb.Start(ctx, metadatadb.Settings{
		DataFolder:   config.DataFolder,
		BinFolder:    metadataBinFolder,
		Port:         config.MetadataDbPort,
		Logger:       logger,
		ServerOutput: childOutput,
	})
	if err != nil {
		return nil, fmt.Errorf("start the metadata database: %w", err)
	}

	return &metadataSource{
		dsn:      bundled.GetDsn(),
		bundled:  bundled,
		exited:   bundled.GetUnexpectedExitChannel(),
		stopFunc: bundled.Stop,
	}, nil
}

// The connection string is read from its restricted file and is never logged. An external server
// that is down is a startup error, so the service log and event log say why the service stopped.
func useExternalMetadataDatabase(ctx context.Context, dsnFile string, logger *slog.Logger) (*metadataSource, error) {
	dsn, err := externaldb.Load(dsnFile)
	if err != nil {
		return nil, err
	}

	if err := externaldb.VerifyConnection(ctx, dsn); err != nil {
		return nil, err
	}

	logger.Info("using the external metadata database; the bundled PostgreSQL is not started")

	return &metadataSource{
		dsn:      dsn,
		exited:   nil, // never fires: the launcher does not own the server
		stopFunc: func() error { return nil },
	}, nil
}
