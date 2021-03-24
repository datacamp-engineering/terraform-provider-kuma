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
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"numretries": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"pertrytimeout": {
										Type:             schema.TypeString,
										Optional:         true,
										DiffSuppressFunc: diffDurations,
									},
									"backoff": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"baseinterval": {
													Type:             schema.TypeString,
													Required:         true,
													DiffSuppressFunc: diffDurations,
												},
												"maxinterval": {
													Type:             schema.TypeString,
													Required:         true,
													DiffSuppressFunc: diffDurations,
												},
											},
										},
									},
									"retriablestatuscodes": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
										Optional: true,
									},
								},
							},
						},
						"grpc": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"numretries": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"pertrytimeout": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"backoff": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"baseinterval": {
													Type:     schema.TypeString,
													Required: true,
												},
												"maxinterval": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"retryon": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
										Optional: true,
									},
								},
							},
						},
						"tcp": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"maxconnectattempts": {
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

// A hack to make sure changes in duration are picked up correctly
// This is needed because if we go from a duration to a string, it's not always
// possible to go back to the original string eg 1.5h => 1h30m0s.
// This is annoying because at that point we will store 1h30m0s into the state.
// When doing the diff again this get's picked up as a change.
func diffDurations(k, old string, new string, d *schema.ResourceData) bool {
	oldDuration, err := readDurationFromString(old)

	if err != nil {
		return false
	}

	newDuration, err := readDurationFromString(new)

	if err != nil {
		return false
	}

	return oldDuration.AsDuration().String() == newDuration.AsDuration().String()
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

		if confArray := attr.([]interface{}); confArray != nil && len(confArray) > 0 {
			retry.Spec.Conf = createKumaRetryConfFromMap(confArray[0].(map[string]interface{}))
		}
	}

	return retry
}

func createKumaRetryConfFromMap(confMap map[string]interface{}) *mesh_proto.Retry_Conf {
	conf := &mesh_proto.Retry_Conf{}

	if http, ok := confMap["http"].([]interface{}); ok && len(http) > 0 {
		conf.Http = createKumaRetryConfHTTPFromMap(http[0].(map[string]interface{}))
	}

	if grpc, ok := confMap["grpc"].([]interface{}); ok && len(grpc) > 0 {
		conf.Grpc = createKumaRetryConfGRPCFromMap(grpc[0].(map[string]interface{}))
	}

	if tcp, ok := confMap["tcp"].([]interface{}); ok && len(tcp) > 0 {
		conf.Tcp = createKumaRetryConfTCPFromMap(tcp[0].(map[string]interface{}))
	}

	return conf
}

func createKumaRetryConfHTTPFromMap(httpMap map[string]interface{}) *mesh_proto.Retry_Conf_Http {
	http := &mesh_proto.Retry_Conf_Http{}

	if numRetries, ok := httpMap["numretries"].(int); ok {
		http.NumRetries = &wrappers.UInt32Value{Value: uint32(numRetries)}
	}

	if interval, ok := httpMap["pertrytimeout"].(string); ok {
		duration, _ := readDurationFromString(interval)
		http.PerTryTimeout = duration
	}

	if backOffSet, ok := httpMap["backoff"].([]interface{}); ok && len(backOffSet) > 0 {
		http.BackOff = createKumaRetryConfHTTPBackOffFromMap(backOffSet[0].(map[string]interface{}))
	}

	if retriableStatusCodes, ok := httpMap["retriablestatuscodes"].([]interface{}); ok {
		uIntArray := make([]uint32, 0, len(http.RetriableStatusCodes))

		for _, data := range retriableStatusCodes {
			uIntArray = append(uIntArray, uint32(data.(int)))
		}

		http.RetriableStatusCodes = uIntArray
	}

	return http
}

func createKumaRetryConfGRPCFromMap(grpcMap map[string]interface{}) *mesh_proto.Retry_Conf_Grpc {
	grpc := &mesh_proto.Retry_Conf_Grpc{}

	if numRetries, ok := grpcMap["numretries"].(uint32); ok {
		grpc.NumRetries = &wrappers.UInt32Value{Value: numRetries}
	}

	if interval, ok := grpcMap["interval"].(string); ok {
		duration, _ := readDurationFromString(interval)
		grpc.PerTryTimeout = duration
	}

	if backOffMap, ok := grpcMap["backoff"].(map[string]interface{}); ok {
		grpc.BackOff = createKumaRetryConfHTTPBackOffFromMap(backOffMap)
	}

	if retiableStatusCodes, ok := grpcMap["expected_statuses"].([]mesh_proto.Retry_Conf_Grpc_RetryOn); ok {
		grpc.RetryOn = retiableStatusCodes
	}

	return grpc
}

