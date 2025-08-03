// data/scripts/first_steps.cue

package data

// Script: First Steps
// 这是“First Steps”剧本的顶层定义。
{
	// 剧本名称
	name: "First Steps"
	// 剧本描述
	description: """
		A beginner script designed to teach the core mechanics of Tragedy Looper.
		The Mastermind is the Key Figure, and their goal is to cause a tragedy by manipulating the characters.
		The Protagonists must uncover the truth and prevent the tragedy from occurring.
		"""
	// 主谋（Mastermind）的定义
	mastermind: {
		// 角色名称
		character: "Key Figure"
		// 阴谋（plots）列表
		plots: ["Serial Murder Plan"]
		// 该剧本中主谋可用的能力
		abilities: []
	}
	// 剧本中的所有角色列表
	characters: [
		"Key Figure",
		"Shrine Maiden",
		"Office Worker",
		"Student",
	]
	// 初始事件列表
	incidents: [
		"Foul Murder",
		"Missing Person",
	]
	// 最终需要猜测的角色
	culprits: [
		"Key Figure",
	]
	// 循环（Loops）的总数
	loops: 4
	// 每一次循环的天数
	days: 2
	// 初始的 불안（Paranoia）值
	paranoia: 2
	// 初始的好感（Goodwill）值
	goodwill: 4
	// 该剧本中可用的卡牌
	cards: []
}