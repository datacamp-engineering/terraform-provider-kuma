package kuma

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func readStringFromResource(d *schema.ResourceData, key string) string {
	if value, ok := d.GetOk(key); ok {
		return value.(string)
	}
	return ""
}

func readArrayFromResource(d *schema.ResourceData, key string) []interface{} {
	if attr, ok := d.GetOk(key); ok {
		return attr.([]interface{})
	}
	return nil
}
