module github.com/nickvdyck/terraform-provider-kuma

go 1.16

require (
	github.com/golang/protobuf v1.4.3
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.2.0
	github.com/kumahq/kuma v0.0.0-20210223093449-638b0a5fcc1f
	github.com/kumahq/kuma/api v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.25.0
)

replace (
	github.com/kumahq/kuma/api => github.com/kumahq/kuma/api v0.0.0-20210223093449-638b0a5fcc1f
	github.com/kumahq/kuma/pkg/plugins/resources/k8s/native => github.com/kumahq/kuma/pkg/plugins/resources/k8s/native v0.0.0-20210223093449-638b0a5fcc1f
	github.com/kumahq/kuma/pkg/transparentproxy/istio => github.com/kumahq/kuma/pkg/transparentproxy/istio v0.0.0-20210223093449-638b0a5fcc1f

	github.com/prometheus/prometheus => github.com/kumahq/kuma/vendored/github.com/prometheus/prometheus v0.0.0-20210223093449-638b0a5fcc1f
	k8s.io/client-go => k8s.io/client-go v0.18.14
)
