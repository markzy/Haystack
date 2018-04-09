package models

type PhotoMeta struct {
	PhotoID       string
	LogicalVolume int
	// 0 deleted, 1 ready, 2 uploading
	State         int
}



