package models

type PhysicalVolume struct {
	MachineID int
	VolumeID  int
	Free      bool
}

type LogicalMapping struct {
	LogicalID int
	Volumes   []PhysicalVolume
}
