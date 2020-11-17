package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/nickvdyck/terraform-provider-kuma/kuma"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return kuma.Provider()
		},
	})
}

// package main

// import (
// 	"fmt"

// 	"github.com/nickvdyck/terraform-provider-kuma/kumaclient"
// )

// func main() {
// 	fmt.Println("hello world")

// 	client, _ := kumaclient.NewClient("http://localhost:5681", "123")
// 	mesh, _ := client.GetMesh("yolo2")

// 	fmt.Printf("Hello %s \n", mesh.Name)

// 	tags := make(map[string]string)

// 	tags["version"] = "123"
// 	tags["service"] = "backend"
// 	tags["kuma.io/service"] = "backend"

// 	dataplane := kumaclient.DataPlane{
// 		Type: "Dataplane",
// 		Mesh: mesh.Name,
// 		Name: "testing-dp",
// 		Networking: &kumaclient.Networking{
// 			Address: "127.0.0.1",
// 			Gateway: &kumaclient.NetworkingGateway{
// 				Tags: tags,
// 			},
// 		},
// 	}

// 	err := client.UpsertDataplane(dataplane)

// 	if err != nil {
// 		fmt.Printf("Shit aint right: %s", err)
// 	}
// }
