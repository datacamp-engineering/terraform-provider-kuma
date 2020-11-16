package kuma

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nickvdyck/terraform-provider-kuma/kumaclient"
)

func resourceDataplane() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataplaneCreate,
		ReadContext:   resourceDataplaneRead,
		UpdateContext: resourceDataplaneUpdate,
		DeleteContext: resourceDataplaneDelete,
		Schema: map[string]*schema.Schema{
			"mesh": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"networking": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"inbound": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"service_port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"tags": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem: &schema.Schema{
											Type:     schema.TypeMap,
											Optional: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
									},
								},
							},
						},
						"outbound": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"service": {
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
	}
}

func resourceDataplaneCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	dataplane := createKumaDataplaneFromResourceData(d)

	c := m.(*kumaclient.KumaClient)

	err := c.UpsertDataplane(*dataplane)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dataplane.Name)
	return resourceDataplaneRead(ctx, d, m)
}

func resourceDataplaneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*kumaclient.KumaClient)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	mesh := d.Get("mesh").(string)

	dataplane, err := c.GetDataPlane(mesh, name)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", dataplane.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("mesh", dataplane.Mesh); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("network", flattenKumaNetworking(&dataplane.Networking)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDataplaneUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// c := m.(*kumaclient.KumaClient)
	dataplane := createKumaDataplaneFromResourceData(d)

	c := m.(*kumaclient.KumaClient)

	err := c.UpsertDataplane(*dataplane)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dataplane.Name)
	return resourceDataplaneRead(ctx, d, m)
}

func resourceDataplaneDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*kumaclient.KumaClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Id()
	mesh := d.Get("mesh").(string)

	err := c.DeleteDataplaneByName(mesh, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func createKumaDataplaneFromResourceData(d *schema.ResourceData) *kumaclient.DataPlane {
	dataplane := &kumaclient.DataPlane{}

	dataplane.Type = "Dataplane"
	dataplane.Mesh = readStringFromResource(d, "mesh")
	dataplane.Name = readStringFromResource(d, "name")

	if networkingArray := readArrayFromResource(d, "networking"); networkingArray != nil && len(networkingArray) > 0 {
		networkingMap := networkingArray[0].(map[string]interface{})
		dataplane.Networking = *createKumaNetworkingFromMap(&networkingMap)
	}

	return dataplane
}

func createKumaNetworkingFromMap(data *map[string]interface{}) *kumaclient.Networking {
	if data != nil {
		dataMap := *data
		networking := &kumaclient.Networking{}

		if dataMap["address"] != nil {
			address := dataMap["address"].(string)
			networking.Address = address
		}

		if dataMap["inbound"] != nil {
			if inboundArray := dataMap["inbound"].([]interface{}); inboundArray != nil && len(inboundArray) > 0 {
				networking.Inbound = *createKumaNetworkingInboundFromArray(&inboundArray)
			}
		}

		if dataMap["outbound"] != nil {
			if outboundArray := dataMap["outbound"].([]interface{}); outboundArray != nil && len(outboundArray) > 0 {
				networking.Outbound = *createKumaNetworkingOutboundFromArray(&outboundArray)
			}
		}

		return networking
	}
	return nil
}

func createKumaNetworkingInboundFromArray(data *[]interface{}) *[]kumaclient.NetworkingInbound {
	if data != nil {
		dataList := *data
		inbound := []kumaclient.NetworkingInbound{}

		for _, inboundData := range dataList {
			inboundMap := inboundData.(map[string]interface{})
			item := kumaclient.NetworkingInbound{}

			if inboundMap["port"] != nil {
				port := inboundMap["port"].(int)
				item.Port = port
			}

			if inboundMap["service_port"] != nil {
				servicePort := inboundMap["service_port"].(int)
				item.ServicePort = servicePort
			}

			if inboundMap["tags"] != nil {
				tags := inboundMap["tags"].(map[string]string)
				item.Tags = tags
			}

			inbound = append(inbound, item)
		}
		return &inbound
	}
	return nil
}

func createKumaNetworkingOutboundFromArray(data *[]interface{}) *[]kumaclient.NetworkingOutbound {
	if data != nil {
		dataList := *data
		outbound := []kumaclient.NetworkingOutbound{}

		for _, inboundData := range dataList {
			inboundMap := inboundData.(map[string]interface{})
			item := kumaclient.NetworkingOutbound{}

			if inboundMap["port"] != nil {
				port := inboundMap["port"].(int)
				item.Port = port
			}

			if inboundMap["service"] != nil {
				servicePort := inboundMap["service"].(string)
				item.Service = servicePort
			}

			outbound = append(outbound, item)
		}
		return &outbound
	}
	return nil
}

func flattenKumaNetworking(in *kumaclient.Networking) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["address"] = in.Address

	if in.Inbound != nil {
		m["inbound"] = flattenKumaNetworkingInbound(in.Inbound)
	}

	if in.Outbound != nil {
		m["outbound"] = flattenKumaNetworkingOutbound(in.Outbound)
	}

	return []interface{}{m}
}

func flattenKumaNetworkingInbound(in []kumaclient.NetworkingInbound) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	flatList := make([]interface{}, len(in))

	for _, item := range in {
		s := make(map[string]interface{})

		s["port"] = item.Port
		s["service_port"] = item.ServicePort
		s["tags"] = item.Tags

		flatList = append(flatList, s)
	}

	return flatList
}

func flattenKumaNetworkingOutbound(in []kumaclient.NetworkingOutbound) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	flatList := make([]interface{}, len(in))

	for _, item := range in {
		s := make(map[string]interface{})

		s["port"] = item.Port
		s["service"] = item.Service

		flatList = append(flatList, s)
	}

	return flatList
}
