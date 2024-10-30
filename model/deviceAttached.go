package model

type DeviceAttached struct {
	DevicePath string `json:"device_path"`
	DryRun     bool   `json:"dry_run"`
	// EjectOnComplete       bool   `json:"eject_on_complete"`
	// EmptyVolumeOnComplete bool   `json:"empty_on_complete"`
}
