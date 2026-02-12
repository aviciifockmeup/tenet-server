package database

import (
	"tenet-server/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"fmt"
)

var DB *gorm.DB

func InitMySQl(cfg config.MySQLConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        cfg.Username,
        cfg.Password,
        cfg.Host,
        cfg.Port,
        cfg.Database,
    )

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}