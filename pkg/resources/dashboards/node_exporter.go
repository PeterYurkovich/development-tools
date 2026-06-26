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

func CreateNodeExporterDashboard(ctx context.Context, kubeClient client.Client, namespace string) error {
	builder, err := dashboard.New("nodeexporterfull",
		dashboard.Name("Node Exporter Full"),
		dashboard.Duration(time.Hour),

		dashboard.AddVariable("job",
			listvariable.List(
				listvariable.DisplayName("job"),
				labelvalues.PrometheusLabelValues("job",
					labelvalues.Matchers("node_uname_info{}"),
				),
			),
		),
		dashboard.AddVariable("node",
			listvariable.List(
				listvariable.DisplayName("node"),
				labelvalues.PrometheusLabelValues("instance",
					labelvalues.Matchers(`node_uname_info{job=~"$job"}`),
				),
			),
		),

		dashboard.AddPanelGroup("CPU",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("CPU Busy",
				panel.Description("Busy state of all CPU cores together"),
				panel.Plugin(common.Plugin{
					Kind: "GaugeChart",
					Spec: gaugeSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "percent"},
						Max:         float64Ptr(100),
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Color: "rgba(50, 172, 45, 0.97)", Value: 0},
								{Color: "rgba(237, 129, 40, 0.89)", Value: 85},
								{Color: "rgba(245, 54, 54, 0.9)", Value: 95},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`(sum by(instance) (irate(node_cpu_seconds_total{instance="$node",job="$job",mode!="idle"}[$__rate_interval])) / on(instance) group_left sum by (instance)((irate(node_cpu_seconds_total{instance="$node",job="$job"}[$__rate_interval])))) * 100`),
				),
			),
			panelgroup.AddPanel("Sys Load (5m avg)",
				panel.Description("Busy state of all CPU cores together (5 min average)"),
				panel.Plugin(common.Plugin{
					Kind: "GaugeChart",
					Spec: gaugeSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "percent"},
						Max:         float64Ptr(100),
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Color: "rgba(50, 172, 45, 0.97)", Value: 0},
								{Color: "rgba(237, 129, 40, 0.89)", Value: 85},
								{Color: "rgba(245, 54, 54, 0.9)", Value: 95},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`avg_over_time(node_load5{instance="$node",job="$job"}[$__rate_interval]) * 100 / on(instance) group_left sum by (instance)(irate(node_cpu_seconds_total{instance="$node",job="$job"}[$__rate_interval]))`),
				),
			),
			panelgroup.AddPanel("CPU Cores",
				panel.Description("Number of CPU cores"),
				panel.Plugin(common.Plugin{
					Kind: "StatChart",
					Spec: statSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "decimal"},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`count(count(node_cpu_seconds_total{instance="$node",job="$job"}) by (cpu))`),
				),
			),
			panelgroup.AddPanel("CPU Usage by Mode",
				panel.Description("CPU usage breakdown by mode"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.PercentDecimalUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (mode) (irate(node_cpu_seconds_total{instance="$node",job="$job"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("{{mode}}"),
					),
				),
			),
		),

		dashboard.AddPanelGroup("Memory",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("RAM Used",
				panel.Description("RAM used percentage"),
				panel.Plugin(common.Plugin{
					Kind: "GaugeChart",
					Spec: gaugeSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "percent"},
						Max:         float64Ptr(100),
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Color: "rgba(50, 172, 45, 0.97)", Value: 0},
								{Color: "rgba(237, 129, 40, 0.89)", Value: 80},
								{Color: "rgba(245, 54, 54, 0.9)", Value: 90},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`100 - ((node_memory_MemAvailable_bytes{instance="$node",job="$job"} * 100) / node_memory_MemTotal_bytes{instance="$node",job="$job"})`),
				),
			),
			panelgroup.AddPanel("RAM Total",
				panel.Description("Total RAM available"),
				panel.Plugin(common.Plugin{
					Kind: "StatChart",
					Spec: statSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "bytes"},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`node_memory_MemTotal_bytes{instance="$node",job="$job"}`),
				),
			),
			panelgroup.AddPanel("Memory Usage",
				panel.Description("Memory usage breakdown over time"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.BinaryBytesUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(`node_memory_MemTotal_bytes{instance="$node",job="$job"} - node_memory_MemFree_bytes{instance="$node",job="$job"} - node_memory_Buffers_bytes{instance="$node",job="$job"} - node_memory_Cached_bytes{instance="$node",job="$job"}`,
						promquery.SeriesNameFormat("used"),
					),
				),
				panel.AddQuery(
					promquery.PromQL(`node_memory_Buffers_bytes{instance="$node",job="$job"}`,
						promquery.SeriesNameFormat("buffers"),
					),
				),
				panel.AddQuery(
					promquery.PromQL(`node_memory_Cached_bytes{instance="$node",job="$job"}`,
						promquery.SeriesNameFormat("cached"),
					),
				),
				panel.AddQuery(
					promquery.PromQL(`node_memory_MemFree_bytes{instance="$node",job="$job"}`,
						promquery.SeriesNameFormat("free"),
					),
				),
			),
			panelgroup.AddPanel("SWAP Used",
				panel.Description("SWAP usage percentage"),
				panel.Plugin(common.Plugin{
					Kind: "GaugeChart",
					Spec: gaugeSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "percent"},
						Max:         float64Ptr(100),
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Color: "rgba(50, 172, 45, 0.97)", Value: 0},
								{Color: "rgba(237, 129, 40, 0.89)", Value: 10},
								{Color: "rgba(245, 54, 54, 0.9)", Value: 25},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`((node_memory_SwapTotal_bytes{instance="$node",job="$job"} - node_memory_SwapFree_bytes{instance="$node",job="$job"}) / (node_memory_SwapTotal_bytes{instance="$node",job="$job"} != 0)) * 100`),
				),
			),
		),

		dashboard.AddPanelGroup("Disk",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Disk Space Used (root)",
				panel.Description("Disk space used on root filesystem"),
				panel.Plugin(common.Plugin{
					Kind: "GaugeChart",
					Spec: gaugeSpec{
						Calculation: "last-number",
						Format:      formatSpec{Unit: "percent"},
						Max:         float64Ptr(100),
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Color: "rgba(50, 172, 45, 0.97)", Value: 0},
								{Color: "rgba(237, 129, 40, 0.89)", Value: 80},
								{Color: "rgba(245, 54, 54, 0.9)", Value: 90},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`100 - ((node_filesystem_avail_bytes{instance="$node",job="$job",mountpoint="/",fstype!="rootfs"} * 100) / node_filesystem_size_bytes{instance="$node",job="$job",mountpoint="/",fstype!="rootfs"})`),
				),
			),
			panelgroup.AddPanel("Disk Read",
				panel.Description("Disk read throughput"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.BinaryBytesUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(`rate(node_disk_read_bytes_total{instance="$node",job="$job",device!=""}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{device}} read"),
					),
				),
			),
			panelgroup.AddPanel("Disk Written",
				panel.Description("Disk write throughput"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.BinaryBytesUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(`rate(node_disk_written_bytes_total{instance="$node",job="$job",device!=""}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{device}} written"),
					),
				),
			),
			panelgroup.AddPanel("Disk IO Time",
				panel.Description("Time spent doing I/O"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.SecondsUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(`rate(node_disk_io_time_seconds_total{instance="$node",job="$job",device!=""}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{device}}"),
					),
				),
			),
		),

		dashboard.AddPanelGroup("Network",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Network Received",
				panel.Description("Network traffic received"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.BinaryBytesUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(`rate(node_network_receive_bytes_total{instance="$node",job="$job",device!~"lo|docker.*|veth.*|br.*|virbr.*"}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{device}} received"),
					),
				),
			),
			panelgroup.AddPanel("Network Transmitted",
				panel.Description("Network traffic transmitted"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.BinaryBytesUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(`rate(node_network_transmit_bytes_total{instance="$node",job="$job",device!~"lo|docker.*|veth.*|br.*|virbr.*"}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{device}} transmitted"),
					),
				),
			),
			panelgroup.AddPanel("Network Errors",
				panel.Description("Network errors received and transmitted"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`rate(node_network_receive_errs_total{instance="$node",job="$job",device!~"lo|docker.*|veth.*|br.*|virbr.*"}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{device}} recv errors"),
					),
				),
				panel.AddQuery(
					promquery.PromQL(`rate(node_network_transmit_errs_total{instance="$node",job="$job",device!~"lo|docker.*|veth.*|br.*|virbr.*"}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{device}} transmit errors"),
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
