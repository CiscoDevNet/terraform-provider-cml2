package cmlclient

import "encoding/json"

type IDlist []string

// {
// 	"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 	"state": "DEFINED_ON_CORE",
// 	"created": "2021-02-28T07:33:47+00:00",
// 	"modified": "2021-02-28T07:33:47+00:00",
// 	"lab_title": "Lab at Mon 17:27 PM",
// 	"owner": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 	"lab_description": "string",
// 	"node_count": 0,
// 	"link_count": 0,
// 	"lab_notes": "string",
// 	"groups": [
// 	  {
// 		"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 		"permission": "read_only"
// 	  }
// 	]
//   }

type Lab struct {
	ID          string  `json:"id"`
	State       string  `json:"state"`
	Created     string  `json:"created"`
	Modified    string  `json:"modified"`
	Title       string  `json:"lab_title"`
	Owner       string  `json:"owner"`
	Description string  `json:"lab_description"`
	Notes       string  `json:"lab_notes"`
	NodeCount   int     `json:"node_count"`
	LinkCount   int     `json:"link_count"`
	Nodes       []*Node `json:"nodes"`
	Links       []*Link `json:"links"`
	// groups
}

// {
// 	"boot_disk_size": 0,
// 	"compute_id": "9c2519bf-dda6-4d31-942e-8068a6349b5e",
//# 	"configuration": "bridge0",
// 	"cpu_limit": 100,
// 	"cpus": 0,
// 	"data_volume": 0,
// 	"hide_links": false,
//# 	"id": "9efb1503-7e2a-4d2a-959e-865209f1acc0",
//# 	"image_definition": null,
// 	"lab_id": "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
//# 	"label": "ext-conn-0",
//# 	"node_definition": "external_connector",
// 	"ram": 0,
// 	"tags": [],
// 	"vnc_key": "",
// 	"x": 317,
// 	"y": 341,
// 	"config_filename": "noname",
// 	"config_mediatype": "ISO",
// 	"config_image_path": "/var/local/virl2/images/52d5c824-e10c-450a-b9c5-b700bd3bc17a/9efb1503-7e2a-4d2a-959e-865209f1acc0/config.img",
// 	"cpu_model": null,
// 	"data_image_path": "/var/local/virl2/images/52d5c824-e10c-450a-b9c5-b700bd3bc17a/9efb1503-7e2a-4d2a-959e-865209f1acc0/data.img",
// 	"disk_image": null,
// 	"disk_image_2": null,
// 	"disk_image_3": null,
// 	"disk_image_path": null,
// 	"disk_image_path_2": null,
// 	"disk_image_path_3": null,
// 	"disk_driver": null,
// 	"driver_id": "external_connector",
// 	"efi_boot": false,
// 	"image_dir": "/var/local/virl2/images/52d5c824-e10c-450a-b9c5-b700bd3bc17a/9efb1503-7e2a-4d2a-959e-865209f1acc0",
// 	"libvirt_image_dir": "/var/lib/libvirt/images/virl-base-images",
// 	"nic_driver": null,
// 	"number_of_serial_devices": 0,
// 	"serial_devices": [],
// 	"video_memory": 0,
// 	"video_model": null,
//# 	"state": "BOOTED",
// 	"boot_progress": "Booted"
//   }

type Node struct {
	ID              string       `json:"id"`
	Label           string       `json:"label"`
	X               int          `json:"x"`
	Y               int          `json:"y"`
	NodeDefinition  string       `json:"node_definition"`
	ImageDefinition string       `json:"image_definition"`
	Configuration   string       `json:"configuration"`
	CPUs            int          `json:"cpus"`
	RAM             int          `json:"ram"`
	State           string       `json:"state"`
	DataVolume      int          `json:"data_volume"`
	Interfaces      []*Interface `json:"interfaces"`
}

// {
// 	"id": "20681832-36e8-4ba9-9d8d-0588e0f7b517",
// 	"lab_id": "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
// 	"node": "9efb1503-7e2a-4d2a-959e-865209f1acc0",
// 	"label": "port",
// 	"slot": 0,
// 	"type": "physical",
// 	"device_name": "",
// 	"dst_udp_port": null,
// 	"src_udp_port": null,
// 	"mac_address": null,
// 	"is_connected": true,
// 	"state": "STARTED"
//   }

type Interface struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	State        string `json:"state"`
	MACaddress   string `json:"mac_address"`
	IsConnecteed bool   `json:"is_connected"`
	DeviceName   string `json:"device_name"`

	// these are extra
	IP4 []string `json:"ip4"`
	IP6 []string `json:"ip6"`
}

// {
// 	"id": "4d76f475-2915-444e-bfd1-425a517120bc",
// 	"interface_a": "20681832-36e8-4ba9-9d8d-0588e0f7b517",
// 	"interface_b": "1959cc9f-361c-410e-a960-9d9a896482a0",
// 	"lab_id": "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
// 	"label": "ext-conn-0-port<->unmanaged-switch-0-port0",
// 	"link_capture_key": "d827ce92-db2e-4933-bc0d-7a2c38e39ad5",
// 	"node_a": "9efb1503-7e2a-4d2a-959e-865209f1acc0",
// 	"node_b": "1cc0cbcd-6b4f-4bbe-9f69-2c3da5e3495a",
// 	"state": "STARTED"
// }

type link struct {
	ID      string `json:"id"`
	State   string `json:"state"`
	SrcID   string `json:"interface_a"`
	DstID   string `json:"interface_b"`
	SrcNode string `json:"node_a"`
	DstNode string `json:"node_b"`
	Label   string `json:"label"`
	PCAPkey string `json:"link_capture_key"`
}

type Link struct {
	ID      string `json:"id"`
	State   string `json:"state"`
	Src     *Interface
	Dst     *Interface
	Label   string `json:"label"`
	PCAPkey string `json:"link_capture_key"`
}

func (l Link) MarshalJSON() ([]byte, error) {
	link := link{
		ID:      l.ID,
		State:   l.State,
		SrcID:   l.Src.ID,
		DstID:   l.Dst.ID,
		SrcNode: "n/a",
		DstNode: "n/a",
		Label:   l.Label,
		PCAPkey: l.PCAPkey,
	}
	return json.Marshal(link)
}

// {
// 	"00da52b6-2683-49c0-ba3a-ace877dea4ca": {
// 	  "name": "alpine-0",
// 	  "interfaces": {
// 		"52:54:00:00:00:09": {
// 		  "id": "3b45184f-7041-4300-aef2-2b97d8e763a8",
// 		  "label": "eth0",
// 		  "ip4": [
// 			"192.168.122.35"
// 		  ],
// 		  "ip6": [
// 			"fe80::5054:ff:fe00:9"
// 		  ]
// 		}
// 	  }
// 	},
// 	"0df7a717-9826-4729-9fe1-bc4932498c83": {
// 	  "name": "alpine-1",
// 	  "interfaces": {
// 		"52:54:00:00:00:08": {
// 		  "id": "6bec8956-f812-4fb3-9551-aef4410807ec",
// 		  "label": "eth0",
// 		  "ip4": [
// 			"192.168.122.34"
// 		  ],
// 		  "ip6": [
// 			"fe80::5054:ff:fe00:8"
// 		  ]
// 		}
// 	  }
// 	}
//   }

type l3nodes map[string]*l3node

type l3node struct {
	Name       string                 `json:"name"`
	Interfaces map[string]l3interface `json:"interfaces"`
}

type l3interface struct {
	ID    string   `json:"id"`
	Label string   `json:"label"`
	IP4   []string `json:"ip4"`
	IP6   []string `json:"ip6"`
}
