#!/bin/bash
# Install Cluster Observability Operator (COO) from OperatorHub (redhat-operators).
# Minimal variant of z_coo.sh — installs only the operator itself.
# No UIPlugin, no PersesGlobalDatasource, no perses-dev project, no Scheduler patch.
#
# OperatorGroup must use AllNamespaces install mode — the COO CSV only supports AllNamespaces (not OwnNamespace).
# Use spec.targetNamespaces: [] (empty); a single entry listing COO_NS triggers UnsupportedOperatorGroup.
#
# Env (optional):
#   KUBECONFIG_PATH
#   COO_NS — target namespace (default openshift-cluster-observability-operator)
#   COO_PACKAGE — PackageManifest name (default cluster-observability-operator; override if your catalog differs)
#   COO_SUBSCRIPTION — Subscription metadata.name (default same as COO_PACKAGE)
#   COO_CHANNEL — default "auto" = pick from PackageManifest .status.channels[*].name (prefer stable-*), else defaultChannel
#   OPERATOR_INSTALL_TIMEOUT (default 900), ROLLOUT_TIMEOUT (default 600), OLM_POLL_INTERVAL (default 5)
#
# Uninstall: z_coo_uninstall.sh removes OLM resources in COO_NS.
#
# Uses subscriptions.operators.coreos.com (not oc delete subscription — ACM API conflict).

set -euo pipefail

if [[ -n "${KUBECONFIG_PATH:-}" ]]; then
  OC_KC=(oc --kubeconfig "${KUBECONFIG_PATH}")
else
  OC_KC=(oc)
fi

COO_NS="${COO_NS:-openshift-cluster-observability-operator}"
COO_PACKAGE="${COO_PACKAGE:-cluster-observability-operator}"
COO_SUBSCRIPTION="${COO_SUBSCRIPTION:-${COO_PACKAGE}}"
COO_CHANNEL="${COO_CHANNEL:-auto}"
OPERATOR_INSTALL_TIMEOUT="${OPERATOR_INSTALL_TIMEOUT:-900}"
ROLLOUT_TIMEOUT="${ROLLOUT_TIMEOUT:-600}"
OLM_POLL_INTERVAL="${OLM_POLL_INTERVAL:-5}"

resolve_channel_from_catalog() {
  local pkg=$1
  local fallback=${2:-stable}
  local ch
  ch=$("${OC_KC[@]}" get packagemanifest "$pkg" -n openshift-marketplace -o jsonpath='{.status.defaultChannel}' 2>/dev/null || true)
  if [[ -n "$ch" ]]; then
    echo "$ch"
  else
    echo "$fallback"
  fi
}

