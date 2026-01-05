package util

import (
	"fmt"
	"time"
)

func GetSystemPrompt() string {
	return SystemPrompt + fmt.Sprintf("\n\n 当前时间: %s", time.Now().Format(time.DateTime))
}
