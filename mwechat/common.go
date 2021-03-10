package mwechat

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// XMLNode xml结构
type XMLNode struct {
	XMLName xml.Name
	Content string    `xml:",chardata"`
	Nodes   []XMLNode `xml:",any"`
}

func walk(nodes []XMLNode, f func(XMLNode) bool) {
	for _, n := range nodes {
		if f(n) {
			walk(n.Nodes, f)
		}
	}
}

// XMLWalk 遍历xml
func XMLWalk(bs []byte) (gin.H, error) {
	buf := bytes.NewBuffer(bs)
	dec := xml.NewDecoder(buf)
	r := make(gin.H)
	var n XMLNode
	err := dec.Decode(&n)
	if err != nil {
		return nil, err
	}
	walk([]XMLNode{n}, func(n XMLNode) bool {
		content := strings.TrimSpace(n.Content)
		if content != "" {
			r[n.XMLName.Local] = n.Content
		}
		return true
	})
	return r, nil
}

// GetSign 获取签名
func GetSign(appSecret string, paramsMap gin.H) string {
	var args []string
	var keys []string
	for k := range paramsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := fmt.Sprintf("%s=%v", k, paramsMap[k])
		args = append(args, v)
	}
	baseString := strings.Join(args, "&")
	baseString += fmt.Sprintf("&key=%s", appSecret)
	data := []byte(baseString)
	r := md5.Sum(data)
	signedString := hex.EncodeToString(r[:])
	return strings.ToUpper(signedString)
}

// CheckSign 检查签名
func CheckSign(appSecret string, paramsMap gin.H) bool {
	noSignMap := gin.H{}
	for k, v := range paramsMap {
		if k != "sign" {
			noSignMap[k] = v
		}
	}
	getSign := GetSign(appSecret, noSignMap)
	if getSign != paramsMap["sign"] {
		return false
	}
	return true
}
