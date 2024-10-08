# development-tools

https://github.com/rhobs/observability-operator/blob/main/docs/user-guides/observability-ui-plugins.md

| __COO Version__ |   __OCP Versions__  | __Dashboards__ | __Distributed Tracing__ | __Logging__ | __Troubleshooting Panel__ |
| --------------- | ------------------- | -------------- | ----------------------- | ----------- | ------------------------- |
| 0.2.0           | 4.11                |       ✔        |             ✘           |       ✘     |             ✘             |
| 0.3.0+          | 4.11 - 4.15         |       ✔        |             ✔           |       ✔     |             ✘             |
| 0.3.0+          | 4.16+               |       ✔        |             ✔           |       ✔     |             ✔             |

### ROSA Clusters 
Requires a STS role 
'ManagedOpenShift'

https://docs.openshift.com/rosa/rosa_install_access_delete_clusters/rosa-sts-creating-a-cluster-quickly.html

### Prerequisites 
- OpenShift Cluster version 4.11+ 
- Logged into an OpenShift Cluster 
- [OpenShift CLI (oc)](https://docs.openshift.com/container-platform/4.16/cli_reference/openshift_cli/getting-started-cli.html) 
- [operator-sdk CLI](https://sdk.operatorframework.io/docs/installation/)
- Logged into an Openshift Cluster

### Makefile 
To pass your own image of the observability operator bundle use the flag `OPERATOR_BUNDLE`. For example: 

`make coo-resources OPERATOR_BUNDLE="quay.io/test/observability-operator-bundle"`