package kuma

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func resourceProxyTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProxyTemplateCreate,
		ReadContext:   resourceProxyTemplateRead,
		UpdateContext: resourceProxyTemplateUpdate,
		DeleteContext: resourceProxyTemplateDelete,
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
			"selectors": {
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
						"imports": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"modifications": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cluster": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"operation": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"value": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"match": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"origin": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"listener": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"operation": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"value": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"match": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"origin": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"network_filter": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"operation": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"value": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"match": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"origin": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"listener_name": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"http_filters": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"operation": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"value": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"match": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"origin": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"listener_name": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"virtual_host": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"operation": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"value": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"match": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"origin": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"route_configuration_name": {
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
func diffDurationsProxyTemplate(k, old string, new string, d *schema.ResourceData) bool {
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

func resourceProxyTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)

	proxy := createKumaProxyTemplateFromResourceData(d)

	meshName := readStringFromResource(d, "mesh")
	name := readStringFromResource(d, "name")

	err := store.Create(ctx, &proxy, core_store.CreateByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(proxy.Meta.GetName())
	return resourceProxyTemplateRead(ctx, d, m)
}

func resourceProxyTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	meshName := d.Get("mesh").(string)

	proxy := &mesh.ProxyTemplateResource{
		Spec: &mesh_proto.ProxyTemplate{},
	}

	err := store.Get(ctx, proxy, core_store.GetByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(proxy.Meta.GetName())

	if err := d.Set("name", proxy.Meta.GetName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("mesh", proxy.Meta.GetMesh()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("selectors", flattenKumaSelector(proxy.Spec.Selectors)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("conf", flattenKumaProxyTemplateConf(proxy.Spec.Conf)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceProxyTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("selector") || d.HasChange("conf") {
		store := m.(core_store.ResourceStore)

		meshName := readStringFromResource(d, "mesh")
		name := readStringFromResource(d, "name")
		proxy := createKumaProxyTemplateFromResourceData(d)

		oldProxyTemplate := mesh.ProxyTemplateResource{
			Spec: &mesh_proto.ProxyTemplate{},
		}

		err := store.Get(ctx, &oldProxyTemplate, core_store.GetByKey(name, meshName))
		if err != nil {
			return diag.FromErr(err)
		}

		proxy.Meta = oldProxyTemplate.Meta

		err = store.Update(ctx, &proxy)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceProxyTemplateRead(ctx, d, m)
}

func resourceProxyTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	store := m.(core_store.ResourceStore)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	meshName := d.Get("mesh").(string)

	proxy := createKumaProxyTemplateFromResourceData(d)

	err := store.Delete(ctx, &proxy, core_store.DeleteByKey(name, meshName))

	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createKumaProxyTemplateFromResourceData(data *schema.ResourceData) mesh.ProxyTemplateResource {
	proxy := mesh.ProxyTemplateResource{
		Spec: &mesh_proto.ProxyTemplate{},
	}

	if attr, ok := data.GetOk("selectors"); ok {
		sourcesArray := attr.(*schema.Set)

		if sourcesArray != nil && sourcesArray.Len() > 0 {
			proxy.Spec.Selectors = createKumaSelectorFromArray(sourcesArray)
		}
	}

	if attr, ok := data.GetOk("conf"); ok {

		if confArray := attr.([]interface{}); confArray != nil && len(confArray) > 0 {
			proxy.Spec.Conf = createKumaProxyTemplateConfFromMap(confArray[0].(map[string]interface{}))
		}
	}

	return proxy
}

func createKumaProxyTemplateConfFromMap(confMap map[string]interface{}) *mesh_proto.ProxyTemplate_Conf {
	conf := &mesh_proto.ProxyTemplate_Conf{}

	if imports, ok := confMap["imports"].([]interface{}); ok && len(imports) > 0 {
		importsStringArray := make([]string, 0, len(imports))
		for _, val := range imports {
			importsStringArray = append(importsStringArray, val.(string))
		}
		conf.Imports = importsStringArray
	}

	if modifications, ok := confMap["modifications"].([]interface{}); ok && len(modifications) > 0 {
		conf.Modifications = createKumaProxyTemplateConfModificationsFromMap(modifications[0].(map[string]interface{}))
	}

	return conf
}

// func createKumaProxyTemplateConfImportsFromMap(confmap map[string]interface{}) []string {
// 	StringArray := make([]string, 0, 0)

// 	if imports, ok := confmap["imports"].([]interface{}); ok {

// 		for _, data := range imports {
// 			StringArray = append(StringArray, fmt.Sprintf("%v", data))
// 		}

// 	}

// 	return StringArray
// }

func createKumaProxyTemplateConfModificationsFromMap(modMap map[string]interface{}) []*mesh_proto.ProxyTemplate_Modifications {
	modifications := make([]*mesh_proto.ProxyTemplate_Modifications, 0, 0)

	if cluster := modMap["cluster"].([]interface{}); len(cluster) > 0 {
		kumaClusters := createKumaProxyTemplateConfModificationCluster(cluster)
		modifications = append(modifications, kumaClusters...)
	}

	if listener, ok := modMap["listener"].([]interface{}); ok && len(listener) > 0 {
		kumaListeners := createKumaProxyTemplateConfModificationListener(listener)
		modifications = append(modifications, kumaListeners...)
	}

	if networkFilter, ok := modMap["network_filter"].([]interface{}); ok && len(networkFilter) > 0 {
		kumaNetworkFilter := createKumaProxyTemplateConfModificationNetworkFilter(networkFilter)
		modifications = append(modifications, kumaNetworkFilter...)
	}

	if httpFilter, ok := modMap["http_filters"].([]interface{}); ok && len(httpFilter) > 0 {
		kumaHttpFilter := createKumaProxyTemplateConfModificationHttpFilter(httpFilter)
		modifications = append(modifications, kumaHttpFilter...)
	}

	if virtualHost, ok := modMap["virtual_host"].([]interface{}); ok && len(virtualHost) > 0 {
		kumaVirtualHost := createKumaProxyTemplateConfModificationVirtualHost(virtualHost)
		modifications = append(modifications, kumaVirtualHost...)
	}

	return modifications
}

func createKumaProxyTemplateConfModificationCluster(clusters []interface{}) []*mesh_proto.ProxyTemplate_Modifications {
	kumaClusterArray := make([]*mesh_proto.ProxyTemplate_Modifications, 0, 0)
	for _, val := range clusters {
		cluster := &mesh_proto.ProxyTemplate_Modifications_Cluster{}
		clusterMap := val.(map[string]interface{})
		if operation, ok := clusterMap["operation"].(string); ok {
			cluster.Operation = operation
		}

		if value, ok := clusterMap["value"].(string); ok {
			cluster.Value = value
		}

		if match, ok := clusterMap["match"].([]interface{}); ok && len(match) > 0 {
			cluster.Match = createKumaProxyTemplateConfModificationClusterMatch(match[0])
		}

		modification := mesh_proto.ProxyTemplate_Modifications{}
		modification.Type = &mesh_proto.ProxyTemplate_Modifications_Cluster_{Cluster: cluster}
		kumaClusterArray = append(kumaClusterArray, &modification)
	}

	return kumaClusterArray
}

func createKumaProxyTemplateConfModificationClusterMatch(matchInterface interface{}) *mesh_proto.ProxyTemplate_Modifications_Cluster_Match {
	match := &mesh_proto.ProxyTemplate_Modifications_Cluster_Match{}
	matchMap := matchInterface.(map[string]interface{})
	if name, ok := matchMap["name"].(string); ok {
		match.Name = name
	}

	if origin, ok := matchMap["origin"].(string); ok {
		match.Origin = origin
	}

	return match
}

func createKumaProxyTemplateConfModificationListener(listeners []interface{}) []*mesh_proto.ProxyTemplate_Modifications {
	kumaListenerArray := make([]*mesh_proto.ProxyTemplate_Modifications, 0, 0)

	for _, val := range listeners {
		listener := &mesh_proto.ProxyTemplate_Modifications_Listener{}
		listenerMap := val.(map[string]interface{})

		if operation, ok := listenerMap["operation"].(string); ok {
			listener.Operation = operation
		}

		if value, ok := listenerMap["value"].(string); ok {
			listener.Value = value
		}

		if match, ok := listenerMap["match"].([]interface{}); ok && len(match) > 0 {
			listener.Match = createKumaProxyTemplateConfModificationListenerMatch(match[0])
		}

		modification := mesh_proto.ProxyTemplate_Modifications{}
		modification.Type = &mesh_proto.ProxyTemplate_Modifications_Listener_{Listener: listener}
		kumaListenerArray = append(kumaListenerArray, &modification)
	}

	return kumaListenerArray
}

func createKumaProxyTemplateConfModificationListenerMatch(matchInterface interface{}) *mesh_proto.ProxyTemplate_Modifications_Listener_Match {
	match := &mesh_proto.ProxyTemplate_Modifications_Listener_Match{}
	matchMap := matchInterface.(map[string]interface{})

	if name, ok := matchMap["name"].(string); ok {
		match.Name = name
	}

	if origin, ok := matchMap["origin"].(string); ok {
		match.Origin = origin
	}

	return match
}

func createKumaProxyTemplateConfModificationNetworkFilter(networkFilters []interface{}) []*mesh_proto.ProxyTemplate_Modifications {
	kumaNetworkFilterArray := make([]*mesh_proto.ProxyTemplate_Modifications, 0, 0)

	for _, val := range networkFilters {
		networkFilter := &mesh_proto.ProxyTemplate_Modifications_NetworkFilter{}
		networkFilterMap := val.(map[string]interface{})

		if operation, ok := networkFilterMap["operation"].(string); ok {
			networkFilter.Operation = operation
		}

		if value, ok := networkFilterMap["value"].(string); ok {
			networkFilter.Value = value
		}

		if match, ok := networkFilterMap["match"].([]interface{}); ok && len(match) > 0 {
			networkFilter.Match = createKumaProxyTemplateConfModificationNetworkFilterMatch(match[0])
		}

		modification := mesh_proto.ProxyTemplate_Modifications{}
		modification.Type = &mesh_proto.ProxyTemplate_Modifications_NetworkFilter_{NetworkFilter: networkFilter}
		kumaNetworkFilterArray = append(kumaNetworkFilterArray, &modification)
	}

	return kumaNetworkFilterArray
}

func createKumaProxyTemplateConfModificationNetworkFilterMatch(matchInterface interface{}) *mesh_proto.ProxyTemplate_Modifications_NetworkFilter_Match {
	match := &mesh_proto.ProxyTemplate_Modifications_NetworkFilter_Match{}
	matchMap := matchInterface.(map[string]interface{})

	if name, ok := matchMap["name"].(string); ok {
		match.Name = name
	}

	if origin, ok := matchMap["origin"].(string); ok {
		match.Origin = origin
	}

	if listenerName, ok := matchMap["listener_name"].(string); ok {
		match.ListenerName = listenerName
	}

	return match
}

func createKumaProxyTemplateConfModificationHttpFilter(httpFilters []interface{}) []*mesh_proto.ProxyTemplate_Modifications {
	kumaHttpFilterArray := make([]*mesh_proto.ProxyTemplate_Modifications, 0, 0)

	for _, val := range httpFilters {
		httpFilter := &mesh_proto.ProxyTemplate_Modifications_HttpFilter{}
		networkFilterMap := val.(map[string]interface{})

		if operation, ok := networkFilterMap["operation"].(string); ok {
			httpFilter.Operation = operation
		}

		if value, ok := networkFilterMap["value"].(string); ok {
			httpFilter.Value = value
		}

		if match, ok := networkFilterMap["match"].([]interface{}); ok && len(match) > 0 {
			httpFilter.Match = createKumaProxyTemplateConfModificationHttpFilterMatch(match[0])
		}

		modification := mesh_proto.ProxyTemplate_Modifications{}
		modification.Type = &mesh_proto.ProxyTemplate_Modifications_HttpFilter_{HttpFilter: httpFilter}
		kumaHttpFilterArray = append(kumaHttpFilterArray, &modification)
	}

	return kumaHttpFilterArray
}

func createKumaProxyTemplateConfModificationHttpFilterMatch(matchInterface interface{}) *mesh_proto.ProxyTemplate_Modifications_HttpFilter_Match {
	match := &mesh_proto.ProxyTemplate_Modifications_HttpFilter_Match{}
	matchMap := matchInterface.(map[string]interface{})

	if name, ok := matchMap["name"].(string); ok {
		match.Name = name
	}

	if origin, ok := matchMap["origin"].(string); ok {
		match.Origin = origin
	}

	if listenerName, ok := matchMap["listener_name"].(string); ok {
		match.ListenerName = listenerName
	}

	return match
}

func createKumaProxyTemplateConfModificationVirtualHost(virtualHosts []interface{}) []*mesh_proto.ProxyTemplate_Modifications {
	kumaVirtualHostArray := make([]*mesh_proto.ProxyTemplate_Modifications, 0, 0)

	for _, val := range virtualHosts {
		virtualHost := &mesh_proto.ProxyTemplate_Modifications_VirtualHost{}
		virtualHostMap := val.(map[string]interface{})

		if operation, ok := virtualHostMap["operation"].(string); ok {
			virtualHost.Operation = operation
		}

		if value, ok := virtualHostMap["value"].(string); ok {
			virtualHost.Value = value
		}

		if match, ok := virtualHostMap["match"].([]interface{}); ok && len(match) > 0 && match[0] != nil {
			virtualHost.Match = createKumaProxyTemplateConfModificationVirtualHostMatch(match[0])
		}

		modification := mesh_proto.ProxyTemplate_Modifications{}
		modification.Type = &mesh_proto.ProxyTemplate_Modifications_VirtualHost_{VirtualHost: virtualHost}
		kumaVirtualHostArray = append(kumaVirtualHostArray, &modification)
	}

	return kumaVirtualHostArray
}

func createKumaProxyTemplateConfModificationVirtualHostMatch(matchInterface interface{}) *mesh_proto.ProxyTemplate_Modifications_VirtualHost_Match {
	match := &mesh_proto.ProxyTemplate_Modifications_VirtualHost_Match{}
	matchMap := matchInterface.(map[string]interface{})

	if name, ok := matchMap["name"].(string); ok {
		match.Name = name
	}

	if origin, ok := matchMap["origin"].(string); ok {
		match.Origin = origin
	}

	if listenerName, ok := matchMap["route_configuration_name"].(string); ok {
		match.RouteConfigurationName = listenerName
	}

	return match
}

func flattenKumaProxyTemplateConf(conf *mesh_proto.ProxyTemplate_Conf) []interface{} {
	confMap := make(map[string]interface{})
	confSet := make([]interface{}, 0, 1)

	if conf == nil {
		return confSet
	}

	if conf.Imports != nil {
		confMap["imports"] = conf.Imports
	}

	if conf.Modifications != nil {
		interfaceList := make([]interface{}, 0, 1)
		interfaceList = append(interfaceList, flattenKumaProxyTemplateConfModifications(conf.Modifications))
		confMap["modifications"] = interfaceList

	}

	confSet = append(confSet, confMap)
	return confSet
}

func flattenKumaProxyTemplateConfModifications(modifications []*mesh_proto.ProxyTemplate_Modifications) map[string][]interface{} {
	modificationMap := make(map[string][]interface{})

	if modifications == nil {
		return modificationMap
	}

	for _, modification := range modifications {
		switch modification.Type.(type) {
		case *mesh_proto.ProxyTemplate_Modifications_Cluster_:
			clusterArray, ok := modificationMap["cluster"]
			if !ok {
				clusterArray = make([]interface{}, 0, 0)
				modificationMap["cluster"] = clusterArray
			}
			cluster := flattenKumaProxyTemplateConfModificationCluster(modification.GetCluster())
			clusterArray = append(clusterArray, cluster)
			modificationMap["cluster"] = clusterArray

		case *mesh_proto.ProxyTemplate_Modifications_Listener_:
			listenerArray, ok := modificationMap["listener"]
			if !ok {
				listenerArray = make([]interface{}, 0, 0)
				modificationMap["listener"] = listenerArray
			}
			listener := flattenKumaProxyTemplateConfModificationListener(modification.GetListener())
			listenerArray = append(listenerArray, listener)
			modificationMap["listener"] = listenerArray

		case *mesh_proto.ProxyTemplate_Modifications_NetworkFilter_:
			networkFilterArray, ok := modificationMap["network_filter"]
			if !ok {
				networkFilterArray = make([]interface{}, 0, 0)
				modificationMap["network_filter"] = networkFilterArray
			}
			networkFiter := flattenKumaProxyTemplateConfModificationNetworkFilter(modification.GetNetworkFilter())
			networkFilterArray = append(networkFilterArray, networkFiter)
			modificationMap["network_filter"] = networkFilterArray

		case *mesh_proto.ProxyTemplate_Modifications_HttpFilter_:
			httpFilterArray, ok := modificationMap["http_filters"]
			if !ok {
				httpFilterArray = make([]interface{}, 0, 0)
				modificationMap["http_filters"] = httpFilterArray
			}
			httpFilter := flattenKumaProxyTemplateConfModificationHttpFilter(modification.GetHttpFilter())
			httpFilterArray = append(httpFilterArray, httpFilter)
			modificationMap["http_filters"] = httpFilterArray

		case *mesh_proto.ProxyTemplate_Modifications_VirtualHost_:
			virtualHostArray, ok := modificationMap["virtual_host"]
			if !ok {
				virtualHostArray = make([]interface{}, 0, 0)
				modificationMap["virtual_host"] = virtualHostArray
			}
			virtualHost := flattenKumaProxyTemplateConfModificationVirtualHost(modification.GetVirtualHost())
			virtualHostArray = append(virtualHostArray, virtualHost)
			modificationMap["virtual_host"] = virtualHostArray

		}

	}

	return modificationMap

}

func flattenKumaProxyTemplateConfModificationCluster(cluster *mesh_proto.ProxyTemplate_Modifications_Cluster) map[string]interface{} {
	clusterMap := make(map[string]interface{})

	clusterMap["operation"] = cluster.Operation
	clusterMap["value"] = cluster.Value
	clusterMap["match"] = flattenKumaProxyTemplateConfModificationClusterMatch(cluster.Match)

	return clusterMap
}

func flattenKumaProxyTemplateConfModificationClusterMatch(match *mesh_proto.ProxyTemplate_Modifications_Cluster_Match) map[string]interface{} {
	matchMap := make(map[string]interface{})

	matchMap["name"] = match.Name
	matchMap["origin"] = match.Origin

	return matchMap
}

func flattenKumaProxyTemplateConfModificationListener(listener *mesh_proto.ProxyTemplate_Modifications_Listener) map[string]interface{} {
	listenerMap := make(map[string]interface{})

	listenerMap["operation"] = listener.Operation
	listenerMap["value"] = listener.Value
	listenerMap["match"] = flattenKumaProxyTemplateConfModificationListenerMatch(listener.Match)

	return listenerMap
}

func flattenKumaProxyTemplateConfModificationListenerMatch(match *mesh_proto.ProxyTemplate_Modifications_Listener_Match) map[string]interface{} {
	matchMap := make(map[string]interface{})

	matchMap["name"] = match.Name
	matchMap["origin"] = match.Origin

	return matchMap
}

func flattenKumaProxyTemplateConfModificationNetworkFilter(networkFilter *mesh_proto.ProxyTemplate_Modifications_NetworkFilter) map[string]interface{} {
	networkFilterMap := make(map[string]interface{})

	networkFilterMap["operation"] = networkFilter.Operation
	networkFilterMap["value"] = networkFilter.Value
	networkFilterMap["match"] = flattenKumaProxyTemplateConfModificationNetworkFilterMatch(networkFilter.Match)

	return networkFilterMap
}

func flattenKumaProxyTemplateConfModificationNetworkFilterMatch(match *mesh_proto.ProxyTemplate_Modifications_NetworkFilter_Match) map[string]interface{} {
	matchMap := make(map[string]interface{})

	matchMap["name"] = match.Name
	matchMap["origin"] = match.Origin
	matchMap["listener_name"] = match.ListenerName

	return matchMap
}

func flattenKumaProxyTemplateConfModificationHttpFilter(httpFilter *mesh_proto.ProxyTemplate_Modifications_HttpFilter) map[string]interface{} {
	httpFilterMap := make(map[string]interface{})

	httpFilterMap["operation"] = httpFilter.Operation
	httpFilterMap["value"] = httpFilter.Value
	httpFilterMap["match"] = flattenKumaProxyTemplateConfModificationHttpFilterMatch(httpFilter.Match)

	return httpFilterMap
}

func flattenKumaProxyTemplateConfModificationHttpFilterMatch(match *mesh_proto.ProxyTemplate_Modifications_HttpFilter_Match) map[string]interface{} {
	matchMap := make(map[string]interface{})

	matchMap["name"] = match.Name
	matchMap["origin"] = match.Origin
	matchMap["listener_name"] = match.ListenerName

	return matchMap
}

func flattenKumaProxyTemplateConfModificationVirtualHost(virtualHost *mesh_proto.ProxyTemplate_Modifications_VirtualHost) map[string]interface{} {
	virtualHostMap := make(map[string]interface{})

	virtualHostMap["operation"] = virtualHost.Operation
	virtualHostMap["value"] = virtualHost.Value
	if virtualHost.Match != nil {
		virtualHostMap["match"] = flattenKumaProxyTemplateConfModificationVirtualHostMatch(virtualHost.Match)
	}

	return virtualHostMap
}

func flattenKumaProxyTemplateConfModificationVirtualHostMatch(match *mesh_proto.ProxyTemplate_Modifications_VirtualHost_Match) map[string]interface{} {
	matchMap := make(map[string]interface{})

	matchMap["name"] = match.Name
	matchMap["origin"] = match.Origin
	matchMap["route_configuration_name"] = match.RouteConfigurationName

	return matchMap
}
