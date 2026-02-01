package util

import (
	"fmt"
	"time"
)

func GetSystemPrompt() string {
	return SoulPrompt + "\n\n" + SystemPrompt + fmt.Sprintf("\n\n 当前时间: %s", time.Now().Format(time.DateTime))
}
