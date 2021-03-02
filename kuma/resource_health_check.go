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

func resourceHealthCheck() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHealthCheckCreate,
		ReadContext:   resourceHealthCheckRead,
		UpdateContext: resourceHealthCheckUpdate,
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
						"interval": {
							Type: schema.TypeString,
						},
						"timeout": {
							Type: schema.TypeString,
						},
						"unhealthy_threshold": {
							Type: schema.TypeString,
						},
						"healthy_threshold": {
							Type: schema.TypeString,
						},
						"tcp": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"send": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"receive": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
						},
						"http": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"path": {
										Type:     schema.TypeString,
										Required: true,
									},
									"request_headers_to_add": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"append": {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  true,
												},
												"header": {
													Type:     schema.TypeMap,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"key": {
																Type:     schema.TypeString,
																Required: true,
															},
															"value": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"expected_statuses": {
										Type:     schema.TypeInt,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type:     schema.TypeString,
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
	}
}

func resourceHealthCheckCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)

	healthCheck := createKumaHealthCheckFromResourceData(d)

	meshName := readStringFromResource(d, "mesh")
	name := readStringFromResource(d, "name")

	err := store.Create(ctx, &healthCheck, core_store.CreateByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(healthCheck.Meta.GetName())
	return resourceHealthCheckRead(ctx, d, m)
}

func resourceHealthCheckRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	meshName := d.Get("mesh").(string)

	healthCheck := &mesh.HealthCheckResource{
		Spec: &mesh_proto.HealthCheck{},
	}

	err := store.Get(ctx, healthCheck, core_store.GetByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(healthCheck.Meta.GetName())

	if err := d.Set("name", healthCheck.Meta.GetName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("mesh", healthCheck.Meta.GetMesh()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("sources", flattenKumaSelector(healthCheck.Spec.Sources)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("destinations", flattenKumaSelector(healthCheck.Spec.Destinations)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("conf", flattenKumaHealthCheckConf(healthCheck.Spec.Conf)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceHealthCheckUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChanges("sources", "destinations", "conf") {
		store := m.(core_store.ResourceStore)

		name := d.Id()
		meshName := d.Get("mesh").(string)

		healthCheck := createKumaHealthCheckFromResourceData(d)

		oldHealthCheck := mesh.HealthCheckResource{
			Spec: &mesh_proto.HealthCheck{},
		}

		err := store.Get(ctx, &oldHealthCheck, core_store.GetByKey(name, meshName))

		if err != nil {
			return diag.FromErr(err)
		}
		healthCheck.Meta = oldHealthCheck.Meta

		err = store.Update(ctx, &healthCheck)

		if err != nil {
			return diag.FromErr((err))
		}
	}

	return resourceTrafficPermissionRead(ctx, d, m)
}

func createKumaHealthCheckFromResourceData(data *schema.ResourceData) mesh.HealthCheckResource {
	healthCheck := mesh.HealthCheckResource{
		Spec: &mesh_proto.HealthCheck{},
	}

	if attr, ok := data.GetOk("sources"); ok {
		sourcesArray := attr.(*schema.Set)

		if sourcesArray != nil && sourcesArray.Len() > 0 {
			healthCheck.Spec.Sources = createKumaSelectorFromArray(sourcesArray)
		}
	}

	if attr, ok := data.GetOk("destinations"); ok {
		destinationsArray := attr.(*schema.Set)

		if destinationsArray != nil && destinationsArray.Len() > 0 {
			healthCheck.Spec.Destinations = createKumaSelectorFromArray(destinationsArray)
		}
	}

	if attr, ok := data.GetOk("conf"); ok {
		confMap := attr.(map[string]interface{})

		if confMap != nil && len(confMap) > 0 {
			healthCheck.Spec.Conf = createKumaHealthCheckConfFromMap(confMap)
		}
	}

	return healthCheck
}

func createKumaHealthCheckConfFromMap(confMap map[string]interface{}) *mesh_proto.HealthCheck_Conf {
	conf := &mesh_proto.HealthCheck_Conf{}

	if interval, ok := confMap["interval"].(string); ok {
		duration, _ := readDurationFromString(interval)
		conf.Interval = duration
	}

	if timeout, ok := confMap["interval"].(string); ok {
		duration, _ := readDurationFromString(timeout)
		conf.Timeout = duration
	}

	if unhealthyThreshold, ok := confMap["unhealthy_threshold"].(uint32); ok {
		conf.UnhealthyThreshold = unhealthyThreshold
	}

	if healthyThreshold, ok := confMap["healthy_threshold"].(uint32); ok {
		conf.HealthyThreshold = healthyThreshold
	}

	if tcpMap, ok := confMap["tcp"].(map[string]interface{}); ok {
		conf.Protocol = createKumaHealthCheckConfTCPFromMap(tcpMap)
	}

	if httpMap, ok := confMap["http"].(map[string]interface{}); ok {
		conf.Protocol = createKumaHealthCheckConfHTTPFromMap(httpMap)
	}

	return conf
}

func createKumaHealthCheckConfTCPFromMap(tcpMap map[string]interface{}) *mesh_proto.HealthCheck_Conf_Tcp_ {
	tcp := &mesh_proto.HealthCheck_Conf_Tcp{}

	if send, ok := tcpMap["send"].(string); ok {
		tcp.Send = &wrappers.BytesValue{
			Value: []byte(send),
		}
	}

	if receive, ok := tcpMap["receive"].([]string); ok {
		bValue := []*wrappers.BytesValue{}

		for _, data := range receive {
			item := wrappers.BytesValue{
				Value: []byte(data),
			}
			bValue = append(bValue, &item)
		}

		tcp.Receive = bValue
	}

	return &mesh_proto.HealthCheck_Conf_Tcp_{
		Tcp: tcp,
	}
}

func createKumaHealthCheckConfHTTPFromMap(httpMap map[string]interface{}) *mesh_proto.HealthCheck_Conf_Http_ {
	http := &mesh_proto.HealthCheck_Conf_Http{
		RequestHeadersToAdd: []*mesh_proto.
			HealthCheck_Conf_Http_HeaderValueOption{
			{
				Header: &mesh_proto.HealthCheck_Conf_Http_HeaderValue{
					Key:   "foobar",
					Value: "foobaz",
				},
				Append: &wrappers.BoolValue{Value: false},
			},
		},
	}

	if path, ok := httpMap["path"].(string); ok {
		http.Path = path
	}

	if headers, ok := httpMap["request_headers_to_add"].([]map[string]interface{}); ok {
		headersArray := []*mesh_proto.HealthCheck_Conf_Http_HeaderValueOption{}
		for _, headerMap := range headers {
			header := createKumaHealthCheckConfRequestHeadersFromMap(headerMap)
			headersArray = append(headersArray, header)
		}
		http.RequestHeadersToAdd = headersArray
	}

	if statuses, ok := httpMap["expected_statuses"].([]uint32); ok {
		uIntArray := []*wrappers.UInt32Value{}

		for _, data := range statuses {
			item := wrappers.UInt32Value{
				Value: data,
			}
			uIntArray = append(uIntArray, &item)
		}

		http.ExpectedStatuses = uIntArray
	}

	return &mesh_proto.HealthCheck_Conf_Http_{
		Http: http,
	}
}

func createKumaHealthCheckConfRequestHeadersFromMap(headerMap map[string]interface{}) *mesh_proto.HealthCheck_Conf_Http_HeaderValueOption {
	header := &mesh_proto.HealthCheck_Conf_Http_HeaderValueOption{}

	if append, ok := headerMap["append"].(bool); ok {
		header.Append = &wrappers.BoolValue{Value: append}
	} else {
		header.Append = &wrappers.BoolValue{Value: true}
	}

	if headerValueMap, ok := headerMap["header"].(map[string]interface{}); ok {
		headerValue := &mesh_proto.HealthCheck_Conf_Http_HeaderValue{}

		if key, ok := headerValueMap["key"].(string); ok {
			headerValue.Key = key
		}

		if value, ok := headerValueMap["value"].(string); ok {
			headerValue.Value = value
		}
		header.Header = headerValue
	}

	return header
}

func flattenKumaHealthCheckConf(conf *mesh_proto.HealthCheck_Conf) map[string]interface{} {
	confMap := make(map[string]interface{})

	if conf == nil {
		return confMap
	}

	if conf.Interval != nil {
		confMap["interval"] = conf.Interval.String()
	}

	if conf.Timeout != nil {
		confMap["timeout"] = conf.Timeout.String()
	}

	confMap["unhealthy_threshold"] = conf.UnhealthyThreshold
	confMap["healthy_threshold"] = conf.HealthyThreshold

	protocol := conf.GetProtocol()

	if tcp, ok := protocol.(*mesh_proto.HealthCheck_Conf_Tcp_); ok {
		confMap["tcp"] = flattenKumaHealthCheckConfTCP(tcp)
	}

	if http, ok := protocol.(*mesh_proto.HealthCheck_Conf_Http_); ok {
		confMap["http"] = flattenKumaHealthCheckConfHTTP(http)
	}

	return confMap
}

func flattenKumaHealthCheckConfTCP(tcp *mesh_proto.HealthCheck_Conf_Tcp_) map[string]interface{} {
	tcpMap := make(map[string]interface{})

	if tcp.Tcp == nil {
		return tcpMap
	}

	if tcp.Tcp.Send != nil {
		tcpMap["send"] = string(tcp.Tcp.Send.String())
	}

	if tcp.Tcp.Receive != nil {
		receiveArray := make([]string, 0, len(tcp.Tcp.Receive))

		for _, byteValue := range tcp.Tcp.Receive {
			if byteValue != nil {
				receiveArray = append(receiveArray, byteValue.String())
			}
		}

		tcpMap["receive"] = receiveArray
	}

	return tcpMap
}

func flattenKumaHealthCheckConfHTTP(http *mesh_proto.HealthCheck_Conf_Http_) map[string]interface{} {
	httpMap := make(map[string]interface{})

	if http.Http == nil {
		return httpMap
	}

	httpMap["path"] = http.Http.Path

	if http.Http.RequestHeadersToAdd != nil {
		headersArray := make([]map[string]interface{}, 0, len(http.Http.RequestHeadersToAdd))

		for _, headerValueOption := range http.Http.RequestHeadersToAdd {
			headerValueOptionMap := flattenKumaHealthCheckConfHTTPHeaderOption(headerValueOption)

			if headerValueOptionMap != nil {
				headersArray = append(headersArray, headerValueOptionMap)
			}
		}
		httpMap["request_headers_to_add"] = headersArray
	}

	if http.Http.ExpectedStatuses != nil {
		statusesArray := make([]uint32, 0, len(http.Http.ExpectedStatuses))

		for _, statusValue := range http.Http.ExpectedStatuses {
			if statusValue != nil {
				statusesArray = append(statusesArray, statusValue.GetValue())
			}
		}
		httpMap["expected_statuse"] = statusesArray
	}

	return httpMap
}

func flattenKumaHealthCheckConfHTTPHeaderOption(headerOption *mesh_proto.HealthCheck_Conf_Http_HeaderValueOption) map[string]interface{} {
	if headerOption == nil {
		return nil
	}

	headerValueOptionMap := make(map[string]interface{})

	if headerOption.Append != nil {
		headerValueOptionMap["append"] = headerOption.Append.GetValue()
	}

	if headerOption.Header != nil {
		headerValueMap := make(map[string]interface{})

		headerValueMap["key"] = headerOption.Header.GetKey()
		headerValueMap["value"] = headerOption.Header.GetValue()

		headerValueOptionMap["header"] = headerValueMap
	}

	return headerValueOptionMap
}
