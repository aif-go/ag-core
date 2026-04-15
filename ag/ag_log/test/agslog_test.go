package test

import (
	"fmt"
	"log/slog"
)

func _test_log(logger *slog.Logger, msg string) {
	logger.Info(fmt.Sprintf("logi: %s", msg))
	logger.Debug(fmt.Sprintf("logd: %s", msg))
	logger.Warn(fmt.Sprintf("logw: %s", msg))
	logger.Error(fmt.Sprintf("loge: %s", msg))
}
