package kuma

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nickvdyck/terraform-provider-kuma/kumaclient"
)

func dataSourceKumaMesh() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMeshRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceMeshRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*kumaclient.KumaClient)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)

	mesh, err := c.GetMesh(name)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)

	if err := d.Set("name", mesh.Name); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
