package kuma

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func resourceTrafficPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTrafficPermissionCreate,
		ReadContext:   resourceTrafficPermissionRead,
		UpdateContext: resourceTrafficPermissionUpdate,
		DeleteContext: resourceTrafficPermissionDelete,
		Schema: map[string]*schema.Schema{
			"mesh": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"sources": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"match": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
			"destinations": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"match": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func resourceTrafficPermissionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)

	permission := createKumaTrafficPermissionFromResourceData(d)

	meshName := readStringFromResource(d, "mesh")
	name := readStringFromResource(d, "name")

	err := store.Create(ctx, &permission, core_store.CreateByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(permission.Meta.GetName())
	return resourceTrafficPermissionRead(ctx, d, m)
}

func resourceTrafficPermissionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	meshName := d.Get("mesh").(string)

	permission := &mesh.TrafficPermissionResource{
		Spec: &mesh_proto.TrafficPermission{},
	}

	err := store.Get(ctx, permission, core_store.GetByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(permission.Meta.GetName())

	if err := d.Set("name", permission.Meta.GetName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("mesh", permission.Meta.GetMesh()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("sources", flattenKumaSelector(permission.Spec.Sources)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("destinations", flattenKumaSelector(permission.Spec.Destinations)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTrafficPermissionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("sources") || d.HasChange("destinations") {
		store := m.(core_store.ResourceStore)

		meshName := readStringFromResource(d, "mesh")
		name := readStringFromResource(d, "name")
		permission := createKumaTrafficPermissionFromResourceData(d)

		oldPermission := mesh.TrafficPermissionResource{
			Spec: &mesh_proto.TrafficPermission{},
		}

		err := store.Get(ctx, &oldPermission, core_store.GetByKey(name, meshName))
		if err != nil {
			return diag.FromErr(err)
		}

		permission.Meta = oldPermission.Meta

		err = store.Update(ctx, &permission)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceTrafficPermissionRead(ctx, d, m)
}

func resourceTrafficPermissionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	meshName := d.Get("mesh").(string)

	permission := createKumaTrafficPermissionFromResourceData(d)

	err := store.Delete(ctx, &permission, core_store.DeleteByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createKumaTrafficPermissionFromResourceData(data *schema.ResourceData) mesh.TrafficPermissionResource {
	permission := mesh.TrafficPermissionResource{
		Spec: &mesh_proto.TrafficPermission{},
	}

	if attr, ok := data.GetOk("sources"); ok {
		sourcesArray := attr.(*schema.Set)

		if sourcesArray != nil && sourcesArray.Len() > 0 {
			permission.Spec.Sources = createKumaSelectorFromArray(sourcesArray)
		}
	}

	if attr, ok := data.GetOk("destinations"); ok {
		destinationsArray := attr.(*schema.Set)

		if destinationsArray != nil && destinationsArray.Len() > 0 {
			permission.Spec.Destinations = createKumaSelectorFromArray(destinationsArray)
		}
	}

	return permission
}

func createKumaSelectorFromArray(data *schema.Set) []*mesh_proto.Selector {
	dataList := data.List()
	selector := []*mesh_proto.Selector{}

	for _, selectorData := range dataList {
		log.Printf("[DEBUG] kumaSelectorFromArray: %s", selectorData)
		selectorMap := selectorData.(map[string]interface{})
		item := mesh_proto.Selector{}

		if selectorMap["match"] != nil && len(selectorMap["match"].(map[string]interface{})) > 0 {
			tags := selectorMap["match"].(map[string]interface{})
			mapString := make(map[string]string)

			for key, value := range tags {
				mapString[key] = fmt.Sprintf("%v", value)
			}

			item.Match = mapString
			selector = append(selector, &item)
		}
	}

	return selector
}

func flattenKumaSelector(selectors []*mesh_proto.Selector) []interface{} {
	if selectors == nil {
		return []interface{}{}
	}

	flatList := make([]interface{}, 0, len(selectors))

	for _, item := range selectors {
		s := make(map[string]interface{})

		s["match"] = item.Match
		flatList = append(flatList, s)
	}

	return flatList
}
