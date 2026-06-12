package output

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/term"

	execctx "github.com/observability-ui/development-tools/pkg/context"
	"github.com/observability-ui/development-tools/pkg/tui"
)

type Handler struct {
	ctx context.Context
}

func NewHandler(ctx context.Context) *Handler {
	return &Handler{ctx: ctx}
}

func (h *Handler) Info(message string) {
	if execctx.IsTUI(h.ctx) {
		fmt.Println(tui.InfoStyle.Render(message))
	} else {
		fmt.Println(message)
	}
}

func (h *Handler) Success(message string) {
	if execctx.IsTUI(h.ctx) {
		fmt.Println(tui.SuccessStyle.Render("✓ " + message))
	} else {
		fmt.Println(message)
	}
}

func (h *Handler) Error(message string) {
	if execctx.IsTUI(h.ctx) {
		fmt.Fprintln(os.Stderr, tui.ErrorStyle.Render("✗ "+message))
	} else {
		fmt.Fprintln(os.Stderr, "Error: "+message)
	}
}

func (h *Handler) Progress(message string) {
	if execctx.IsTUI(h.ctx) {
		fmt.Println(tui.ProgressStyle.Render("⋯ " + message))
	} else {
		fmt.Println(message)
	}
}

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
