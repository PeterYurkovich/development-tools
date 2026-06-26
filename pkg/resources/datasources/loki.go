package datasources

import (
	"context"
	"fmt"

	specCommon "github.com/perses/spec/go/common"
	dsSpec "github.com/perses/spec/go/datasource"
	persesv1alpha2 "github.com/rhobs/perses-operator/api/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observability-ui/development-tools/internal/constants"
)

type lokiPluginSpec struct {
	Proxy httpProxy `json:"proxy"`
}

func EnsureLokiDatasource(ctx context.Context, kubeClient client.Client) (bool, error) {
	datasource := &persesv1alpha2.PersesGlobalDatasource{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.LokiDatasourceName,
		},
		Spec: persesv1alpha2.DatasourceSpec{
			Config: persesv1alpha2.Datasource{
				Spec: dsSpec.Spec{
					Default: true,
					Plugin: specCommon.Plugin{
						Kind: "LokiDatasource",
						Spec: lokiPluginSpec{
							Proxy: httpProxy{
								Kind: "HTTPProxy",
								Spec: httpProxySpec{
									URL: constants.LokiGatewayURL,
									Headers: map[string]string{
										"X-Scope-OrgID": "application",
									},
									Secret: constants.LokiDatasourceSecretName,
								},
							},
						},
					},
				},
			},
			Client: tlsClientConfig(),
		},
	}

	err := kubeClient.Create(ctx, datasource)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to create PersesGlobalDatasource %s: %w", constants.LokiDatasourceName, err)
	}
	return true, nil
}
