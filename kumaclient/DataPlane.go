package kumaclient

type DataPlane struct {
	Type       string      `json:"type"`
	Name       string      `json:"name"`
	Mesh       string      `json:"mesh"`
	Networking *Networking `json:"networking"`
}

type Networking struct {
	Address  string                `json:"address"`
	Gateway  *NetworkingGateway    `json:"gateway"`
	Inbound  *[]NetworkingInbound  `json:"inbound"`
	Outbound *[]NetworkingOutbound `json:"outbound"`
}

type NetworkingGateway struct {
	Tags map[string]string `json:"tags"`
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
