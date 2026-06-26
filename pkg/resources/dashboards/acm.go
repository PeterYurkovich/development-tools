package dashboards

import (
	"context"
	"time"

	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/perses/go-sdk/panel"
	panelgroup "github.com/perses/perses/go-sdk/panel-group"
	listvariable "github.com/perses/perses/go-sdk/variable/list-variable"
	"github.com/perses/perses/pkg/model/api/v1/common"
	promquery "github.com/perses/plugins/prometheus/sdk/go/query"
	labelvalues "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateACMDashboard(ctx context.Context, kubeClient client.Client, namespace string) error {
	builder, err := dashboard.New("acm-vm-status",
		dashboard.Name("ACM VM Status"),
		dashboard.Duration(time.Hour),

		dashboard.AddVariable("cluster",
			listvariable.List(
				listvariable.DisplayName("cluster"),
				listvariable.AllowMultiple(true),
				labelvalues.PrometheusLabelValues("cluster",
					labelvalues.Matchers("kubevirt_vm_info{}"),
				),
			),
		),
		dashboard.AddVariable("namespace",
			listvariable.List(
				listvariable.DisplayName("namespace"),
				listvariable.AllowMultiple(true),
				labelvalues.PrometheusLabelValues("namespace",
					labelvalues.Matchers(`kubevirt_vm_info{cluster=~"$cluster"}`),
				),
			),
		),
		dashboard.AddVariable("name",
			listvariable.List(
				listvariable.DisplayName("name"),
				listvariable.AllowMultiple(true),
				labelvalues.PrometheusLabelValues("name",
					labelvalues.Matchers(`kubevirt_vm_info{cluster=~"$cluster",namespace=~"$namespace"}`),
				),
			),
		),

		dashboard.AddPanelGroup("VM Resources",
			panelgroup.PanelsPerLine(3),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Total Allocated CPU",
				panel.Description("The total CPUs of the VMs that are listed in the dashboard"),
				panel.Plugin(common.Plugin{
					Kind: "StatChart",
					Spec: statSpec{
						Calculation: "last-number",
						ColorMode:   "value",
						Format:      formatSpec{Unit: "decimal"},
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Color: "#73bf69", Value: 0},
								{Color: "#f2495c", Value: 80},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(
						`sum(max by (cluster, namespace, name, status)(kubevirt_vm_resource_requests{cluster=~"$cluster", name=~"$name", namespace=~"$namespace", resource="cpu"}) + on(cluster, name, namespace) group_left(status) 0 * kubevirt_vm_info{cluster=~"$cluster", name=~"$name", namespace=~"$namespace"})`,
					),
				),
			),
			panelgroup.AddPanel("Total Allocated Memory",
				panel.Description("The total Memory of the VMs that are listed in the dashboard"),
				panel.Plugin(common.Plugin{
					Kind: "StatChart",
					Spec: statSpec{
						Calculation: "last-number",
						ColorMode:   "value",
						Format:      formatSpec{Unit: "bytes"},
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Color: "#73bf69", Value: 0},
								{Color: "#f2495c", Value: 80},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(
						`sum(max by (cluster, namespace, name, status)(kubevirt_vm_resource_requests{cluster=~"$cluster", name=~"$name", namespace=~"$namespace", resource="memory"}) + on(cluster, name, namespace) group_left(status) 0 * kubevirt_vm_info{cluster=~"$cluster", name=~"$name", namespace=~"$namespace"})`,
					),
				),
			),
			panelgroup.AddPanel("Total Allocated Disk",
				panel.Description("The total disk size of the VMs that are listed in the dashboard"),
				panel.Plugin(common.Plugin{
					Kind: "StatChart",
					Spec: statSpec{
						Calculation: "last-number",
						ColorMode:   "value",
						Format:      formatSpec{Unit: "bytes"},
						Thresholds: &thresholds{
							Steps: []thresholdStep{
								{Color: "#73bf69", Value: 0},
								{Color: "#f2495c", Value: 80},
							},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(
						`sum(kubevirt_vm_disk_allocated_size_bytes{cluster=~"$cluster", name=~"$name", namespace=~"$namespace"} + on(cluster, name, namespace) group_left(status) 0 * kubevirt_vm_info{cluster=~"$cluster", name=~"$name", namespace=~"$namespace"})`,
					),
				),
			),
		),

		dashboard.AddPanelGroup("VM Status",
			panelgroup.PanelsPerLine(1),
			panelgroup.Collapsed(false),
			panelgroup.AddPanel("Virtual Machines List by Time In Status",
				panel.Plugin(common.Plugin{
					Kind: "Table",
					Spec: tableSpec{
						ColumnSettings: []columnSetting{
							{Name: "timestamp", Hide: true},
							{Name: "cluster", Header: "Cluster"},
							{Name: "clusterID", Hide: true},
							{Name: "container", Hide: true},
							{Name: "endpoint", Hide: true},
							{Name: "instance", Hide: true},
							{Name: "namespace", Header: "Namespace"},
							{Name: "name", Header: "VM Name"},
							{Name: "status", Header: "Status"},
							{Name: "value", Header: "Time in Status (s)"},
						},
					},
				}),
				panel.AddQuery(
					promquery.PromQL(
						`sum by (cluster, namespace, name, status)((time() - label_replace(kubevirt_vm_starting_status_last_transition_timestamp_seconds{cluster=~"$cluster", name=~"$name", namespace=~"$namespace"} > 0, "status", "starting", "", "")) or (time() - label_replace(kubevirt_vm_running_status_last_transition_timestamp_seconds{cluster=~"$cluster", name=~"$name", namespace=~"$namespace"} > 0, "status", "running", "", "")) or (time() - label_replace(kubevirt_vm_stopped_status_last_transition_timestamp_seconds{cluster=~"$cluster", name=~"$name", namespace=~"$namespace"} > 0, "status", "stopped", "", "")))`,
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
