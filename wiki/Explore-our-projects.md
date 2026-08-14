# Explore Cluster Observability Operator and its UIPlugins

## COO - Cluster Observability Operator installation

### From UI
1. Depending on OpenShift version, find in the left menu `Operators > Operator Hub` or `Ecosystem > Software Catalog`
2. Search for and install `Cluster Observability Operator`
3. Under `Installed Operators`, verify Cluster Observability Operator succeeded <br>
3.1 In the left menu ‘Workloads’ > Pods and change the Project dropdown to `openshift-cluster-observability-operator` namespace <br>
3.2 Take a look at pods created when you have Cluster Observability Operator installed. For our team, it is important to take a look at `perses-operator` pod

### From OC commands
It will install the released version from Operator Hub:
- https://github.com/observability-ui/development-tools/blob/main/perses/demo/z_coo_only.sh

### Checking `openshift-cluster-observability-operator` pods
- `obo-prometheus-operator`
- `obo-prometheus-admission-webhook`
- `observability-operator`
- `perses-operator` - this one is important for Obs UI team

## COO - UIPlugins

References: <br>
- [Observability UIPlugin](https://github.com/rhobs/observability-operator/blob/main/docs/user-guides/observability-ui-plugins.md)<br>
- [UIPlugins official doc](https://docs.redhat.com/en/documentation/red_hat_openshift_cluster_observability_operator/1-latest/html-single/ui_plugins_for_red_hat_openshift_cluster_observability_operator/index#observability-ui-plugins-overview)
- [Cluster Observability Operator Acceptance Test](https://docs.google.com/document/d/1rBio1lFD3GzWqVP35ZLUdfVDNPwjDSxG08wy5Yo2tG8/edit?usp=sharing)

4. Back to OpenShift, under `Installed Operators`, click `Cluster Observability Operator`, you will see many tabs with a horizontal scroll
5. You can fill the form or apply YAML files

### Monitoring UIPlugin

When deploying Monitoring UIPlugin, `monitoring` pod is created. <br>

Monitoring UIPlugin contains “feature flags” that will enable/disable functionalities in the UI:
- Incidents - it will create a new tab under menu `Observe > Alerting > Incidents`
    - `health-analyzer` pod
- Perses - it will create a new page under menu `Observe > Dashboards (Perses)`
    - `perses-0` pod
- ACM (aka Advanced Cluster Management) - if ACM is installed in the cluster, then a new perspective is enabled: Fleet management. Observe menu will be enabled containing Alerting and Dashboards pages, and in this case, even if Dashboards menu entry does not contain `(Perses)` in the name, it is related to it, and not Grafana.

*Incidents tab is not enabled on*:
- Under `Developer` perspective
- Under `Fleet management` perspective
- Under `Fleet virtualization` or `Virtualization` perspective

### Logging UIPlugin

It depends on `Red Hat OpenShift Logging` and `Loki Operator` operators to get reconciled. 

By running this script, it will install both, configure and also install Logging UIPlugin:
- https://github.com/observability-ui/development-tools/blob/main/perses/demo/z_logging.sh <br>

This script also creates a log generator to create data to be shown on Logs page.

When it is successfully installed and reconciled, it enables:
- Under `Core platform` or `Administrator` perspective
    - `Observe` > `Logs`
- Under `Developer` perspective
    - `Workloads > Pods > in a pod > Aggregated Logs` <br>
- Under `Workloads > Pods`, change Project `openshift-cluster-observability-operator` namespace
    - `logging` pod

*Logs page is not enabled on*:
- Under `Fleet management` perspective
- Under `Fleet virtualization` or `Virtualization` perspective


### Distributed Tracing UIPlugin

It does not depend on `Red Hat build of OpenTelemetry` and `Tempo Operator` operators to get reconciled and be enabled under `Observe > Traces`, however they are required to have data being shown in this page.

By running this script, it will install both, configure and also install Distributed Tracing UIPlugin:
- https://github.com/observability-ui/development-tools/blob/main/perses/demo/z_tracing.sh <br>

This script also creates a log generator to create data to be shown on Traces page.

When it is successfully installed and reconciled, it enables:
- Under `Core platform` or `Administrator` perspective
    - `Observe` > `Traces`
- Under `Workloads > Pods`, change Project `openshift-cluster-observability-operator` namespace
    - `distributed-tracing` pod

*Traces page is not enabled on*:
- Under `Developer` perspective
- Under `Fleet management` perspective
- Under `Fleet virtualization` or `Virtualization` perspective

### Trobleshooting Panel UIPlugin

It does not depend on any operator to get reconciled, however it integrates to other operators, aka Observability signals and UIPlugins.

This plugin enables you to "correlate" cluster resources to troubleshoot problems, including the integration to:<br>
- Logging<br>
- Distributed Tracing<br>
- Network Observability<br>

When it is successfully installed and reconciled, it enables:
- Under `Core platform` or `Administrator` perspective
    - In masthead, Application switcher icon (9 dots), a `Signal correlation` option, showing a sliding drawer on the right side
    - `Observe > Alerting > Alerts > click on an alert > Troubleshooting panel link above the metric chart`
- Under `Workloads > Pods`, change Project `openshift-cluster-observability-operator` namespace
    -`korrel8r` pod
    - `troubleshooting-panel` pod

*These entries are not enabled on*:
- Under `Developer` perspective
- Under `Fleet management` perspective
- Under `Fleet virtualization` or `Virtualization` perspective