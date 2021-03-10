package mquery

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

// SQLAble sql语句生成接口
type SQLAble interface {
	AppendToQuery(bytes.Buffer, gin.H) (bytes.Buffer, gin.H, error)
}
