package model

// LocationType 定义游戏中的可能地点。
type LocationType string

const (
	LocationHospital LocationType = "Hospital" // 医院
	LocationShrine   LocationType = "Shrine"   // 神社
	LocationCity     LocationType = "City"     // 城市
	LocationSchool   LocationType = "School"   // 学校
)