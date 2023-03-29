package cmlclient

import (
	"context"
	"sort"
)

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
	SchemaVersion string `json:"schema_version"`
	NodeDefID     string `json:"node_definition_id"`
	Description   string `json:"description"`
	Label         string `json:"label"`
	DiskImage1    string `json:"disk_image"`
	DiskImage2    string `json:"disk_image_2"`
	DiskImage3    string `json:"disk_image_3"`
	ReadOnly      bool   `json:"read_only"`
	DiskSubfolder string `json:"disk_subfolder"`
	RAM           *int   `json:"ram"`
	CPUs          *int   `json:"cpus"`
	CPUlimit      *int   `json:"cpu_limit"`
	DataVolume    *int   `json:"data_volume"`
	BootDiskSize  *int   `json:"boot_disk_size"`
}

// ImageDefinitions returns a list of image definitions known to the controller.
func (c *Client) ImageDefinitions(ctx context.Context) ([]ImageDefinition, error) {
	imgDef := []ImageDefinition{}
	err := c.jsonGet(ctx, "image_definitions", &imgDef, 0)
	if err != nil {
		return nil, err
	}

	// sort the image list by their ID
	sort.Slice(imgDef, func(i, j int) bool {
		return imgDef[i].ID > imgDef[j].ID
	})

	return imgDef, nil
}
