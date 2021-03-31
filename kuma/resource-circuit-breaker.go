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

func resourceCircuitBreaker() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCircuitBreakerCreate,
		ReadContext:   resourceCircuitBreakerRead,
		UpdateContext: resourceCircuitBreakerUpdate,
		DeleteContext: resourceCircuitBreakerDelete,
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
						"interval": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: diffDurationsCircuitBreaker,
						},
						"base_ejection_time": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: diffDurationsCircuitBreaker,
						},
						"max_ejection_percent": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"split_external_and_local_errors": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"detectors": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"total_errors": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"consecutive": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"gateway_errors": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"consecutive": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"local_errors": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"consecutive": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"standart_deviation": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"request_volume": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"minimun_hosts": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"factor": {
													Type:     schema.TypeFloat,
													Optional: true,
												},
											},
										},
									},
									"failure": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"request_volume": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"minimun_hosts": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"treshold": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
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
func diffDurationsCircuitBreaker(k, old string, new string, d *schema.ResourceData) bool {
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

func resourceCircuitBreakerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)

	circuitBreaker := createKumaCircuitBreakerFromResourceData(d)

	meshName := readStringFromResource(d, "mesh")
	name := readStringFromResource(d, "name")

	err := store.Create(ctx, &circuitBreaker, core_store.CreateByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(circuitBreaker.Meta.GetName())
	return resourceCircuitBreakerRead(ctx, d, m)
}

func resourceCircuitBreakerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	meshName := d.Get("mesh").(string)

	circuitBreaker := &mesh.CircuitBreakerResource{
		Spec: &mesh_proto.CircuitBreaker{},
	}

	err := store.Get(ctx, circuitBreaker, core_store.GetByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(circuitBreaker.Meta.GetName())

	if err := d.Set("name", circuitBreaker.Meta.GetName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("mesh", circuitBreaker.Meta.GetMesh()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("sources", flattenKumaSelector(circuitBreaker.Spec.Sources)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("destinations", flattenKumaSelector(circuitBreaker.Spec.Destinations)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("conf", flattenKumaCircuitBreakerConf(circuitBreaker.Spec.Conf)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCircuitBreakerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("sources") || d.HasChange("destinations") || d.HasChange("conf") {
		store := m.(core_store.ResourceStore)

		meshName := readStringFromResource(d, "mesh")
		name := readStringFromResource(d, "name")
		circuitBreaker := createKumaCircuitBreakerFromResourceData(d)

		oldCircuitBreaker := mesh.CircuitBreakerResource{
			Spec: &mesh_proto.CircuitBreaker{},
		}

		err := store.Get(ctx, &oldCircuitBreaker, core_store.GetByKey(name, meshName))
		if err != nil {
			return diag.FromErr(err)
		}

		circuitBreaker.Meta = oldCircuitBreaker.Meta

		err = store.Update(ctx, &circuitBreaker)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceCircuitBreakerRead(ctx, d, m)
}

func resourceCircuitBreakerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	meshName := d.Get("mesh").(string)

	circuitBreaker := createKumaCircuitBreakerFromResourceData(d)

	err := store.Delete(ctx, &circuitBreaker, core_store.DeleteByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createKumaCircuitBreakerFromResourceData(data *schema.ResourceData) mesh.CircuitBreakerResource {
	circuitBreaker := mesh.CircuitBreakerResource{
		Spec: &mesh_proto.CircuitBreaker{},
	}

	if attr, ok := data.GetOk("sources"); ok {
		sourcesArray := attr.(*schema.Set)

		if sourcesArray != nil && sourcesArray.Len() > 0 {
			circuitBreaker.Spec.Sources = createKumaSelectorFromArray(sourcesArray)
		}
	}

	if attr, ok := data.GetOk("destinations"); ok {
		destinationsArray := attr.(*schema.Set)

		if destinationsArray != nil && destinationsArray.Len() > 0 {
			circuitBreaker.Spec.Destinations = createKumaSelectorFromArray(destinationsArray)
		}
	}

	if attr, ok := data.GetOk("conf"); ok {

		if confArray := attr.([]interface{}); confArray != nil && len(confArray) > 0 {
			circuitBreaker.Spec.Conf = createKumaCircuitBreakerConfFromMap(confArray[0].(map[string]interface{}))
		}
	}

	return circuitBreaker
}

func createKumaCircuitBreakerConfFromMap(confMap map[string]interface{}) *mesh_proto.CircuitBreaker_Conf {
	conf := &mesh_proto.CircuitBreaker_Conf{}

	if interval, ok := confMap["interval"].(string); ok {
		duration, _ := readDurationFromString(interval)
		conf.Interval = duration
	}

	if baseEjectionTime, ok := confMap["base_ejection_time"].(string); ok {
		duration, _ := readDurationFromString(baseEjectionTime)
		conf.BaseEjectionTime = duration
	}

	if maxEjectionPercent, ok := confMap["max_ejection_persent"].(int); ok {
		conf.MaxEjectionPercent = &wrappers.UInt32Value{Value: uint32(maxEjectionPercent)}
	}

	if splitExternalAndLocalErrors, ok := confMap["split_external_and_local_errors"].(bool); ok {
		conf.SplitExternalAndLocalErrors = splitExternalAndLocalErrors
	}

	if detectors, ok := confMap["detectors"].([]interface{}); ok && len(detectors) > 0 {
		conf.Detectors = createKumaCircuitBreakerConfDetectorsFromMap(detectors[0].(map[string]interface{}))
	}

	return conf
}

func createKumaCircuitBreakerConfDetectorsFromMap(detectorsMap map[string]interface{}) *mesh_proto.CircuitBreaker_Conf_Detectors {
	detectors := &mesh_proto.CircuitBreaker_Conf_Detectors{}

	if totalErrorsSet, ok := detectorsMap["total_errors"].([]interface{}); ok && len(totalErrorsSet) > 0 {
		detectors.TotalErrors = createKumaRetryConfDetectorsErrorsFromMap(totalErrorsSet[0].(map[string]interface{}))
	}

	if gatewayErrorsSet, ok := detectorsMap["gateway_errors"].([]interface{}); ok && len(gatewayErrorsSet) > 0 {
		detectors.GatewayErrors = createKumaRetryConfDetectorsErrorsFromMap(gatewayErrorsSet[0].(map[string]interface{}))
	}

	if localErrorsSet, ok := detectorsMap["total_errors"].([]interface{}); ok && len(localErrorsSet) > 0 {
		detectors.LocalErrors = createKumaRetryConfDetectorsErrorsFromMap(localErrorsSet[0].(map[string]interface{}))
	}

	if standardDeviation, ok := detectorsMap["standard_deviation"].([]interface{}); ok && len(standardDeviation) > 0 {
		detectors.StandardDeviation = createKumaRetryConfDetectorsStandardDeviationFromMap(standardDeviation[0].(map[string]interface{}))
	}

	return detectors
}

func createKumaRetryConfDetectorsErrorsFromMap(errorsMap map[string]interface{}) *mesh_proto.CircuitBreaker_Conf_Detectors_Errors {
	errors := &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{}

	if errorVal, ok := errorsMap["consecutive"].(int); ok {
		errors.Consecutive = &wrappers.UInt32Value{Value: uint32(errorVal)}
	}

	return errors
}

func createKumaRetryConfDetectorsStandardDeviationFromMap(standardDeviationMap map[string]interface{}) *mesh_proto.CircuitBreaker_Conf_Detectors_StandardDeviation {
	standardDeviation := &mesh_proto.CircuitBreaker_Conf_Detectors_StandardDeviation{}

	if requestVolumeVal, ok := standardDeviationMap["request_volume"].(int); ok {
		standardDeviation.RequestVolume = &wrappers.UInt32Value{Value: uint32(requestVolumeVal)}
	}

	if minimunHostsVal, ok := standardDeviationMap["minimun_hosts"].(int); ok {
		standardDeviation.MinimumHosts = &wrappers.UInt32Value{Value: uint32(minimunHostsVal)}
	}

	if factorVal, ok := standardDeviationMap["factor"].(wrappers.DoubleValue); ok {
		standardDeviation.Factor = &factorVal
	}

	return standardDeviation
}

func createKumaRetryConfDetectorsFailureFromMap(failureMap map[string]interface{}) *mesh_proto.CircuitBreaker_Conf_Detectors_Failure {
	failure := &mesh_proto.CircuitBreaker_Conf_Detectors_Failure{}

	if requestVolumeVal, ok := failureMap["request_volume"].(int); ok {
		failure.RequestVolume = &wrappers.UInt32Value{Value: uint32(requestVolumeVal)}
	}

	if minimunHostsVal, ok := failureMap["minimun_hosts"].(int); ok {
		failure.MinimumHosts = &wrappers.UInt32Value{Value: uint32(minimunHostsVal)}
	}

	if factorVal, ok := failureMap["factor"].(int); ok {
		failure.Threshold = &wrappers.UInt32Value{Value: uint32(factorVal)}
	}

	return failure
}

func flattenKumaCircuitBreakerConf(conf *mesh_proto.CircuitBreaker_Conf) []interface{} {
	confMap := make(map[string]interface{})
	confSet := make([]interface{}, 0, 1)

	confSet = append(confSet, confMap)
	return confSet
}
