package models

type PhysicalVolume struct {
	MachineID int
	VolumeID  int
}

type LogicalMapping struct {
	LogicalID int
	Volumes   []PhysicalVolume
	Free      bool
}
