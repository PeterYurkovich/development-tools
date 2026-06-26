package dashboards

import (
	"context"
	"time"

	sdkCommon "github.com/perses/perses/go-sdk/common"
	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/perses/go-sdk/panel"
	panelgroup "github.com/perses/perses/go-sdk/panel-group"
	listvariable "github.com/perses/perses/go-sdk/variable/list-variable"
	"github.com/perses/perses/pkg/model/api/v1/common"
	promquery "github.com/perses/plugins/prometheus/sdk/go/query"
	labelvalues "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
	timeseries "github.com/perses/plugins/timeserieschart/sdk/go"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateThanosCompactDashboard(ctx context.Context, kubeClient client.Client, namespace string) error {
	builder, err := dashboard.New("thanos-compact-overview",
		dashboard.Name("Thanos / Compact / Overview"),
		dashboard.Duration(time.Hour),

		dashboard.AddVariable("job",
			listvariable.List(
				listvariable.DisplayName("job"),
				listvariable.AllowMultiple(true),
				labelvalues.PrometheusLabelValues("job",
					labelvalues.Matchers(`thanos_build_info{container="thanos-compact"}`),
				),
			),
		),
		dashboard.AddVariable("namespace",
			listvariable.List(
				listvariable.DisplayName("namespace"),
				labelvalues.PrometheusLabelValues("namespace",
					labelvalues.Matchers("thanos_status{}"),
				),
			),
		),

		dashboard.AddPanelGroup("TODO Queue",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("TODO Compaction Blocks",
				panel.Description("Shows number of blocks planned to be compacted."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.DecimalUnit))}}),
					timeseries.WithVisual(timeseries.Visual{Display: timeseries.LineDisplay, LineWidth: 0.25, AreaOpacity: 1, Stack: timeseries.AllStack, Palette: &timeseries.Palette{Mode: timeseries.AutoMode}}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (namespace, job) (thanos_compact_todo_compaction_blocks{job=~"$job",namespace="$namespace"})`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
			panelgroup.AddPanel("TODO Compactions",
				panel.Description("Shows number of compaction operations to be done."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.DecimalUnit))}}),
					timeseries.WithVisual(timeseries.Visual{Display: timeseries.LineDisplay, LineWidth: 0.25, AreaOpacity: 1, Stack: timeseries.AllStack, Palette: &timeseries.Palette{Mode: timeseries.AutoMode}}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (namespace, job) (thanos_compact_todo_compactions{job=~"$job",namespace="$namespace"})`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
			panelgroup.AddPanel("TODO Deletions",
				panel.Description("Shows number of block deletions to be done."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.DecimalUnit))}}),
					timeseries.WithVisual(timeseries.Visual{Display: timeseries.LineDisplay, LineWidth: 0.25, AreaOpacity: 1, Stack: timeseries.AllStack, Palette: &timeseries.Palette{Mode: timeseries.AutoMode}}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (namespace, job) (thanos_compact_todo_deletion_blocks{job=~"$job",namespace="$namespace"})`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
			panelgroup.AddPanel("TODO Downsamples",
				panel.Description("Shows number of downsampling operations to be done."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.DecimalUnit))}}),
					timeseries.WithVisual(timeseries.Visual{Display: timeseries.LineDisplay, LineWidth: 0.25, AreaOpacity: 1, Stack: timeseries.AllStack, Palette: &timeseries.Palette{Mode: timeseries.AutoMode}}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (namespace, job) (thanos_compact_todo_downsample_blocks{job=~"$job",namespace="$namespace"})`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
		),

		dashboard.AddPanelGroup("Compaction",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Group Compactions",
				panel.Description("Rate of group compaction operations."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (namespace, job) (rate(thanos_compact_group_compactions_total{job=~"$job",namespace="$namespace"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
			panelgroup.AddPanel("Group Compaction Errors",
				panel.Description("Rate of group compaction errors."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (namespace, job) (rate(thanos_compact_group_compactions_failures_total{job=~"$job",namespace="$namespace"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
		),

		dashboard.AddPanelGroup("Downsampling",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Downsample Rate",
				panel.Description("Rate of downsampling operations."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (namespace, job) (rate(thanos_compact_downsample_total{job=~"$job",namespace="$namespace"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
			panelgroup.AddPanel("Downsample Errors",
				panel.Description("Rate of downsampling errors."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (namespace, job) (rate(thanos_compact_downsample_failures_total{job=~"$job",namespace="$namespace"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
		),

		dashboard.AddPanelGroup("Bucket Operations",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Bucket Operations",
				panel.Description("Rate of object storage operations."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (namespace, job, operation) (rate(thanos_objstore_bucket_operations_total{job=~"$job",namespace="$namespace"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("{{job}} {{operation}}"),
					),
				),
			),
			panelgroup.AddPanel("Bucket Operation Errors",
				panel.Description("Rate of object storage operation errors."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (namespace, job, operation) (rate(thanos_objstore_bucket_operation_failures_total{job=~"$job",namespace="$namespace"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("{{job}} {{operation}}"),
					),
				),
			),
			panelgroup.AddPanel("Bucket Operation Latency",
				panel.Description("P99 latency of object storage operations."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.SecondsUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(`histogram_quantile(0.99, sum by (namespace, job, operation, le) (rate(thanos_objstore_bucket_operation_duration_seconds_bucket{job=~"$job",namespace="$namespace"}[$__rate_interval])))`,
						promquery.SeriesNameFormat("p99 {{job}} {{operation}}"),
					),
				),
			),
		),

		dashboard.AddPanelGroup("Resources",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("CPU Usage",
				panel.Description("CPU usage of the Thanos Compact component."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.DecimalUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(`rate(process_cpu_seconds_total{job=~"$job",namespace="$namespace"}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
			panelgroup.AddPanel("Memory Usage",
				panel.Description("Memory usage of the Thanos Compact component."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.BinaryBytesUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(`go_memstats_alloc_bytes{job=~"$job",namespace="$namespace"}`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
			panelgroup.AddPanel("Goroutines",
				panel.Description("Number of goroutines in the Thanos Compact component."),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`go_goroutines{job=~"$job",namespace="$namespace"}`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
			panelgroup.AddPanel("Halted Compactors",
				panel.Description("Number of halted Thanos Compact instances."),
				panel.Plugin(common.Plugin{
					Kind: "StatChart",
					Spec: statSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "decimal"},
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Color: "green", Value: 0},
								{Color: "red", Value: 1},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`max by (namespace, job) (thanos_compact_halted{job=~"$job",namespace="$namespace"})`,
						promquery.SeriesNameFormat("{{job}} {{namespace}}"),
					),
				),
			),
		),
	)
	if err != nil {
		return err
	}

	return applyDashboard(ctx, kubeClient, namespace, builder)
}
