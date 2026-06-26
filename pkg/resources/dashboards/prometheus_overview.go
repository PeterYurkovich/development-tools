package dashboards

import (
	"context"
	"time"

	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/perses/go-sdk/panel"
	panelgroup "github.com/perses/perses/go-sdk/panel-group"
	listvariable "github.com/perses/perses/go-sdk/variable/list-variable"
	sdkCommon "github.com/perses/perses/go-sdk/common"
	promquery "github.com/perses/plugins/prometheus/sdk/go/query"
	labelvalues "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
	timeseries "github.com/perses/plugins/timeserieschart/sdk/go"
	"github.com/perses/perses/pkg/model/api/v1/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreatePrometheusOverviewDashboard(ctx context.Context, kubeClient client.Client, namespace string) error {
	builder, err := dashboard.New("prometheus-overview",
		dashboard.Name("Prometheus / Overview"),
		dashboard.Duration(time.Hour),

		dashboard.AddVariable("job",
			listvariable.List(
				listvariable.DisplayName("job"),
				labelvalues.PrometheusLabelValues("job",
					labelvalues.Matchers("prometheus_build_info{}"),
				),
			),
		),
		dashboard.AddVariable("instance",
			listvariable.List(
				listvariable.DisplayName("instance"),
				labelvalues.PrometheusLabelValues("instance",
					labelvalues.Matchers(`prometheus_build_info{job=~"$job"}`),
				),
			),
		),

		dashboard.AddPanelGroup("Overview",
			panelgroup.PanelsPerLine(1),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Prometheus Stats",
				panel.Description("Build info for all Prometheus instances"),
				panel.Plugin(common.Plugin{
					Kind: "Table",
					Spec: tableSpec{
						ColumnSettings: []columnSetting{
							{Name: "job", Header: "Job"},
							{Name: "instance", Header: "Instance"},
							{Name: "version", Header: "Version"},
							{Name: "value", Hide: true},
							{Name: "timestamp", Hide: true},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(`count by (job, instance, version) (prometheus_build_info{instance=~"$instance",job=~"$job"})`),
				),
			),
		),

		dashboard.AddPanelGroup("Targets",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Target Sync",
				panel.Description("Monitors target synchronization time for Prometheus instances"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.SecondsUnit))}}),
				),
				panel.AddQuery(
					promquery.PromQL(
						`sum(rate(prometheus_target_sync_length_seconds_sum{instance=~"$instance",job=~"$job"}[$__rate_interval])) by (scrape_job) * 1e3`,
						promquery.SeriesNameFormat("{{scrape_job}}"),
					),
				),
			),
			panelgroup.AddPanel("Targets",
				panel.Description("Number of targets discovered by Prometheus"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (job, instance) (prometheus_sd_discovered_targets{instance=~"$instance",job=~"$job"})`,
						promquery.SeriesNameFormat("{{job}} {{instance}}"),
					),
				),
			),
		),

		dashboard.AddPanelGroup("Retrieval",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Average Scrape Interval Duration",
				panel.Description("Average duration of scrape intervals"),
				timeseries.Chart(
					timeseries.WithYAxis(timeseries.YAxis{Format: &sdkCommon.Format{Unit: strPtr(string(sdkCommon.SecondsUnit))}}),
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(
						`rate(prometheus_target_interval_length_seconds_sum{instance=~"$instance",job=~"$job"}[$__rate_interval]) / rate(prometheus_target_interval_length_seconds_count{instance=~"$instance",job=~"$job"}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{interval}} configured"),
					),
				),
			),
			panelgroup.AddPanel("Scrape failures",
				panel.Description("Number of scrape failures per interval"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (job) (rate(prometheus_target_scrapes_exceeded_sample_limit_total{instance=~"$instance",job=~"$job"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("exceeded sample limit: {{job}}"),
					),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (job) (rate(prometheus_target_scrapes_sample_duplicate_timestamp_total{instance=~"$instance",job=~"$job"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("duplicate timestamp: {{job}}"),
					),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (job) (rate(prometheus_target_scrapes_sample_out_of_bounds_total{instance=~"$instance",job=~"$job"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("out of bounds: {{job}}"),
					),
				),
				panel.AddQuery(
					promquery.PromQL(`sum by (job) (rate(prometheus_target_scrapes_sample_out_of_order_total{instance=~"$instance",job=~"$job"}[$__rate_interval]))`,
						promquery.SeriesNameFormat("out of order: {{job}}"),
					),
				),
			),
		),

		dashboard.AddPanelGroup("TSDB",
			panelgroup.PanelsPerLine(2),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Appended Samples",
				panel.Description("Rate of samples appended to TSDB"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`rate(prometheus_tsdb_head_samples_appended_total{instance=~"$instance",job=~"$job"}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{job}} {{instance}}"),
					),
				),
			),
			panelgroup.AddPanel("Head Series",
				panel.Description("Number of active series in TSDB head"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`prometheus_tsdb_head_series{instance=~"$instance",job=~"$job"}`,
						promquery.SeriesNameFormat("{{job}} {{instance}}"),
					),
				),
			),
			panelgroup.AddPanel("Head Chunks",
				panel.Description("Number of chunks in TSDB head"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`prometheus_tsdb_head_chunks{instance=~"$instance",job=~"$job"}`,
						promquery.SeriesNameFormat("{{job}} {{instance}}"),
					),
				),
			),
			panelgroup.AddPanel("Query Rate",
				panel.Description("Rate of queries handled by Prometheus"),
				timeseries.Chart(
					timeseries.WithLegend(timeseries.Legend{Position: timeseries.BottomPosition, Mode: timeseries.TableMode}),
				),
				panel.AddQuery(
					promquery.PromQL(`rate(prometheus_engine_query_duration_seconds_count{instance=~"$instance",job=~"$job",slice="inner_eval"}[$__rate_interval])`,
						promquery.SeriesNameFormat("{{job}} {{instance}}"),
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
