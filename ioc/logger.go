package ioc

import (
	"go.uber.org/zap"
	"project/pkg/logger"
)

/**
 * @Description
 * @Date 2025/4/2 22:22
 **/
func InitLogger() logger.LoggerV1 {
	cfg := zap.NewDevelopmentConfig()
	L, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(L)
}
