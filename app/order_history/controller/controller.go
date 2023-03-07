package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	httpresponse "sh-rk-test/app/httpResponse"
	"sh-rk-test/app/order_history/model"
)

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

// OrderHistoryController
type OrderHistoryController struct {
	Conn    gorm.DB
	RClient redis.Client
}

func NewOrderHistoriesController(e *echo.Echo, conn gorm.DB, cli redis.Client) {
	controller := &OrderHistoryController{conn, cli}

	e.GET("/history", controller.List)
	e.POST("/history", controller.Create)
	e.GET("/history/:id", controller.Detail)
	e.PUT("history/:id", controller.Update)
	e.DELETE("history/:id", controller.Delete)
}

func (oi *OrderHistoryController) List(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":   c.RealIP(),
		"path": c.Path(),
	}).Info("Incoming")
	var art []model.OrderHistory
	err := oi.Conn.Joins("User").Joins("OrderItem").Scopes(Paginate(c)).Find(&art).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	logrus.WithFields(logrus.Fields{
		"user": "Get List",
	}).Info("Response")
	return c.JSON(http.StatusOK, art)
}

func (oi *OrderHistoryController) Detail(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":   c.RealIP(),
		"path": c.Path(),
	}).Info("Incoming")
	res, err := oi.RClient.Get(context.Background(), "hist-"+c.Param("id")).Result()

	if err == nil {
		var jVar model.OrderHistory
		err = json.Unmarshal([]byte(res), &jVar)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"history": jVar.Model,
			}).Info("Response")
			return c.JSON(http.StatusOK, jVar)
		} else {
			fmt.Println(err)
		}
	}

	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, httpresponse.ErrNotFound.Error())
	}

	id := int64(idP)

	var art model.OrderHistory
	err = oi.Conn.Joins("User").Joins("OrderItem").First(&art, id).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	a, err := json.Marshal(art)
	n := len(a)
	rVar := string(a[:n])
	err = oi.RClient.Set(context.Background(), "hist-"+c.Param("id"), rVar, 0).Err()
	if err != nil {
		panic(err)
	}

	logrus.WithFields(logrus.Fields{
		"history": art.Model,
	}).Info("Response")
	return c.JSON(http.StatusOK, art)
}

func (oi *OrderHistoryController) Create(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":      c.RealIP(),
		"path":    c.Path(),
		"request": c.Request(),
	}).Info("Incoming")
	var o model.OrderHistoryPayload
	if err := c.Bind(&o); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}
	if err := c.Validate(o); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	orderHistory := model.OrderHistory{
		Descriptions: o.Descriptions,
		UserID:       o.UserID,
		OrderItemID:  o.OrderItemID,
	}

	oi.Conn.Create(&orderHistory)
	logrus.WithFields(logrus.Fields{
		"history": orderHistory.Model,
	}).Info("Response")
	return c.JSON(http.StatusOK, orderHistory)
}

func (oi *OrderHistoryController) Update(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":      c.RealIP(),
		"path":    c.Path(),
		"request": c.Request(),
	}).Info("Incoming")
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, httpresponse.ErrNotFound.Error())
	}

	id := int64(idP)

	var art model.OrderHistory
	err = oi.Conn.Joins("User").Joins("OrderItem").First(&art, id).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	var o model.OrderHistoryPayload
	if err := c.Bind(&o); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}
	if err := c.Validate(o); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	art.Descriptions = o.Descriptions
	art.UserID = o.UserID
	art.OrderItemID = o.OrderItemID

	oi.Conn.Save(&art)
	logrus.WithFields(logrus.Fields{
		"history": art.Model,
	}).Info("Response")
	return c.JSON(http.StatusOK, art)
}

func (oi *OrderHistoryController) Delete(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":   c.RealIP(),
		"path": c.Path(),
	}).Info("Incoming")
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, httpresponse.ErrNotFound.Error())
	}

	id := int64(idP)

	var art model.OrderHistory
	err = oi.Conn.First(&art, id).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	oi.Conn.Delete(&art)
	logrus.WithFields(logrus.Fields{
		"history": "deleted",
	}).Info("Response")
	return c.JSON(http.StatusOK, "Successfully Deleted")
}

func Paginate(c echo.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page == '0' {
			page = '1'
		}

		pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
