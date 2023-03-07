package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	_orderHistoryModel "sh-rk-test/app/order_history/model"
	_orderItemModel "sh-rk-test/app/order_item/model"
	_userModel "sh-rk-test/app/user/model"
	"sh-rk-test/configs"

	_historyController "sh-rk-test/app/order_history/controller"
	_itemController "sh-rk-test/app/order_item/controller"
	_userController "sh-rk-test/app/user/controller"
)

type PayloadValidator struct {
	validator *validator.Validate
}

func (cv *PayloadValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	configs.InitEnvConfigs()

	f, err := os.OpenFile("logs/logger.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	logrus.SetFormatter(Formatter)
	logrus.SetOutput(f)

	dbHost := configs.EnvConfigs.DBHost
	dbPort := configs.EnvConfigs.DBPort
	dbUser := configs.EnvConfigs.DBUser
	dbPass := configs.EnvConfigs.DBPass
	dbName := configs.EnvConfigs.DBName

	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Jakarta")
	val.Add("charset", "utf8mb4")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(
		&_userModel.User{},
		&_orderHistoryModel.OrderHistory{},
		&_orderItemModel.OrderItem{},
	)

	rHost := configs.EnvConfigs.RedisHost
	rPort := configs.EnvConfigs.RedisPort
	rPass := configs.EnvConfigs.RedisPass
	rName := configs.EnvConfigs.RedisName

	addr := fmt.Sprintf("%s:%s", rHost, rPort)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: rPass, // no password set
		DB:       rName, // use default DB
	})

	e := echo.New()
	e.Validator = &PayloadValidator{validator: validator.New()}

	// ur := repository.NewUserRepository(*db)

	_userController.NewUsersController(e, *db, *rdb)
	_itemController.NewOrderItemsController(e, *db, *rdb)
	_historyController.NewOrderHistoriesController(e, *db, *rdb)

	httpServer := http.Server{
		Addr:    fmt.Sprintf(":%s", configs.EnvConfigs.ServerAddress),
		Handler: e, // <-- because Echo implements http.Handler interface
	}
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Print(fmt.Errorf("error when starting HTTP server: %w", err))
	}
}
