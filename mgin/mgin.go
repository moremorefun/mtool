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

// GinResp 通用返回
type GinResp struct {
	ErrCode int64  `json:"error"`
	ErrMsg  string `json:"error_msg"`
	Data    gin.H  `json:"data,omitempty"`
}

// GinRespSuccess 成功返回
var GinRespSuccess = GinResp{
	ErrCode: ErrorSuccess,
	ErrMsg:  ErrorSuccessMsg,
}

// GinRespInternalErr 成功返回
var GinRespInternalErr = GinResp{
	ErrCode: ErrorInternal,
	ErrMsg:  ErrorInternalMsg,
}

// GinRepeatReadBody 创建可重复度body
func GinRepeatReadBody(c *gin.Context) error {
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

// GinFillBindError 检测gin输入绑定错误
func GinFillBindError(c *gin.Context, err error) {
	repeatErr := GinRepeatReadBody(c)
	if repeatErr != nil {
		mlog.Log.Errorf("err: [%T] %s", repeatErr, repeatErr.Error())
	} else {
		body, _ := ioutil.ReadAll(c.Request.Body)
		mlog.Log.Warnf("bind error body is: %s", body)
	}
	GinDoRespErr(
		c,
		ErrorBind,
		fmt.Sprintf("[%T] %s", err, err.Error()),
		nil,
	)
}

// GinFillSuccessData 填充返回数据
func GinFillSuccessData(data gin.H) GinResp {
	return GinResp{
		ErrCode: ErrorSuccess,
		ErrMsg:  ErrorSuccessMsg,
		Data:    data,
	}
}

// GinDoRespSuccess 返回成功信息
func GinDoRespSuccess(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, GinResp{
		ErrCode: ErrorSuccess,
		ErrMsg:  ErrorSuccessMsg,
		Data:    data,
	})
}

// GinDoRespInternalErr 返回错误信息
func GinDoRespInternalErr(c *gin.Context) {
	c.JSON(http.StatusOK, GinResp{
		ErrCode: ErrorInternal,
		ErrMsg:  ErrorInternalMsg,
	})
}

// GinDoRespErr 返回特殊错误
func GinDoRespErr(c *gin.Context, code int64, msg string, data gin.H) {
	c.JSON(http.StatusOK, GinResp{
		ErrCode: code,
		ErrMsg:  msg,
		Data:    data,
	})
}

// GinDoEncRespSuccess 返回成功信息
func GinDoEncRespSuccess(c *gin.Context, key string, isAll bool, data gin.H) {
	var err error
	resp := GinResp{
		ErrCode: ErrorSuccess,
		ErrMsg:  ErrorSuccessMsg,
		Data:    data,
	}
	respBs := []byte("{}")
	if data != nil {
		respBs, err = jsoniter.Marshal(data)
		if err != nil {
			GinDoRespInternalErr(c)
			return
		}
	} else {
		resp.Data = gin.H{}
	}
	encResp, err := mencrypt.AesEncrypt(string(respBs), key)
	if err != nil {
		GinDoRespInternalErr(c)
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

// GinMidRepeatReadBody 创建可重复度body
func GinMidRepeatReadBody(c *gin.Context) {
	err := GinRepeatReadBody(c)
	if err != nil {
		mlog.Log.Errorf("err: [%T] %s", err, err.Error())
		GinDoRespInternalErr(c)
		c.Abort()
		return
	}
}

// GinMinTokenToUserID token转换为user_id
func GinMinTokenToUserID(tx mmysql.DbExeAble, getUserIDByToken func(ctx context.Context, tx mmysql.DbExeAble, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		err := GinRepeatReadBody(c)
		if err != nil {
			GinDoRespInternalErr(c)
			c.Abort()
			return
		}
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		err = c.ShouldBind(&req)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			GinFillBindError(c, err)
			c.Abort()
			return
		}
		bodyErr := GinRepeatReadBody(c)
		if bodyErr != nil {
			mlog.Log.Errorf("err: [%T] %s", bodyErr, bodyErr.Error())
			GinDoRespInternalErr(c)
			c.Abort()
			return
		}
		userID, err := getUserIDByToken(c, tx, req.Token)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			GinDoRespInternalErr(c)
			c.Abort()
			return
		}
		if userID == 0 {
			GinDoRespErr(c, ErrorToken, ErrorTokenMsg, nil)
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

// GinMinTokenToUserIDRedis token转换为user_id
func GinMinTokenToUserIDRedis(tx mmysql.DbExeAble, redisClient *redis.Client, getUserIDByToken func(ctx context.Context, tx mmysql.DbExeAble, redisClient *redis.Client, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		err := GinRepeatReadBody(c)
		if err != nil {
			GinDoRespInternalErr(c)
			c.Abort()
			return
		}
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		err = c.ShouldBind(&req)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			GinFillBindError(c, err)
			c.Abort()
			return
		}
		bodyErr := GinRepeatReadBody(c)
		if bodyErr != nil {
			mlog.Log.Errorf("err: [%T] %s", bodyErr, bodyErr.Error())
			GinDoRespInternalErr(c)
			c.Abort()
			return
		}
		userID, err := getUserIDByToken(c, tx, redisClient, req.Token)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			GinDoRespInternalErr(c)
			c.Abort()
			return
		}
		if userID == 0 {
			GinDoRespErr(c, ErrorToken, ErrorTokenMsg, nil)
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

// GinMinTokenToUserIDRedisIgnore token转换为user_id
func GinMinTokenToUserIDRedisIgnore(tx mmysql.DbExeAble, redisClient *redis.Client, getUserIDByToken func(ctx context.Context, tx mmysql.DbExeAble, redisClient *redis.Client, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		err := GinRepeatReadBody(c)
		if err != nil {
			GinDoRespInternalErr(c)
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
		bodyErr := GinRepeatReadBody(c)
		if bodyErr != nil {
			mlog.Log.Errorf("err: [%T] %s", bodyErr, bodyErr.Error())
			GinDoRespInternalErr(c)
			c.Abort()
			return
		}
		userID, err := getUserIDByToken(c, tx, redisClient, req.Token)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			GinDoRespInternalErr(c)
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
