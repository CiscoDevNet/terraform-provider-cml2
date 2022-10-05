package cmlclient

import "context"

// [
//   {
//     "id": "alpine-3-10-base",
//     "node_definition_id": "alpine",
//     "description": "Alpine Linux and network tools",
//     "label": "Alpine 3.10",
//     "disk_image": "alpine-3-10-base.qcow2",
//     "read_only": true,
//     "ram": null,
//     "cpus": null,
//     "cpu_limit": null,
//     "data_volume": null,
//     "boot_disk_size": null,
//     "disk_subfolder": "alpine-3-10-base",
//     "schema_version": "0.0.1"
//   },
// ]

type ImageDefinition struct {
	ID            string `json:"id"`
	NodeDefID     string `json:"node_definition_id"`
	Description   string `json:"description"`
	Label         string `json:"label"`
	DiskImage     string `json:"disk_image"`
	ReadOnly      bool   `json:"read_only"`
	RAM           *int   `json:"ram"`
	CPUs          *int   `json:"cpus"`
	CPUlimit      *int   `json:"cpu_limit"`
	DataVolume    *int   `json:"data_volume"`
	BootDiskSize  *int   `json:"boot_disk_size"`
	DiskSubfolder string `json:"disk_subfolder"`
	SchemaVersion string `json:"schema_version"`
}

func (c *Client) GetImageDefs(ctx context.Context) ([]ImageDefinition, error) {
	imgDef := []ImageDefinition{}
	err := c.jsonGet(ctx, "image_definitions", &imgDef)
	if err != nil {
		return nil, err
	}
	return imgDef, nil
}
