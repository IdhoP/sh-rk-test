package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	httpresponse "sh-rk-test/app/httpResponse"
	"sh-rk-test/app/user/model"
)

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

// UserController
type UserController struct {
	Conn    gorm.DB
	RClient redis.Client
}

func NewUsersController(e *echo.Echo, conn gorm.DB, cli redis.Client) {
	controller := &UserController{conn, cli}

	e.GET("/users", controller.List)
	e.POST("/users", controller.Create)
	e.GET("/users/:id", controller.Detail)
	e.PUT("users/:id", controller.Update)
	e.DELETE("users/:id", controller.Delete)
}

func (us *UserController) List(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":   c.RealIP(),
		"path": c.Path(),
	}).Info("Incoming")
	var art []model.User
	err := us.Conn.Scopes(Paginate(c)).Find(&art).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	logrus.WithFields(logrus.Fields{
		"user": "Get List",
	}).Info("Response")
	return c.JSON(http.StatusOK, art)
}

func (us *UserController) Detail(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":   c.RealIP(),
		"path": c.Path(),
	}).Info("Incoming")
	res, err := us.RClient.Get(context.Background(), "user-"+c.Param("id")).Result()

	if err == nil {
		var jVar model.User
		err = json.Unmarshal([]byte(res), &jVar)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"user": jVar.Model,
			}).Info("Response")
			return c.JSON(http.StatusOK, jVar)
		}
	}

	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, httpresponse.ErrNotFound.Error())
	}

	id := int64(idP)

	var art model.User
	err = us.Conn.First(&art, id).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	a, err := json.Marshal(art)
	n := len(a)
	rVar := string(a[:n])
	err = us.RClient.Set(context.Background(), "user-"+c.Param("id"), rVar, 0).Err()
	if err != nil {
		panic(err)
	}
	logrus.WithFields(logrus.Fields{
		"user": art.Model,
	}).Info("Response")
	return c.JSON(http.StatusOK, art)
}

func (us *UserController) Create(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":      c.RealIP(),
		"path":    c.Path(),
		"request": c.Request(),
	}).Info("Incoming")
	var u model.UserPayload
	if err := c.Bind(&u); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}
	if err := c.Validate(u); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	user := model.User{
		FullName:   u.FullName,
		FirstOrder: u.FirstOrder,
	}

	us.Conn.Create(&user)
	logrus.WithFields(logrus.Fields{
		"user": user.Model,
	}).Info("Response")
	return c.JSON(http.StatusOK, user)
}

func (us *UserController) Update(c echo.Context) error {
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

	var art model.User
	err = us.Conn.First(&art, id).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	var u model.UserPayload
	if err := c.Bind(&u); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}
	if err := c.Validate(u); err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	art.FullName = u.FullName
	art.FirstOrder = u.FirstOrder

	us.Conn.Save(&art)
	logrus.WithFields(logrus.Fields{
		"user": art.Model,
	}).Info("Response")
	return c.JSON(http.StatusOK, art)
}

func (us *UserController) Delete(c echo.Context) error {
	logrus.WithFields(logrus.Fields{
		"IP":   c.RealIP(),
		"path": c.Path(),
	}).Info("Incoming")
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, httpresponse.ErrNotFound.Error())
	}

	id := int64(idP)

	var art model.User
	err = us.Conn.First(&art, id).Error
	if err != nil {
		return c.JSON(httpresponse.GetStatusCode(err), ResponseError{Message: err.Error()})
	}

	name := art.FullName
	us.Conn.Delete(&art)
	logrus.WithFields(logrus.Fields{
		"user": name,
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
