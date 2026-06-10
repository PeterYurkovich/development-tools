package context

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type contextKey int

const (
	clientKey contextKey = iota
	tuiKey
)

func WithClient(ctx context.Context, kubeClient client.Client) context.Context {
	return context.WithValue(ctx, clientKey, kubeClient)
}

func GetClient(ctx context.Context) (client.Client, error) {
	kubeClient, ok := ctx.Value(clientKey).(client.Client)
	if !ok {
		return nil, fmt.Errorf("kubernetes client not found in context")
	}
	return kubeClient, nil
}

func WithTUI(ctx context.Context, isTUI bool) context.Context {
	return context.WithValue(ctx, tuiKey, isTUI)
}

func IsTUI(ctx context.Context) bool {
	isTUI, ok := ctx.Value(tuiKey).(bool)
	if !ok {
		return false
	}
	return isTUI
}
