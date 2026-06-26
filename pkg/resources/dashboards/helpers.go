package dashboards

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/perses/perses/go-sdk/dashboard"
	persesv1alpha2 "github.com/rhobs/perses-operator/api/v1alpha2"
	specDashboard "github.com/perses/spec/go/dashboard"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// gaugeSpec is used for GaugeChart panels.
type gaugeSpec struct {
	Calculation string      `json:"calculation"`
	Format      formatSpec  `json:"format"`
	Max         *float64    `json:"max,omitempty"`
	Thresholds  *thresholds `json:"thresholds,omitempty"`
}

// statSpec is used for StatChart panels.
type statSpec struct {
	Calculation string      `json:"calculation"`
	ColorMode   string      `json:"colorMode,omitempty"`
	Format      formatSpec  `json:"format"`
	MetricLabel string      `json:"metricLabel,omitempty"`
	Sparkline   *sparkline  `json:"sparkline,omitempty"`
	Thresholds  *thresholds `json:"thresholds,omitempty"`
}

type formatSpec struct {
	Unit          string `json:"unit"`
	DecimalPlaces *int   `json:"decimalPlaces,omitempty"`
	ShortValues   bool   `json:"shortValues,omitempty"`
}

type sparkline struct {
	Color string  `json:"color,omitempty"`
	Width float64 `json:"width,omitempty"`
}

type thresholds struct {
	Steps []thresholdStep `json:"steps"`
}

type thresholdStep struct {
	Color string  `json:"color,omitempty"`
	Name  string  `json:"name,omitempty"`
	Value float64 `json:"value"`
}

// tableSpec is used for Table panels.
type tableSpec struct {
	ColumnSettings []columnSetting `json:"columnSettings,omitempty"`
}

type columnSetting struct {
	Name   string `json:"name"`
	Header string `json:"header,omitempty"`
	Hide   bool   `json:"hide,omitempty"`
}

// applyDashboard marshals a Go SDK builder's spec into a PersesDashboard CR and creates it.
func applyDashboard(ctx context.Context, kubeClient client.Client, namespace string, builder dashboard.Builder) error {
	specJSON, err := json.Marshal(builder.Dashboard.Spec)
	if err != nil {
		return fmt.Errorf("failed to marshal dashboard spec: %w", err)
	}

	var dashSpec specDashboard.Spec
	if err := json.Unmarshal(specJSON, &dashSpec); err != nil {
		return fmt.Errorf("failed to unmarshal dashboard spec: %w", err)
	}

	cr := &persesv1alpha2.PersesDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Dashboard.Metadata.Name,
			Namespace: namespace,
		},
		Spec: persesv1alpha2.PersesDashboardSpec{
			Config: persesv1alpha2.Dashboard{
				Spec: dashSpec,
			},
		},
	}

	err = kubeClient.Create(ctx, cr)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create PersesDashboard %s: %w", builder.Dashboard.Metadata.Name, err)
	}
	return nil
}

func intPtr(val int) *int {
	return &val
}

func float64Ptr(val float64) *float64 {
	return &val
}

func strPtr(val string) *string {
	return &val
}
