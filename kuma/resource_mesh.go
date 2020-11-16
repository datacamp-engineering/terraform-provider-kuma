package kuma

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nickvdyck/terraform-provider-kuma/kumaclient"
)

func resourceMesh() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMeshCreate,
		ReadContext:   resourceMeshRead,
		UpdateContext: resourceMeshUpdate,
		DeleteContext: resourceMeshDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceMeshCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*kumaclient.KumaClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	mesh := kumaclient.Mesh{}

	mesh.Name = name
	mesh.Type = "Mesh"

	err := c.UpsertMesh(mesh)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(mesh.Name)

	resourceMeshRead(ctx, d, m)

	return diags
}

func resourceMeshRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*kumaclient.KumaClient)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()

	mesh, err := c.GetMesh(name)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", mesh.Name); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceMeshUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// c := m.(*kumaclient.KumaClient)

	name := d.Id()

	// Actualy do an update here
	fmt.Printf("Do update for %s \n", name)

	return resourceMeshRead(ctx, d, m)
}

func resourceMeshDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*kumaclient.KumaClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()

	err := c.DeleteMeshByName(name)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
