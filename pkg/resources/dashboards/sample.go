package dashboards

import (
	"context"
	"time"

	sdkCommon "github.com/perses/perses/go-sdk/common"
	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/perses/go-sdk/panel"
	panelgroup "github.com/perses/perses/go-sdk/panel-group"
	listvariable "github.com/perses/perses/go-sdk/variable/list-variable"
	textvariable "github.com/perses/perses/go-sdk/variable/text-variable"
	"github.com/perses/perses/pkg/model/api/v1/common"
	promquery "github.com/perses/plugins/prometheus/sdk/go/query"
	labelvalues "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
	timeseries "github.com/perses/plugins/timeserieschart/sdk/go"
	staticlist "github.com/perses/plugins/staticlistvariable/sdk/go"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreatePersesSampleDashboard(ctx context.Context, kubeClient client.Client, namespace string) error {
	builder, err := dashboard.New("perses-dashboard-sample",
		dashboard.Name("Perses Dashboard Sample"),
		dashboard.Description("This is a sample dashboard"),
		dashboard.Duration(5*time.Minute),

		dashboard.AddVariable("job",
			listvariable.List(
				listvariable.DisplayName("job"),
				labelvalues.PrometheusLabelValues("job"),
			),
		),
		dashboard.AddVariable("instance",
			listvariable.List(
				listvariable.DisplayName("instance"),
				labelvalues.PrometheusLabelValues("instance",
					labelvalues.Matchers(`up{job=~"$job"}`),
				),
			),
		),
		dashboard.AddVariable("interval",
			listvariable.List(
				listvariable.DisplayName("interval"),
				staticlist.StaticList(staticlist.Values("1m", "5m")),
			),
		),
		dashboard.AddVariable("text",
			textvariable.Text("test", textvariable.Constant(true)),
		),

		dashboard.AddPanelGroup("Row 1",
			panelgroup.PanelsPerLine(3),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("RAM Used",
				panel.Description("This is a stat chart"),
				panel.Plugin(common.Plugin{
					Kind: "StatChart",
					Spec: statSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "percent"},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`100 - ((node_memory_MemAvailable_bytes{job='$job',instance=~'$instance'} * 100) / node_memory_MemTotal_bytes{job='$job',instance=~'$instance'})`),
				),
			),
			panelgroup.AddPanel("RAM Total",
				panel.Description("This is a stat chart"),
				panel.Plugin(common.Plugin{
					Kind: "StatChart",
					Spec: statSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "bytes", DecimalPlaces: intPtr(1)},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`node_memory_MemTotal_bytes{job='$job',instance=~'$instance'}`),
				),
			),
			panelgroup.AddPanel("RAM Used (Gauge)",
				panel.Description("This is a stat chart"),
				panel.Plugin(common.Plugin{
					Kind: "GaugeChart",
					Spec: gaugeSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "percent"},
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Value: 85},
								{Value: 95},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`100 - ((node_memory_MemAvailable_bytes{job='$job',instance=~'$instance'} * 100) / node_memory_MemTotal_bytes{job='$job',instance=~'$instance'})`),
				),
			),
		),

		dashboard.AddPanelGroup("Row 2",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Legend Example",
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition}),
					timeseries.WithYAxis(timeseries.YAxis{
						Show:   true,
						Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.BinaryBytesUnit)), ShortValues: true},
					}),
				),
				panel.AddQuery(
					promquery.PromQL(`node_memory_MemTotal_bytes{job='$job',instance=~'$instance'} - node_memory_MemFree_bytes{job='$job',instance=~'$instance'} - node_memory_Buffers_bytes{job='$job',instance=~'$instance'} - node_memory_Cached_bytes{job='$job',instance=~'$instance'}`,
						promquery.SeriesNameFormat("Node memory total"),
					),
				),
				panel.AddQuery(
					promquery.PromQL(`node_memory_Buffers_bytes{job='$job',instance=~'$instance'}`,
						promquery.SeriesNameFormat("Memory (buffers) - {{instance}}"),
					),
				),
				panel.AddQuery(
					promquery.PromQL(`node_memory_Cached_bytes{job='$job',instance=~'$instance'}`,
						promquery.SeriesNameFormat("Cached Bytes"),
					),
				),
				panel.AddQuery(
					promquery.PromQL(`node_memory_MemFree_bytes{job='$job',instance=~'$instance'}`,
						promquery.SeriesNameFormat("MemFree Bytes"),
					),
				),
			),
			panelgroup.AddPanel("Single Query",
				timeseries.Chart(
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.DecimalUnit))}}),
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.RightPosition}),
				),
				panel.AddQuery(
					promquery.PromQL(`1 - node_filesystem_free_bytes{job='$job',instance=~'$instance',fstype!="rootfs",mountpoint!~"/(run|var).*",mountpoint!=""} / node_filesystem_size_bytes{job='$job',instance=~'$instance'}`,
						promquery.SeriesNameFormat("Node memory - {{device}} {{instance}}"),
					),
				),
			),
		),

		dashboard.AddPanelGroup("Row 3",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(true),
			panelgroup.AddPanel("CPU - Gauge (Multi Series)",
				panel.Description("This is a gauge chart test"),
				panel.Plugin(common.Plugin{
					Kind: "GaugeChart",
					Spec: gaugeSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "percent-decimal"},
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Value: 0.2},
								{Value: 0.35},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`avg without (cpu)(rate(node_cpu_seconds_total{job='$job',instance=~'$instance',mode!="nice",mode!="steal",mode!="irq"}[$interval]))`,
						promquery.SeriesNameFormat("{{mode}} mode - {{job}} {{instance}}"),
					),
				),
			),
			panelgroup.AddPanel("CPU - Line (Multi Series)",
				panel.Description("This is a line chart test"),
				timeseries.Chart(
					timeseries.WithYAxis(timeseries.YAxis{
						Show:   false,
						Label:  "CPU Label",
						Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.PercentDecimalUnit))},
					}),
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition}),
				),
				panel.AddQuery(
					promquery.PromQL(`avg without (cpu)(rate(node_cpu_seconds_total{job='$job',instance=~'$instance',mode!="nice",mode!="steal",mode!="irq"}[$interval]))`,
						promquery.SeriesNameFormat("{{mode}} mode - {{job}} {{instance}}"),
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
