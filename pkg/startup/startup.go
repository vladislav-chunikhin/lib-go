package startup

import (
	"fmt"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
)

func SetMaxGoProcs(logger logger.Logger) error {
	// GOMAXPROCS
	// Want to see what maxprocs reports.
	opt := maxprocs.Logger(logger.Infof)
	// Set the correct number of threads for the service
	// based on what is available either by the machine or quotas.
	if _, err := maxprocs.Set(opt); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	return nil
}
