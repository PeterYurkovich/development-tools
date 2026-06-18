package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
	"golang.org/x/term"

	"github.com/observability-ui/development-tools/pkg/executor"
)

type CLIHandler struct {
	logger *log.Logger
}

func NewCLIHandler() *CLIHandler {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: false,
		ReportCaller:    false,
	})
	logger.SetColorProfile(termenv.Ascii)

	return &CLIHandler{logger: logger}
}

func (h *CLIHandler) HandleUpdate(update executor.ProgressUpdate) error {
	if update.Message != "" {
		h.logger.Print(update.Message)
		return nil
	}

	switch update.Status {
	case executor.StatusInProgress:
		h.logger.Print(update.Step + "...")
	case executor.StatusComplete:
		h.logger.Info("✓ " + update.Step)
	case executor.StatusFailed:
		h.logger.Error(fmt.Sprintf("✗ %s: %v", update.Step, update.Error))
		return update.Error
	}

	return nil
}

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
