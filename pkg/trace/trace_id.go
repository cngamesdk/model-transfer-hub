package trace

import (
	"github.com/google/uuid"
)

// GenerateTraceID 生成链路追踪ID
func GenerateTraceID() string {
	return uuid.New().String()
}
