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

	retry := createKumaRetryFromResourceData(d)

	meshName := readStringFromResource(d, "mesh")
	name := readStringFromResource(d, "name")

	err := store.Create(ctx, &retry, core_store.CreateByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(retry.Meta.GetName())
	return resourceRetryRead(ctx, d, m)
}

func resourceRetryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	meshName := d.Get("mesh").(string)

	retry := &mesh.RetryResource{
		Spec: &mesh_proto.Retry{},
	}

	err := store.Get(ctx, retry, core_store.GetByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(retry.Meta.GetName())

	if err := d.Set("name", retry.Meta.GetName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("mesh", retry.Meta.GetMesh()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("sources", flattenKumaSelector(retry.Spec.Sources)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("destinations", flattenKumaSelector(retry.Spec.Destinations)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("conf", flattenKumaRetryConf(retry.Spec.Conf)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRetryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("sources") || d.HasChange("destinations") || d.HasChange("conf") {
		store := m.(core_store.ResourceStore)

		meshName := readStringFromResource(d, "mesh")
		name := readStringFromResource(d, "name")
		retry := createKumaRetryFromResourceData(d)

		oldRetry := mesh.RetryResource{
			Spec: &mesh_proto.Retry{},
		}

		err := store.Get(ctx, &oldRetry, core_store.GetByKey(name, meshName))
		if err != nil {
			return diag.FromErr(err)
		}

		retry.Meta = oldRetry.Meta

		err = store.Update(ctx, &retry)

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

	retry := createKumaRetryFromResourceData(d)

	err := store.Delete(ctx, &retry, core_store.DeleteByKey(name, meshName))

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

	if http, ok := confMap["http"].(map[string]interface{}); ok {
		conf.Http = createKumaRetryConfHTTPFromMap(http)
	}

	if grpc, ok := confMap["grpc"].(map[string]interface{}); ok {
		conf.Grpc = createKumaRetryConfGRPCFromMap(grpc)
	}

	if tcp, ok := confMap["tcp"].(map[string]interface{}); ok {
		conf.Tcp = createKumaRetryConfTCPFromMap(tcp)
	}

	return conf
}

func createKumaRetryConfHTTPFromMap(httpMap map[string]interface{}) *mesh_proto.Retry_Conf_Http {
	http := &mesh_proto.Retry_Conf_Http{}

	if numRetries, ok := httpMap["numRetries"].(uint32); ok {
		http.NumRetries = &wrappers.UInt32Value{Value: numRetries}
	}

	if interval, ok := httpMap["interval"].(string); ok {
		duration, _ := readDurationFromString(interval)
		http.PerTryTimeout = duration
	}

	if backOffMap, ok := httpMap["backOff"].(map[string]interface{}); ok {
		http.BackOff = createKumaRetryConfHTTPBackOffFromMap(backOffMap)
	}

	if retiableStatusCodes, ok := httpMap["expected_statuses"].([]uint32); ok {
		// uIntArray := []*wrappers.UInt32Value{}

		// for _, data := range retiableStatusCodes {
		// 	item := wrappers.UInt32Value{
		// 		Value: data,
		// 	}
		// 	uIntArray = append(uIntArray, &item)
		// }

		http.RetriableStatusCodes = retiableStatusCodes
	}

	return http

}

func createKumaRetryConfGRPCFromMap(grpcMap map[string]interface{}) *mesh_proto.Retry_Conf_Grpc {
	grpc := &mesh_proto.Retry_Conf_Grpc{}

	if numRetries, ok := grpcMap["numRetries"].(uint32); ok {
		grpc.NumRetries = &wrappers.UInt32Value{Value: numRetries}
	}

	if interval, ok := grpcMap["interval"].(string); ok {
		duration, _ := readDurationFromString(interval)
		grpc.PerTryTimeout = duration
	}

	if backOffMap, ok := grpcMap["backOff"].(map[string]interface{}); ok {
		grpc.BackOff = createKumaRetryConfHTTPBackOffFromMap(backOffMap)
	}

	if retiableStatusCodes, ok := grpcMap["expected_statuses"].([]mesh_proto.Retry_Conf_Grpc_RetryOn); ok {
		grpc.RetryOn = retiableStatusCodes
	}

	return grpc
}

func createKumaRetryConfTCPFromMap(tcpMap map[string]interface{}) *mesh_proto.Retry_Conf_Tcp {
	tcp := &mesh_proto.Retry_Conf_Tcp{}

	if maxConnectAttempts, ok := tcpMap["maxConnectAttempts"].(uint32); ok {
		tcp.MaxConnectAttempts = maxConnectAttempts
	}

	return tcp
}

func createKumaRetryConfHTTPBackOffFromMap(backOffMap map[string]interface{}) *mesh_proto.Retry_Conf_BackOff {
	backOff := &mesh_proto.Retry_Conf_BackOff{}

	if baseInterval, ok := backOffMap["baseInterval"].(string); ok {
		duration, _ := readDurationFromString(baseInterval)
		backOff.BaseInterval = duration
	}

	if maxInterval, ok := backOffMap["maxInterval"].(string); ok {
		duration, _ := readDurationFromString(maxInterval)
		backOff.MaxInterval = duration
	}

	return backOff

}

func flattenKumaRetryConf(conf *mesh_proto.Retry_Conf) map[string]interface{} {
	confMap := make(map[string]interface{})

	if conf == nil {
		return confMap
	}

	if conf.Http != nil {
		confMap["http"] = flattenKumaRetryConfHTTP(conf.Http)
	}

	if conf.Grpc != nil {
		confMap["grpc"] = flattenKumaRetryConfGRPC(conf.Grpc)
	}

	if conf.Tcp != nil {
		confMap["tcp"] = flattenKumaRetryConfTCP(conf.Tcp)
	}

	return confMap
}

func flattenKumaRetryConfHTTP(http *mesh_proto.Retry_Conf_Http) map[string]interface{} {
	httpMap := make(map[string]interface{})

	if http == nil {
		return httpMap
	}

	if http.NumRetries != nil {
		httpMap["numRetries"] = http.NumRetries
	}

	if http.PerTryTimeout != nil {
		httpMap["perTryTimeout"] = http.PerTryTimeout.String()
	}

	if http.BackOff != nil {
		httpMap["backOff"] = flattenKumaHealthCheckConfBackoff(http.BackOff)
	}

	if http.RetriableStatusCodes != nil {
		httpMap["retriableStatusCodes"] = http.RetriableStatusCodes

	}

	return httpMap

}

func flattenKumaRetryConfGRPC(grpc *mesh_proto.Retry_Conf_Grpc) map[string]interface{} {
	grpcMap := make(map[string]interface{})

	if grpc == nil {
		return grpcMap
	}

	if grpc.NumRetries != nil {
		grpcMap["numRetries"] = grpc.NumRetries
	}

	if grpc.PerTryTimeout != nil {
		grpcMap["perTryTimeout"] = grpc.PerTryTimeout.String()
	}

	if grpc.BackOff != nil {
		grpcMap["backOff"] = flattenKumaHealthCheckConfBackoff(grpc.BackOff)
	}

	if grpc.RetryOn != nil {
		grpcMap["retryOn"] = grpc.RetryOn

	}

	return grpcMap

}

func flattenKumaRetryConfTCP(tcp *mesh_proto.Retry_Conf_Tcp) map[string]interface{} {
	tcpMap := make(map[string]interface{})

	if tcp == nil {
		return tcpMap
	}
	// is this okey? Uint32 is not nilable
	if tcp.MaxConnectAttempts != 0 {
		tcpMap["maxConnectAttempts"] = tcp.MaxConnectAttempts
	}

	return tcpMap
}

func flattenKumaHealthCheckConfBackoff(backoff *mesh_proto.Retry_Conf_BackOff) map[string]interface{} {
	backOffMap := make(map[string]interface{})

	if backoff != nil {
		return backOffMap
	}

	if backoff.BaseInterval != nil {
		backOffMap["baseInterval"] = backoff.BaseInterval
	}

	if backoff.MaxInterval != nil {
		backOffMap["maxInterval"] = backoff.MaxInterval
	}

	return backOffMap
}
