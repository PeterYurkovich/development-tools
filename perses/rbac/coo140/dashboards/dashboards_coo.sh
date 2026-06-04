#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

oc new-project perses-dev 2>/dev/null || oc project perses-dev

echo ""
echo "Applying dashboards and datasources to namespace: perses-dev"
echo "============================================================"

# List of YAML files to apply
YAML_FILES=(
    "openshift-cluster-sample-dashboard.yaml"
    "perses-dashboard-sample.yaml"
    "prometheus-overview-variables.yaml"
    "thanos-compact-overview-1var.yaml"
    "thanos-querier-datasource.yaml"
)

# Apply each file with namespace replaced
for file in "${YAML_FILES[@]}"; do
    echo "Applying $file..."
    oc apply -f "$SCRIPT_DIR/$file"
done

echo "=== PersesDashboard exercising dynamic queryParams ==="
oc apply -n perses-dev -f - <<'EOF'
apiVersion: perses.dev/v1alpha2
kind: PersesDashboard
metadata:
  name: dynamic-queryparams-test
spec:
  config:
    display:
      name: "Dynamic QueryParams Test (perses#3940)"
      description: >-
        Exercises the dynamic query parameter interpolation feature.
        The "namespace" variable value is forwarded as a query parameter
        on every proxied request to Thanos Querier via the datasource's
        queryParams configuration.
    duration: 30m
    variables:
      - kind: ListVariable
        spec:
          name: namespace
          display:
            name: Namespace
            hidden: false
          allowMultiple: true
          allowAllValue: false
          plugin:
            kind: PrometheusLabelValuesVariable
            spec:
              datasource:
                kind: PrometheusDatasource
                name: thanos-querier-dynamic-queryparams
              labelName: namespace
      - kind: ListVariable
        spec:
          name: workload_type
          display:
            name: Workload Type
            hidden: false
          allowMultiple: false
          allowAllValue: false
          plugin:
            kind: StaticListVariable
            spec:
              values:
                - deployment
                - daemonset
                - statefulset
    panels:
      pod_count:
        kind: Panel
        spec:
          display:
            name: "Running Pods by Namespace"
            description: >-
              Shows running pod count per namespace.
              The namespace query parameter is interpolated from the variable.
          plugin:
            kind: TimeSeriesChart
            spec:
              legend:
                position: bottom
                mode: table
              yAxis:
                format:
                  unit: decimal
          queries:
            - kind: TimeSeriesQuery
              spec:
                plugin:
                  kind: PrometheusTimeSeriesQuery
                  spec:
                    datasource:
                      kind: PrometheusDatasource
                      name: thanos-querier-dynamic-queryparams
                    query: |
                      sum by (namespace) (
                        kube_pod_status_phase{namespace=~"$namespace", phase="Running"}
                      )
                    seriesNameFormat: "{{namespace}}"
      container_cpu:
        kind: Panel
        spec:
          display:
            name: "Container CPU Usage"
            description: "CPU usage rate by pod in the selected namespace(s)"
          plugin:
            kind: TimeSeriesChart
            spec:
              legend:
                position: bottom
                mode: list
              yAxis:
                format:
                  unit: decimal
                  decimalPlaces: 3
          queries:
            - kind: TimeSeriesQuery
              spec:
                plugin:
                  kind: PrometheusTimeSeriesQuery
                  spec:
                    datasource:
                      kind: PrometheusDatasource
                      name: thanos-querier-dynamic-queryparams
                    query: |
                      topk(10,
                        sum by (namespace, pod) (
                          rate(container_cpu_usage_seconds_total{namespace=~"$namespace", container!=""}[5m])
                        )
                      )
                    seriesNameFormat: "{{namespace}}/{{pod}}"
      container_memory:
        kind: Panel
        spec:
          display:
            name: "Container Memory Working Set"
            description: "Memory working set by pod in the selected namespace(s)"
          plugin:
            kind: TimeSeriesChart
            spec:
              legend:
                position: bottom
                mode: list
              yAxis:
                format:
                  unit: bytes
                  shortValues: true
          queries:
            - kind: TimeSeriesQuery
              spec:
                plugin:
                  kind: PrometheusTimeSeriesQuery
                  spec:
                    datasource:
                      kind: PrometheusDatasource
                      name: thanos-querier-dynamic-queryparams
                    query: |
                      topk(10,
                        sum by (namespace, pod) (
                          container_memory_working_set_bytes{namespace=~"$namespace", container!=""}
                        )
                      )
                    seriesNameFormat: "{{namespace}}/{{pod}}"
      workload_replicas:
        kind: Panel
        spec:
          display:
            name: "Workload Replicas (Available vs Desired)"
            description: "Shows available vs desired replicas for the selected workload type"
          plugin:
            kind: TimeSeriesChart
            spec:
              legend:
                position: right
                mode: table
              yAxis:
                format:
                  unit: decimal
          queries:
            - kind: TimeSeriesQuery
              spec:
                plugin:
                  kind: PrometheusTimeSeriesQuery
                  spec:
                    datasource:
                      kind: PrometheusDatasource
                      name: thanos-querier-dynamic-queryparams
                    query: |
                      kube_deployment_status_replicas_available{namespace=~"$namespace"}
                    seriesNameFormat: "available: {{namespace}}/{{deployment}}"
            - kind: TimeSeriesQuery
              spec:
                plugin:
                  kind: PrometheusTimeSeriesQuery
                  spec:
                    datasource:
                      kind: PrometheusDatasource
                      name: thanos-querier-dynamic-queryparams
                    query: |
                      kube_deployment_spec_replicas{namespace=~"$namespace"}
                    seriesNameFormat: "desired: {{namespace}}/{{deployment}}"
      stat_pod_total:
        kind: Panel
        spec:
          display:
            name: "Total Running Pods"
            description: "Stat: total running pods across selected namespaces"
          plugin:
            kind: StatChart
            spec:
              calculation: last-number
              format:
                unit: decimal
          queries:
            - kind: TimeSeriesQuery
              spec:
                plugin:
                  kind: PrometheusTimeSeriesQuery
                  spec:
                    datasource:
                      kind: PrometheusDatasource
                      name: thanos-querier-dynamic-queryparams
                    query: |
                      sum(kube_pod_status_phase{namespace=~"$namespace", phase="Running"})
      stat_container_restarts:
        kind: Panel
        spec:
          display:
            name: "Container Restarts (1h)"
            description: "Total container restarts in the last hour"
          plugin:
            kind: StatChart
            spec:
              calculation: last-number
              format:
                unit: decimal
              sparkline: {}
          queries:
            - kind: TimeSeriesQuery
              spec:
                plugin:
                  kind: PrometheusTimeSeriesQuery
                  spec:
                    datasource:
                      kind: PrometheusDatasource
                      name: thanos-querier-dynamic-queryparams
                    query: |
                      sum(increase(kube_pod_container_status_restarts_total{namespace=~"$namespace"}[1h]))
      network_receive:
        kind: Panel
        spec:
          display:
            name: "Network Receive (bytes/s)"
            description: "Network bytes received per pod"
          plugin:
            kind: TimeSeriesChart
            spec:
              legend:
                position: bottom
              yAxis:
                format:
                  unit: decimal
                  shortValues: true
          queries:
            - kind: TimeSeriesQuery
              spec:
                plugin:
                  kind: PrometheusTimeSeriesQuery
                  spec:
                    datasource:
                      kind: PrometheusDatasource
                      name: thanos-querier-dynamic-queryparams
                    query: |
                      topk(10,
                        sum by (namespace, pod) (
                          rate(container_network_receive_bytes_total{namespace=~"$namespace"}[5m])
                        )
                      )
                    seriesNameFormat: "{{namespace}}/{{pod}}"
    layouts:
      - kind: Grid
        spec:
          display:
            title: Overview
          items:
            - x: 0
              "y": 0
              width: 4
              height: 4
              content:
                "$ref": "#/spec/panels/stat_pod_total"
            - x: 4
              "y": 0
              width: 4
              height: 4
              content:
                "$ref": "#/spec/panels/stat_container_restarts"
            - x: 8
              "y": 0
              width: 16
              height: 8
              content:
                "$ref": "#/spec/panels/pod_count"
      - kind: Grid
        spec:
          display:
            title: CPU & Memory
          items:
            - x: 0
              "y": 0
              width: 12
              height: 8
              content:
                "$ref": "#/spec/panels/container_cpu"
            - x: 12
              "y": 0
              width: 12
              height: 8
              content:
                "$ref": "#/spec/panels/container_memory"
      - kind: Grid
        spec:
          display:
            title: Workloads & Network
          items:
            - x: 0
              "y": 0
              width: 12
              height: 8
              content:
                "$ref": "#/spec/panels/workload_replicas"
            - x: 12
              "y": 0
              width: 12
              height: 8
              content:
                "$ref": "#/spec/panels/network_receive"
EOF

echo ""
echo "Done!"
