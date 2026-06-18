package output

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
	"golang.org/x/term"

	execctx "github.com/observability-ui/development-tools/pkg/context"
)

type Handler struct {
	ctx    context.Context
	logger *log.Logger
}

func NewHandler(ctx context.Context) *Handler {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: false,
		ReportCaller:    false,
	})

	if execctx.IsTUI(ctx) {
		logger.SetColorProfile(termenv.TrueColor)
	} else {
		logger.SetColorProfile(termenv.Ascii)
	}

	return &Handler{
		ctx:    ctx,
		logger: logger,
	}
}

func (h *Handler) Info(message string) {
	if execctx.IsTUI(h.ctx) {
		h.logger.Info(message)
	} else {
		h.logger.Print(message)
	}
}

func (h *Handler) Success(message string) {
	if execctx.IsTUI(h.ctx) {
		h.logger.SetPrefix("✓ ")
		h.logger.Info(message)
		h.logger.SetPrefix("")
	} else {
		h.logger.Print(message)
	}
}

func (h *Handler) Error(message string) {
	h.logger.Error(message)
}

func (h *Handler) Progress(message string) {
	if execctx.IsTUI(h.ctx) {
		h.logger.SetPrefix("⋯ ")
		h.logger.Info(message)
		h.logger.SetPrefix("")
	} else {
		h.logger.Print(message)
	}
}

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
