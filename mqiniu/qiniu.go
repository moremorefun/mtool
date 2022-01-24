package mqiniu

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/moremorefun/mtool/mlog"
	"github.com/parnurzeal/gorequest"
	"image"
	"time"

	"github.com/disintegration/imaging"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type StRespTextCensor struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Result  struct {
		Scenes struct {
			Antispam struct {
				Details []struct {
					Score float64 `json:"score"`
					Label string  `json:"label"`
				} `json:"details"`
				Suggestion string `json:"suggestion"`
			} `json:"antispam"`
		} `json:"scenes"`
		Suggestion string `json:"suggestion"`
	} `json:"result"`
}

type StRespImageCensor struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Result  struct {
		Scenes struct {
			Terror struct {
				Details []struct {
					Score      float64 `json:"score"`
					Suggestion string  `json:"suggestion"`
					Label      string  `json:"label"`
				} `json:"details"`
				Suggestion string `json:"suggestion"`
			} `json:"terror"`
			Politician struct {
				Suggestion string `json:"suggestion"`
			} `json:"politician"`
			Pulp struct {
				Details []struct {
					Score      float64 `json:"score"`
					Suggestion string  `json:"suggestion"`
					Label      string  `json:"label"`
				} `json:"details"`
				Suggestion string `json:"suggestion"`
			} `json:"pulp"`
		} `json:"scenes"`
		Suggestion string `json:"suggestion"`
	} `json:"result"`
}

// Upload 上传到qiniu
func Upload(ctx context.Context, access string, secret string, zone *storage.Zone, bucket string, fileKey string, bs []byte) error {
	buf := bytes.NewBuffer(bs)
	mac := qbox.NewMac(
		access,
		secret,
	)
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = zone
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}
	retry := 0
GotoUpload:
	err := formUploader.Put(ctx, &ret, upToken, fileKey, buf, int64(buf.Len()), &putExtra)
	if err != nil {
		qiniuErr, ok := err.(*storage.ErrorInfo)
		if ok {
			if qiniuErr.Code == 614 {
				// file exists
				return nil
			}
		}
		retry++
		if retry < 3 {
			// 重试
			goto GotoUpload
		}
		return err
	}
	return nil
}

// GetDownloadURL 获取私有下载链接
func GetDownloadURL(access string, secret string, domain string, fileKey string, deadline int64) string {
	mac := qbox.NewMac(access, secret)

	// 私有空间访问
	privateAccessURL := storage.MakePrivateURL(mac, domain, fileKey, deadline)
	return privateAccessURL
}

// UploadToken 获取上传token
func UploadToken(access string, secret string, bucket string) string {
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	putPolicy.Expires = 7200 //示例2小时有效期
	mac := qbox.NewMac(access, secret)
	upToken := putPolicy.UploadToken(mac)
	return upToken
}

// UploadImg 上传到qiniu
func UploadImg(ctx context.Context, access string, secret string, bucket string, fileKey string, img image.Image) error {
	buf := bytes.NewBuffer(make([]byte, 0))
	err := imaging.Encode(buf, img, imaging.JPEG, imaging.JPEGQuality(100))
	if err != nil {
		return err
	}

	mac := qbox.NewMac(
		access,
		secret,
	)
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuanan
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}
	retry := 0
GotoUpload:
	err = formUploader.Put(ctx, &ret, upToken, fileKey, buf, int64(buf.Len()), &putExtra)
	if err != nil {
		qiniuErr, ok := err.(*storage.ErrorInfo)
		if ok {
			if qiniuErr.Code == 614 {
				// file exists
				return nil
			}
		}
		retry++
		if retry < 3 {
			// 重试
			goto GotoUpload
		}
		return err
	}
	return nil
}

// TextCensor 脏字过滤
func TextCensor(access string, secret string, text string) (*StRespTextCensor, error) {
	mac := qbox.NewMac(access, secret)
	query := gorequest.New().
		Post("https://ai.qiniuapi.com/v3/text/censor").
		Send(gin.H{
			"data": gin.H{
				"text": text,
			},
			"params": gin.H{
				"scenes": []string{"antispam"},
			},
		}).
		Timeout(time.Second * 10)
	req, err := query.MakeRequest()
	if err != nil {
		return nil, err
	}
	token, err := mac.SignRequestV2(req)
	if err != nil {
		return nil, err
	}
	query.AppendHeader("Authorization", "Qiniu "+token)
	_, respBody, errs := query.EndBytes()
	if errs != nil {
		return nil, err
	}
	mlog.Log.Debugf("TextCensor resp: %s", respBody)
	var respObj StRespTextCensor
	err = json.Unmarshal(respBody, &respObj)
	if errs != nil {
		return nil, err
	}
	return &respObj, nil
}

// ImageCensor 脏字过滤
func ImageCensor(access string, secret string, uri string) (*StRespImageCensor, error) {
	mac := qbox.NewMac(access, secret)
	query := gorequest.New().
		Post("https://ai.qiniuapi.com/v3/image/censor").
		Send(gin.H{
			"data": gin.H{
				"uri": uri,
			},
			"params": gin.H{
				"scenes": []string{"pulp", "terror", "politician", "ads"},
			},
		}).
		Timeout(time.Second * 10)
	req, err := query.MakeRequest()
	if err != nil {
		return nil, err
	}
	token, err := mac.SignRequestV2(req)
	if err != nil {
		return nil, err
	}
	query.AppendHeader("Authorization", "Qiniu "+token)
	_, respBody, errs := query.EndBytes()
	if errs != nil {
		return nil, err
	}
	mlog.Log.Debugf("ImageCensor resp: %s", respBody)
	var respObj StRespImageCensor
	err = json.Unmarshal(respBody, &respObj)
	if errs != nil {
		return nil, err
	}
	return &respObj, nil
}
