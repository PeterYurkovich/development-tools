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

type tempoPluginSpec struct {
	Proxy httpProxy `json:"proxy"`
}

func EnsureTempoDatasource(ctx context.Context, kubeClient client.Client) (bool, error) {
	datasource := &persesv1alpha2.PersesGlobalDatasource{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.TempoDatasourceName,
		},
		Spec: persesv1alpha2.DatasourceSpec{
			Config: persesv1alpha2.Datasource{
				Spec: dsSpec.Spec{
					Default: true,
					Plugin: specCommon.Plugin{
						Kind: "TempoDatasource",
						Spec: tempoPluginSpec{
							Proxy: httpProxy{
								Kind: "HTTPProxy",
								Spec: httpProxySpec{
									URL: constants.TempoGatewayURL,
									Headers: map[string]string{
										"X-Scope-OrgID": "platform",
									},
									Secret: constants.TempoDatasourceSecretName,
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
		return false, fmt.Errorf("failed to create PersesGlobalDatasource %s: %w", constants.TempoDatasourceName, err)
	}
	return true, nil
}
