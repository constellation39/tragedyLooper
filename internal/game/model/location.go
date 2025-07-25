package model

// LocationType 定义游戏中的可能地点。
type LocationType string

const (
	LocationHospital LocationType = "Hospital"
	LocationShrine   LocationType = "Shrine"
	LocationCity     LocationType = "City"
	LocationSchool   LocationType = "School"
)
