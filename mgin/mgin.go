package mgin

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/moremorefun/mtool/mdb"

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

func GinBodyRepeat(r io.Reader) (io.ReadCloser, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &nopBodyRepeat{body: body}, nil
}

// GinMidFilterEnc 获取加密中间件
func GinMidFilterEnc(key string, isForce bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Enc string `form:"enc" binding:"omitempty"`
		}
		err := c.ShouldBind(&req)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			DoRespInternalErr(c)
			c.Abort()
			return
		}
		if len(req.Enc) > 0 {
			// 解密
			deStr, err := mencrypt.AesDecrypt(req.Enc, key)
			if err != nil {
				mlog.Log.Errorf("err: [%T] %s", err, err.Error())
				DoRespInternalErr(c)
				c.Abort()
				return
			}
			c.Request.Body = &nopBodyRepeat{body: []byte(deStr)}
		} else {
			if isForce {
				mlog.Log.Errorf("err: [%T] %s", err, err.Error())
				DoRespInternalErr(c)
				c.Abort()
				return
			}
		}
	}
}

type nopBodyRepeat struct {
	body []byte
	i    int
}

func (o *nopBodyRepeat) Read(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return 0, nil
	}
	remain := len(o.body) - o.i
	if remain < n {
		n = copy(p, o.body[o.i:])
		o.i = 0
		return n, io.EOF
	}
	n = copy(p, o.body[o.i:])
	o.i += n
	return n, nil
}

func (*nopBodyRepeat) Close() error { return nil }

// FillBindError 检测gin输入绑定错误
func FillBindError(c *gin.Context, err error) {
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
	var err error
	c.Request.Body, err = GinBodyRepeat(c.Request.Body)
	if err != nil {
		mlog.Log.Errorf("err: [%T] %s", err, err.Error())
		DoRespInternalErr(c)
		c.Abort()
		return
	}
}

// MinTokenToUserID token转换为user_id
func MinTokenToUserID(tx mdb.ExecuteAble, getUserIDByToken func(ctx context.Context, tx mdb.ExecuteAble, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		err := c.ShouldBind(&req)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			FillBindError(c, err)
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
func MinTokenToUserIDRedis(tx mdb.ExecuteAble, redisClient *redis.Client, getUserIDByToken func(ctx context.Context, tx mdb.ExecuteAble, redisClient *redis.Client, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		err := c.ShouldBind(&req)
		if err != nil {
			mlog.Log.Errorf("err: [%T] %s", err, err.Error())
			FillBindError(c, err)
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
func MinTokenToUserIDRedisIgnore(tx mdb.ExecuteAble, redisClient *redis.Client, getUserIDByToken func(ctx context.Context, tx mdb.ExecuteAble, redisClient *redis.Client, token string) (int64, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"omitempty"`
		}
		err := c.ShouldBind(&req)
		if err != nil {
			c.Next()
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

func GinCors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if len(origin) == 0 {
			// request is not a CORS request
			return
		}
		reqHeader := c.Request.Header.Get("Access-Control-Request-Headers")
		method := c.Request.Method
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Max-Age", "43200")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", reqHeader)

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
