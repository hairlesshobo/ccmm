package model

type ImportVolume struct {
	VolumePath            string `json:"volume_path"`
	EjectOnComplete       bool   `json:"eject_on_complete"`
	EmptyVolumeOnComplete bool   `json:"empty_on_complete"`
	DevicePath            string `json:"device_path"`
	DeviceID              string `json:"device_id"`
	DryRun                bool   `json:"dry_run"`
}