func createKumaRetryConfTCPFromMap(tcpMap map[string]interface{}) *mesh_proto.Retry_Conf_Tcp {
	tcp := &mesh_proto.Retry_Conf_Tcp{}

	if maxConnectAttempts, ok := tcpMap["maxconnectattempts"].(uint32); ok {
		tcp.MaxConnectAttempts = maxConnectAttempts
	}

	return tcp
}

func createKumaRetryConfHTTPBackOffFromMap(backOffMap map[string]interface{}) *mesh_proto.Retry_Conf_BackOff {
	backOff := &mesh_proto.Retry_Conf_BackOff{}

	if baseInterval, ok := backOffMap["baseinterval"].(string); ok {
		duration, _ := readDurationFromString(baseInterval)
		backOff.BaseInterval = duration
	}

	if maxInterval, ok := backOffMap["maxinterval"].(string); ok {
		duration, _ := readDurationFromString(maxInterval)
		backOff.MaxInterval = duration
	}

	return backOff

}

func flattenKumaRetryConf(conf *mesh_proto.Retry_Conf) []interface{} {
	confMap := make(map[string]interface{})
	confSet := make([]interface{}, 0, 1)

	if conf == nil {
		return confSet
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
	confSet = append(confSet, confMap)
	return confSet
}

func flattenKumaRetryConfHTTP(http *mesh_proto.Retry_Conf_Http) []interface{} {
	httpMap := make(map[string]interface{})

	httpSet := make([]interface{}, 0, 1)

	if http == nil {
		return httpSet
	}

	if http.NumRetries != nil {
		httpMap["numretries"] = int(http.NumRetries.GetValue())
	}

	if http.PerTryTimeout != nil {
		httpMap["pertrytimeout"] = http.PerTryTimeout.AsDuration().String()
	}

	if http.BackOff != nil {
		httpMap["backoff"] = flattenKumaRetryConfBackoff(http.BackOff)
	}

	if http.RetriableStatusCodes != nil {
		intArray := make([]int, 0, len(http.RetriableStatusCodes))

		for _, data := range http.RetriableStatusCodes {
			intArray = append(intArray, int(data))
		}
		httpMap["retriablestatuscodes"] = intArray
	}

	httpSet = append(httpSet, httpMap)
	return httpSet

}

func flattenKumaRetryConfGRPC(grpc *mesh_proto.Retry_Conf_Grpc) []interface{} {
	grpcMap := make(map[string]interface{})
	grpcSet := make([]interface{}, 0, 1)
	if grpc == nil {
		return grpcSet
	}

	if grpc.NumRetries != nil {
		grpcMap["numretries"] = int(grpc.NumRetries.GetValue())
	}

	if grpc.PerTryTimeout != nil {
		grpcMap["pertrytimeout"] = grpc.PerTryTimeout.String()
	}

	if grpc.BackOff != nil {
		grpcMap["backoff"] = flattenKumaRetryConfBackoff(grpc.BackOff)
	}

	if grpc.RetryOn != nil {
		grpcMap["retryon"] = grpc.RetryOn

	}

	grpcSet = append(grpcSet, grpcMap)
	return grpcSet

}

func flattenKumaRetryConfTCP(tcp *mesh_proto.Retry_Conf_Tcp) []interface{} {
	tcpMap := make(map[string]interface{})
	tcpSet := make([]interface{}, 1)

	if tcp == nil {
		return tcpSet
	}
	// is this okey? Uint32 is not nilable
	if tcp.MaxConnectAttempts != 0 {
		tcpMap["maxconnectattempts"] = tcp.MaxConnectAttempts
	}
	tcpSet = append(tcpSet, tcpMap)
	return tcpSet
}

func flattenKumaRetryConfBackoff(backoff *mesh_proto.Retry_Conf_BackOff) []interface{} {
	backOffMap := make(map[string]interface{})
	backOffSet := make([]interface{}, 0, 1)

	if backoff == nil {
		return backOffSet
	}

	if backoff.BaseInterval != nil {
		backOffMap["baseinterval"] = backoff.BaseInterval.AsDuration().String()
	}

	if backoff.MaxInterval != nil {
		backOffMap["maxinterval"] = backoff.MaxInterval.AsDuration().String()
	}

	backOffSet = append(backOffSet, backOffMap)
	return backOffSet
}
