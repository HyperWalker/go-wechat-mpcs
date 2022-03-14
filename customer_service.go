// Package customerservice 微信小程序客服.
/* 目前支持:CheckSignature()验证消息是否来自微信服务器,SendMessage()发送客服消息给用户,UploadTempMedia()把媒体文件上传到微信服务器 */
package customerservice

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"strings"
)

type (
	commonResp struct {
		Errcode int    `json:"errcode,omitempty"`
		Errmsg  string `json:"errmsg,omitempty"`
	}

	// MessageImg 图片消息
	MessageImg struct {
		MediaID string `json:"media_id,omitempty"`
	}

	// MessageText 文本消息
	MessageText struct {
		Content string `json:"content,omitempty"`
	}

	// MessageLink 图文链接
	MessageLink struct {
		// 图文链接消息
		Description string `json:"description,omitempty"`
		// 图文链接消息的图片链接
		ThumbURL string `json:"thumb_url,omitempty"`
		// 消息标题
		Title string `json:"title,omitempty"`
		// 图文链接消息被点击后跳转的链接
		URL string `json:"url,omitempty"`
	}

	// MessageMiniProgramPage 小程序卡片
	MessageMiniProgramPage struct {
		// 消息标题
		Title string `json:"title,omitempty"`
		// 小程序的页面路径
		Pagepath string `json:"pagepath,omitempty"`
		// 小程序消息卡片的封面(media_id)
		ThumbMediaID string `json:"thumb_media_id,omitempty"`
	}

	// SendMessageReq 发送客服消息请求参数
	SendMessageReq struct {
		AccessToken string     `json:"access_token,omitempty"`
		Image       MessageImg `json:"image,omitempty"`
		// 图文链接
		Link MessageLink `json:"link,omitempty"`
		// 小程序卡片
		Miniprogrampage MessageMiniProgramPage `json:"miniprogrampage,omitempty"`
		// 消息类型:text,image,link,miniprogrampage
		Msgtype string `json:"msgtype,omitempty"`
		// 文本消息
		Text MessageText `json:"text,omitempty"`
		// 用户的 OpenID
		Touser string `json:"touser,omitempty"`
	}

	// SendMessageResp 发送客服消息响应参数
	SendMessageResp struct {
		commonResp
	}

	// UploadTempMediaReq 文件上传请求参数.
	UploadTempMediaReq struct {
		AccessToken string `json:"access_token,omitempty"`
		// 图片内容
		Image []byte `json:"image,omitempty"`
	}

	// UploadTempMediaResp 文件上传响应参数
	UploadTempMediaResp struct {
		// 媒体文件上传时间戳
		CreatedAt int64 `json:"created_at,omitempty"`
		// 媒体文件上传后，获取标识，3天内有效
		MediaID string `json:"media_id,omitempty"`
		// 文件类型
		Type string `json:"type,omitempty"`

		commonResp
	}
)

// CheckSignature 验证消息是否来自微信服务器
func CheckSignature(accessToken, nonce, timestamp, signature string) bool {
	qsArr := []string{accessToken, nonce, timestamp}
	sort.Strings(qsArr)
	qs := strings.Join(qsArr, "")

	h := sha1.New()
	h.Write([]byte(qs))
	cipherBuf := h.Sum(nil)

	cipherText := hex.EncodeToString(cipherBuf)
	if cipherText == signature {
		return true
	}
	return false
}

// SendMessage 发送客服消息给用户
func SendMessage(req SendMessageReq) error {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s",
		req.AccessToken)
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var smr SendMessageResp
	err = json.Unmarshal(body, &smr)
	if err != nil {
		return err
	}

	if smr.Errcode == 0 {
		return nil
	}
	return errors.New(smr.Errmsg)
}

// UploadTempMedia 把媒体文件上传到微信服务器
func UploadTempMedia(req UploadTempMediaReq) (utmr UploadTempMediaResp, err error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=image",
		req.AccessToken)

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	header, err := writer.CreateFormFile("media", "temp.png")
	if err != nil {
		return
	}
	header.Write(req.Image)
	writer.Close()

	resp, err := http.Post(url, writer.FormDataContentType(), buf)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &utmr)
	if err != nil {
		return
	}
	if utmr.Errcode == 0 {
		return
	}
	return utmr, errors.New(utmr.Errmsg)
}
