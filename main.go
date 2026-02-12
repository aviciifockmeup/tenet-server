package main

import (
	"fmt"
	"tenet-server/config"
	"tenet-server/database"
	"tenet-server/logger"
	"tenet-server/router"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// 初始化日志
	if err := logger.Init(cfg.Log.Level); err != nil {
		panic("Failed to init logger: " + err.Error())
	}
	defer logger.Logger.Sync()

	logger.Info("Starting Tenet Server...")
    logger.Info("Config loaded", zap.String("name", cfg.Server.Name))

	// 初始化数据库/缓存
	if err := database.InitMySQl(cfg.Database.MySQL); err != nil {
		logger.Error("Failed to init MySQL", zap.Error(err))
		panic(err)
	}
	logger.Info("MySQL connected")

	if err := database.InitRedis(cfg.Database.Redis); err != nil {
        logger.Error("Failed to init Redis", zap.Error(err))
        panic(err)
    }
    logger.Info("Redis connected")

	// 创建Hertz服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
    h := server.Default(server.WithHostPorts(addr))

	router.Setup(h, cfg)

	// 启动服务
	logger.Info("Tenet Server is running on port 8081")
	h.Spin()
}