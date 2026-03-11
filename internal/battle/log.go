package battle

import "strings"

const MaxLogEntries = 10

// AddBattleLog appends a message to the battle log (fixed-size).
// LastMessage is kept as a quick alias to the last entry.
func (c *BattleContext) AddBattleLog(message string) {
	if c == nil {
		return
	}
	msg := strings.TrimSpace(message)
	if msg == "" {
		return
	}
	c.LastMessage = msg
	c.BattleLog = append(c.BattleLog, msg)
	if len(c.BattleLog) > MaxLogEntries {
		overflow := len(c.BattleLog) - MaxLogEntries
		c.BattleLog = c.BattleLog[overflow:]
	}
}

