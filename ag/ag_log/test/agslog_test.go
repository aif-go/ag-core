package test

import (
	"fmt"
	"log/slog"
)

func _test_log(logger *slog.Logger, msg string) {
	logger.Debug(fmt.Sprintf("logd: %s", msg))
	logger.Info(fmt.Sprintf("logi: %s", msg))
	logger.Warn(fmt.Sprintf("logw: %s", msg))
	logger.Error(fmt.Sprintf("loge: %s", msg))
}
