package model

type ImportVolume struct {
	VolumePath string `json:"volume_path"`
	// DevicePath string `json:"device_path"`
	DryRun bool `json:"dry_run"`
}
