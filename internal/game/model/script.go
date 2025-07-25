package model

// Script 定义一个特定的游戏场景。
type Script struct {
	ID          string             `json:"id"`            // 唯一标识符
	Name        string             `json:"name"`          // 剧本名称
	Description string             `json:"description"`   // 剧本描述
	MainPlot    string             `json:"main_plot"`     // 主线剧情
	SubPlots    []string           `json:"sub_plots"`     // 支线剧情
	Characters  []CharacterConfig  `json:"characters"`    // 角色配置
	Tragedies   []TragedyCondition `json:"tragedies"`     // 悲剧条件
	LoopCount   int                `json:"loop_count"`    // 循环总次数
	DaysPerLoop int                `json:"days_per_loop"` // 每循环天数
}