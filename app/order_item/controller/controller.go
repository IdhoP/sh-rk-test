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
	"sh-rk-test/app/order_item/model"
)

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

// OrderItemController
type OrderItemController struct {
	Conn    gorm.DB
	RClient redis.Client
}

func NewOrderItemsController(e *echo.Echo, conn gorm.DB, cli redis.Client) {
	controller := &OrderItemController{conn, cli}

	e.GET("/items", controller.List)
	e.POST("/items", controller.Create)
	e.GET("/items/:id", controller.Detail)
	e.PUT("items/:id", controller.Update)
	e.DELETE("items/:id", controller.Delete)
}

func (oi *OrderItemController) List(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":      c.RealIP(),
		"path":    c.Path(),
		"request": c.Request(),
	}).Info("Incoming")
	var art []model.OrderItem
	err := oi.Conn.Scopes(Paginate(c)).Find(&art).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	logrus.WithFields(logrus.Fields{
		"user": "Get List",
	}).Info("Response")
	return c.JSON(http.StatusOK, art)
}

func (oi *OrderItemController) Detail(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":      c.RealIP(),
		"path":    c.Path(),
		"request": c.Request(),
	}).Info("Incoming")
	res, err := oi.RClient.Get(context.Background(), "hist-"+c.Param("id")).Result()

	if err == nil {
		var jVar model.OrderItem
		err = json.Unmarshal([]byte(res), &jVar)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"item": jVar.Model,
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

	var art model.OrderItem
	err = oi.Conn.First(&art, id).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	a, err := json.Marshal(art)
	n := len(a)
	rVar := string(a[:n])
	err = oi.RClient.Set(context.Background(), "item-"+c.Param("id"), rVar, 0).Err()
	if err != nil {
		panic(err)
	}

	logrus.WithFields(logrus.Fields{
		"item": art.Model,
	}).Info("Response")
	return c.JSON(http.StatusOK, art)
}

func (oi *OrderItemController) Create(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":      c.RealIP(),
		"path":    c.Path(),
		"request": c.Request(),
	}).Info("Incoming")
	var o model.OrderItemPayload
	if err := c.Bind(&o); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}
	if err := c.Validate(o); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	orderItem := model.OrderItem{
		Name:      o.Name,
		Price:     o.Price,
		ExpiredAt: o.ExpiredAt,
	}

	oi.Conn.Create(&orderItem)
	logrus.WithFields(logrus.Fields{
		"item": orderItem.Model,
	}).Info("Response")
	return c.JSON(http.StatusOK, orderItem)
}

func (oi *OrderItemController) Update(c echo.Context) error {
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

	var art model.OrderItem
	err = oi.Conn.First(&art, id).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	var o model.OrderItemPayload
	if err := c.Bind(&o); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}
	if err := c.Validate(o); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	art.Name = o.Name
	art.Price = o.Price
	art.ExpiredAt = o.ExpiredAt

	oi.Conn.Save(&art)
	logrus.WithFields(logrus.Fields{
		"user": art.Model,
	}).Info("Response")
	return c.JSON(http.StatusOK, art)
}

func (oi *OrderItemController) Delete(c echo.Context) error {
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

	var art model.OrderItem
	err = oi.Conn.First(&art, id).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	name := art.Name
	oi.Conn.Delete(&art)
	logrus.WithFields(logrus.Fields{
		"item": name,
	}).Info("Response")
	return c.JSON(http.StatusOK, name+" Successfully Deleted")
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
