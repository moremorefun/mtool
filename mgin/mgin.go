package mgin

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/moremorefun/mtool/mmysql"

	"github.com/go-redis/redis"
	"github.com/moremorefun/mtool/mencrypt"

	"github.com/moremorefun/mtool/mlog"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// Resp 通用返回
type Resp struct {
	ErrCode int64  `json:"error"`
	ErrMsg  string `json:"error_msg"`
	Data    gin.H  `json:"data,omitempty"`
}

// RespSuccess 成功返回
var RespSuccess = Resp{
	ErrCode: ErrorSuccess,
	ErrMsg:  ErrorSuccessMsg,
}

// RespInternalErr 成功返回
var RespInternalErr = Resp{
	ErrCode: ErrorInternal,
	ErrMsg:  ErrorInternalMsg,
}

// RepeatReadBody 创建可重复度body
func RepeatReadBody(c *gin.Context) error {
	var err error
	var body []byte
	if cb, ok := c.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body == nil {
		body, err = ioutil.ReadAll(c.Request.Body)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			c.Abort()
			return err
		}
		c.Set(gin.BodyBytesKey, body)
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return nil
}

// FillBindError 检测gin输入绑定错误
func FillBindError(c *gin.Context, err error) {
	repeatErr := RepeatReadBody(c)
	if repeatErr != nil {
		mlog.Log.Errorf("err: [%T] %s", repeatErr, repeatErr.Error())
	} else {
		body, _ := ioutil.ReadAll(c.Request.Body)
		mlog.Log.Warnf("bind error body is: %s", body)
	}
	DoRespErr(
		c,
		ErrorBind,
		fmt.Sprintf("[%T] %s", err, err.Error()),
		nil,
	)
}

// DoRespSuccess 返回成功信息
func DoRespSuccess(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, Resp{
		ErrCode: ErrorSuccess,
		ErrMsg:  ErrorSuccessMsg,
		Data:    data,
	})
}

// DoRespInternalErr 返回错误信息
func DoRespInternalErr(c *gin.Context) {
	c.JSON(http.StatusOK, Resp{
		ErrCode: ErrorInternal,
		ErrMsg:  ErrorInternalMsg,
	})
}

// DoRespErr 返回特殊错误
func DoRespErr(c *gin.Context, code int64, msg string, data gin.H) {
	c.JSON(http.StatusOK, Resp{
		ErrCode: code,
		ErrMsg:  msg,
		Data:    data,
	})
}

// DoEncRespSuccess 返回成功信息
func DoEncRespSuccess(c *gin.Context, key string, isAll bool, data gin.H) {
	var err error
	resp := Resp{
		ErrCode: ErrorSuccess,
		ErrMsg:  ErrorSuccessMsg,
		Data:    data,
	}
	respBs := []byte("{}")
	if data != nil {
		respBs, err = jsoniter.Marshal(data)
		if err != nil {
			DoRespInternalErr(c)
			return
		}
	} else {
		resp.Data = gin.H{}
	}
	encResp, err := mencrypt.AesEncrypt(string(respBs), key)
	if err != nil {
		DoRespInternalErr(c)
		return
	}
	if isAll {
		resp.Data["enc"] = encResp
	} else {
		resp.Data = gin.H{
			"enc": encResp,
		}
	}
	c.JSON(http.StatusOK, resp)
}

// MidRepeatReadBody 创建可重复度body
func MidRepeatReadBody(c *gin.Context) {
	err := RepeatReadBody(c)
	if err != nil {
		mlog.Log.Errorf("err: [%T] %s", err, err.Error())
		DoRespInternalErr(c)
		c.Abort()
		return
	}
}

// MinTokenToUserID token转换为user_id
func MinTokenToUserID(tx mmysql.DbExeAble, getUserIDByToken func(ctx context.Context, tx mmysql.DbExeAble, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		err := RepeatReadBody(c)
		if err != nil {
			DoRespInternalErr(c)
			c.Abort()
			return
		}
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		err = c.ShouldBind(&req)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			FillBindError(c, err)
			c.Abort()
			return
		}
		bodyErr := RepeatReadBody(c)
		if bodyErr != nil {
			mlog.Log.Errorf("err: [%T] %s", bodyErr, bodyErr.Error())
			DoRespInternalErr(c)
			c.Abort()
			return
		}
		userID, err := getUserIDByToken(c, tx, req.Token)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			DoRespInternalErr(c)
			c.Abort()
			return
		}
		if userID == 0 {
			DoRespErr(c, ErrorToken, ErrorTokenMsg, nil)
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

// MinTokenToUserIDRedis token转换为user_id
func MinTokenToUserIDRedis(tx mmysql.DbExeAble, redisClient *redis.Client, getUserIDByToken func(ctx context.Context, tx mmysql.DbExeAble, redisClient *redis.Client, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		err := RepeatReadBody(c)
		if err != nil {
			DoRespInternalErr(c)
			c.Abort()
			return
		}
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		err = c.ShouldBind(&req)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			FillBindError(c, err)
			c.Abort()
			return
		}
		bodyErr := RepeatReadBody(c)
		if bodyErr != nil {
			mlog.Log.Errorf("err: [%T] %s", bodyErr, bodyErr.Error())
			DoRespInternalErr(c)
			c.Abort()
			return
		}
		userID, err := getUserIDByToken(c, tx, redisClient, req.Token)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			DoRespInternalErr(c)
			c.Abort()
			return
		}
		if userID == 0 {
			DoRespErr(c, ErrorToken, ErrorTokenMsg, nil)
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

// MinTokenToUserIDRedisIgnore token转换为user_id
func MinTokenToUserIDRedisIgnore(tx mmysql.DbExeAble, redisClient *redis.Client, getUserIDByToken func(ctx context.Context, tx mmysql.DbExeAble, redisClient *redis.Client, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		err := RepeatReadBody(c)
		if err != nil {
			DoRespInternalErr(c)
			c.Abort()
			return
		}
		var req struct {
			Token string `json:"token" binding:"omitempty"`
		}
		err = c.ShouldBind(&req)
		if err != nil {
			c.Next()
			return
		}
		bodyErr := RepeatReadBody(c)
		if bodyErr != nil {
			mlog.Log.Errorf("err: [%T] %s", bodyErr, bodyErr.Error())
			DoRespInternalErr(c)
			c.Abort()
			return
		}
		userID, err := getUserIDByToken(c, tx, redisClient, req.Token)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			DoRespInternalErr(c)
			c.Abort()
			return
		}
		if userID == 0 {
			c.Next()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}
