package kumaclient

type DataPlane struct {
	Type       string     `json:"type"`
	Name       string     `json:"name"`
	Mesh       string     `json:"mesh"`
	Networking Networking `json:"networking"`
}

type Networking struct {
	Address  string               `json:"address"`
	Inbound  []NetworkingInbound  `json:"inbound"`
	Outbound []NetworkingOutbound `json:"outbound"`
}

type NetworkingInbound struct {
	Port        int               `json:"port"`
	ServicePort int               `json:"servicePort"`
	Tags        map[string]string `json:"tags"`
}

type NetworkingOutbound struct {
	Port    int    `json:"port"`
	Service string `json:"service"`
}

func NewDataplane(mesh string, name string, address string, inbound NetworkingInbound) DataPlane {
	dataplane := DataPlane{}
	dataplane.Type = "Dataplane"
	dataplane.Name = name
	dataplane.Mesh = mesh
	dataplane.Networking = Networking{}
	dataplane.Networking.Address = address
	dataplane.Networking.Inbound = []NetworkingInbound{
		inbound,
	}

	return dataplane
}