resolve_coo_channel_from_packagemanifest_channels() {
  local raw first best
  raw=$("${OC_KC[@]}" get packagemanifest "$COO_PACKAGE" -n openshift-marketplace -o jsonpath='{.status.channels[*].name}' 2>/dev/null || true)
  raw=${raw//$'\r'/}
  [[ -z "${raw// }" ]] && {
    resolve_channel_from_catalog "$COO_PACKAGE"
    return
  }
  read -r first _ <<<"$raw"

  best=$(echo "$raw" | tr ' ' '\n' | grep -E '^stable-[0-9]' | sort -V | tail -1 || true)
  if [[ -n "$best" ]]; then
    echo "$best"
    return
  fi
  if echo "$raw" | tr ' ' '\n' | grep -qx 'stable'; then
    echo 'stable'
    return
  fi
  best=$(echo "$raw" | tr ' ' '\n' | grep -E '^stable' | sort -V | tail -1 || true)
  if [[ -n "$best" ]]; then
    echo "$best"
    return
  fi
  if [[ -n "$first" ]]; then
    echo "$first"
    return
  fi
  resolve_channel_from_catalog "$COO_PACKAGE"
}

if [[ "${COO_CHANNEL}" == "auto" ]]; then
  COO_CHANNEL=$(resolve_coo_channel_from_packagemanifest_channels)
  echo "cluster-observability-operator Subscription channel: ${COO_CHANNEL} (PackageManifest ${COO_PACKAGE} .status.channels / fallback defaultChannel)"
fi

repair_orphaned_subscription() {
  local ns=$1
  local sub_name=$2
  local installplan_label_key=$3
  if ! "${OC_KC[@]}" get subscriptions.operators.coreos.com "$sub_name" -n "$ns" &>/dev/null; then
    return 0
  fi
  local installed
  installed=$("${OC_KC[@]}" get subscriptions.operators.coreos.com "$sub_name" -n "$ns" -o jsonpath='{.status.installedCSV}' 2>/dev/null || true)
  [[ -z "$installed" ]] && return 0
  if "${OC_KC[@]}" get csv "$installed" -n "$ns" &>/dev/null; then
    return 0
  fi
  echo "OLM repair: ${sub_name} (${ns}) references missing CSV ${installed}; deleting Subscription and InstallPlans"
  "${OC_KC[@]}" delete subscriptions.operators.coreos.com "$sub_name" -n "$ns" --ignore-not-found --wait=true --timeout=120s 2>/dev/null || true
  if [[ -n "${installplan_label_key}" ]]; then
    "${OC_KC[@]}" delete installplan -n "$ns" -l "${installplan_label_key}" --ignore-not-found 2>/dev/null || true
  fi
  sleep 3
}

wait_csv_from_subscription() {
  local ns=$1
  local sub_name=$2
  local deadline=$(( $(date +%s) + OPERATOR_INSTALL_TIMEOUT ))
  local last_log=$(( $(date +%s) ))
  echo "Waiting for Subscription ${sub_name} CSV Succeeded in ${ns}"
  while (( $(date +%s) < deadline )); do
    local csv_name phase
    csv_name=$("${OC_KC[@]}" get subscriptions.operators.coreos.com "$sub_name" -n "$ns" -o jsonpath='{.status.currentCSV}' 2>/dev/null || true)
    if [[ -n "$csv_name" ]] && "${OC_KC[@]}" get csv "$csv_name" -n "$ns" &>/dev/null; then
      phase=$("${OC_KC[@]}" get csv "$csv_name" -n "$ns" -o jsonpath='{.status.phase}' 2>/dev/null || true)
      if [[ "$phase" == "Succeeded" ]]; then
        echo "CSV ready: ${csv_name}"
        return 0
      fi
      if [[ "$phase" == "Failed" ]]; then
        echo "CSV ${csv_name} phase=Failed in ${ns}" >&2
        "${OC_KC[@]}" describe csv "$csv_name" -n "$ns" 2>/dev/null | tail -50 >&2 || true
        return 1
      fi
    fi
    local now=$(( $(date +%s) ))
    if (( now - last_log >= 60 )); then
      last_log=$now
      echo "... still waiting (${sub_name}): currentCSV=${csv_name:-<none>} phase=${phase:-n/a}"
    fi
    sleep "${OLM_POLL_INTERVAL}"
  done
  echo "Timeout waiting for Subscription ${sub_name} CSV in ${ns}" >&2
  "${OC_KC[@]}" get subscriptions.operators.coreos.com "$sub_name" -n "$ns" -o yaml 2>/dev/null | tail -40 >&2 || true
  "${OC_KC[@]}" get csv -n "$ns" -o wide 2>/dev/null || true
  return 1
}

wait_coo_operator_rollout() {
  echo "Waiting for Cluster Observability Operator workload in ${COO_NS}"
  local deadline=$(( $(date +%s) + ROLLOUT_TIMEOUT ))
  local dep
  while (( $(date +%s) < deadline )); do
    dep=$("${OC_KC[@]}" get deploy -n "$COO_NS" -o name 2>/dev/null | grep -E 'observability-operator|cluster-observability' | head -1 || true)
    if [[ -n "$dep" ]]; then
      "${OC_KC[@]}" rollout status "$dep" -n "$COO_NS" --timeout="${ROLLOUT_TIMEOUT}s"
      return 0
    fi
    if "${OC_KC[@]}" get pods -n "$COO_NS" -o name 2>/dev/null | grep -qi observability; then
      local pn
      pn=$("${OC_KC[@]}" get pods -n "$COO_NS" -o name 2>/dev/null | grep -i observability | head -1 | tr -d '\r')
      [[ -n "$pn" ]] && "${OC_KC[@]}" wait --for=condition=Ready "$pn" -n "$COO_NS" --timeout="${ROLLOUT_TIMEOUT}s" && return 0
    fi
    sleep 10
  done
  echo "Timeout: COO deployment/pod not found in ${COO_NS}" >&2
  "${OC_KC[@]}" get deploy,pods -n "$COO_NS" 2>/dev/null || true
  return 1
}

COO_IP_LABEL="operators.coreos.com/${COO_SUBSCRIPTION}.${COO_NS}="

echo "=== Apply Project ${COO_NS} + OperatorGroup + Subscription (source=redhat-operators, package=${COO_PACKAGE}) ==="
repair_orphaned_subscription "$COO_NS" "$COO_SUBSCRIPTION" "$COO_IP_LABEL"

if ! "${OC_KC[@]}" get project "${COO_NS}" &>/dev/null && ! "${OC_KC[@]}" get namespace "${COO_NS}" &>/dev/null; then
  "${OC_KC[@]}" apply -f - <<EOF
apiVersion: project.openshift.io/v1
kind: Project
metadata:
  name: ${COO_NS}
  labels:
    openshift.io/cluster-monitoring: "true"
EOF
else
  echo "Project/namespace ${COO_NS} already exists; ensuring openshift.io/cluster-monitoring label only"
  "${OC_KC[@]}" label namespace "${COO_NS}" openshift.io/cluster-monitoring=true --overwrite 2>/dev/null \
    || "${OC_KC[@]}" label project "${COO_NS}" openshift.io/cluster-monitoring=true --overwrite 2>/dev/null \
    || true
fi

cat <<EOF | "${OC_KC[@]}" apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: ${COO_NS}
  namespace: ${COO_NS}
spec:
  targetNamespaces: []
  upgradeStrategy: Default
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: ${COO_SUBSCRIPTION}
  namespace: ${COO_NS}
spec:
  channel: ${COO_CHANNEL}
  installPlanApproval: Automatic
  name: ${COO_PACKAGE}
  source: redhat-operators
  sourceNamespace: openshift-marketplace
EOF

while IFS= read -r _csv; do
  [[ -z "${_csv}" ]] && continue
  _base=${_csv##*/}
  _ph=$("${OC_KC[@]}" get csv "${_base}" -n "${COO_NS}" -o jsonpath='{.status.phase}' 2>/dev/null || true)
  if [[ "${_ph}" == "Failed" ]]; then
    echo "Removing failed CSV ${_base} (retry after OperatorGroup AllNamespaces fix)"
    "${OC_KC[@]}" delete csv "${_base}" -n "${COO_NS}" --ignore-not-found --wait=false 2>/dev/null || true
  fi
done < <("${OC_KC[@]}" get csv -n "${COO_NS}" -o name 2>/dev/null || true)

echo "=== Wait for Cluster Observability Operator CSV ==="
wait_csv_from_subscription "$COO_NS" "$COO_SUBSCRIPTION"

echo "=== Wait for operator Deployment rollout ==="
wait_coo_operator_rollout

echo "Done (z_coo_only.sh). COO installed in namespace: ${COO_NS}."
