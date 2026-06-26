package datasources

import (
	persesv1alpha2 "github.com/rhobs/perses-operator/api/v1alpha2"
)

type httpProxy struct {
	Kind string        `json:"kind"`
	Spec httpProxySpec `json:"spec"`
}

type httpProxySpec struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Secret  string            `json:"secret,omitempty"`
}

func tlsClientConfig() *persesv1alpha2.Client {
	enabled := true
	return &persesv1alpha2.Client{
		TLS: &persesv1alpha2.TLS{
			Enable: &enabled,
			CaCert: &persesv1alpha2.Certificate{
				SecretSource: persesv1alpha2.SecretSource{
					Type: persesv1alpha2.SecretSourceTypeFile,
				},
				CertPath: "/ca/service-ca.crt",
			},
		},
	}
}
