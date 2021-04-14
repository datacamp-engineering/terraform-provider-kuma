package kuma

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUMA_HOST", nil),
			},
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUMA_API_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"kuma_traffic_permission": resourceTrafficPermission(),
			"kuma_retry":              resourceRetry(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiToken := d.Get("api_token").(string)

	var host string

	hostVal, ok := d.GetOk("host")
	if ok {
		host = hostVal.(string)
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if apiToken != "" {

		cp := &config_proto.ControlPlaneCoordinates_ApiServer{
			Url: host,
		}

		c, err := newResourceStore(cp)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Kuma client",
				Detail:   "Unable to authenticate user for authenticated Kuma client",
			})

			return nil, diags
		}

		return c, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create Kuma client",
		Detail:   "Unable to create anonymous Kuma client",
	})

	return nil, diags
}
