package kuma

import (
	"context"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func resourceRetry() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRetryCreate,
		ReadContext:   resourceRetryRead,
		UpdateContext: resourceRetryUpdate,
		DeleteContext: resourceRetryDelete,
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
			"conf": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"numRetries": {
										Type:     schema.TypeInt,
										Required: false,
									},
									"perTryTimeout": {
										Type:     schema.TypeString,
										Required: false,
									},
									"backOff": {
										Type:     schema.TypeMap,
										Required: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"baseInterval": {
													Type:     schema.TypeString,
													Required: true,
												},
												"maxInterval": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"retriableStatusCodes": {
										Type:     schema.TypeList,
										Required: false,
									},
								},
							},
						},
						"grpc": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"numRetries": {
										Type:     schema.TypeInt,
										Required: false,
									},
									"perTryTimeout": {
										Type:     schema.TypeString,
										Required: false,
									},
									"backOff": {
										Type:     schema.TypeMap,
										Required: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"baseInterval": {
													Type:     schema.TypeString,
													Required: true,
												},
												"maxInterval": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"retryOn": {
										Type:     schema.TypeList,
										Required: false,
									},
								},
							},
						},
						"tcp": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"maxConnectAttempts": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceRetryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)

	permission := createKumaRetryFromResourceData(d)

	meshName := readStringFromResource(d, "mesh")
	name := readStringFromResource(d, "name")

	err := store.Create(ctx, &permission, core_store.CreateByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(permission.Meta.GetName())
	return resourceRetryRead(ctx, d, m)
}

func resourceRetryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	meshName := d.Get("mesh").(string)

	permission := &mesh.RetryResource{
		Spec: &mesh_proto.Retry{},
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

func resourceRetryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("sources") || d.HasChange("destinations") {
		store := m.(core_store.ResourceStore)

		meshName := readStringFromResource(d, "mesh")
		name := readStringFromResource(d, "name")
		permission := createKumaRetryFromResourceData(d)

		oldPermission := mesh.RetryResource{
			Spec: &mesh_proto.Retry{},
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

	return resourceRetryRead(ctx, d, m)
}

func resourceRetryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	meshName := d.Get("mesh").(string)

	permission := createKumaRetryFromResourceData(d)

	err := store.Delete(ctx, &permission, core_store.DeleteByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createKumaRetryFromResourceData(data *schema.ResourceData) mesh.RetryResource {
	retry := mesh.RetryResource{
		Spec: &mesh_proto.Retry{},
	}

	if attr, ok := data.GetOk("sources"); ok {
		sourcesArray := attr.(*schema.Set)

		if sourcesArray != nil && sourcesArray.Len() > 0 {
			retry.Spec.Sources = createKumaSelectorFromArray(sourcesArray)
		}
	}

	if attr, ok := data.GetOk("destinations"); ok {
		destinationsArray := attr.(*schema.Set)

		if destinationsArray != nil && destinationsArray.Len() > 0 {
			retry.Spec.Destinations = createKumaSelectorFromArray(destinationsArray)
		}
	}

	if attr, ok := data.GetOk("conf"); ok {
		confMap := attr.(map[string]interface{})

		if confMap != nil && len(confMap) > 0 {
			retry.Spec.Conf = createKumaRetryConfFromMap(confMap)
		}
	}

	return retry
}

func createKumaRetryConfFromMap(confMap map[string]interface{}) *mesh_proto.Retry_Conf {
	conf := &mesh_proto.Retry_Conf{}
	if httpMap, ok := confMap["http"].(map[string]interface{}); ok {
		conf.Http = createKumaRetryConfHTTPFromMap(httpMap)
	}

	return conf

}

func createKumaRetryConfHTTPFromMap(httpMap map[string]interface{}) *mesh_proto.Retry_Conf_Http {
	http := &mesh_proto.Retry_Conf_Http{}
	if numRetries, ok := httpMap["numRetries"].(uint32); ok {
		http.NumRetries = &wrappers.UInt32Value{Value: numRetries}
	}

	return http

}
