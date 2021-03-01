package kuma

import (
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/durationpb"
)

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

func readDurationFromString(d string) (*duration.Duration, error) {
	parsed, err := time.ParseDuration(d)

	if err != nil {
		return nil, err
	}

	return durationpb.New(parsed), nil
}
