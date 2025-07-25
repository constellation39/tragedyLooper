package model

// Script 定义一个特定的游戏场景。
type Script struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	MainPlot    string             `json:"main_plot"`     // 例如："谋杀计划", "封印物品"
	SubPlots    []string           `json:"sub_plots"`     // 例如："朋友圈", "阴谋"
	Characters  []CharacterConfig  `json:"characters"`    // 此剧本的初始角色配置
	Tragedies   []TragedyCondition `json:"tragedies"`     // 此剧本预定义的悲剧
	LoopCount   int                `json:"loop_count"`    // 允许的总循环次数
	DaysPerLoop int                `json:"days_per_loop"` // 每循环天数
	// 添加任何剧本特定的规则或初始设置
}
