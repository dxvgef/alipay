package notify

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"hash"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/dxvgef/alipay/config"
)

// 校验异步通知的签名
func Verity(alipayConfig *config.Config, req *http.Request) (*Params, error) {
	if err := req.ParseForm(); err != nil {
		return nil, err
	}

	// 解析异步通知参数到结构体
	params, err := parseNotifyParams(req)
	if err != nil {
		return nil, err
	}

	// 校验签名
	if err := veritySign(req.PostForm, alipayConfig); err != nil {
		return nil, err
	}

	return params, nil
}

// 校验签名
func veritySign(values url.Values, alipayConfig *config.Config) error {
	// 解析参数
	sign := values.Get("sign")

	values.Del("sign")
	values.Del("sign_type")

	tmp := values.Encode()

	queryStr, err := url.QueryUnescape(tmp)
	if err != nil {
		return err
	}

	querySlice := strings.Split(queryStr, "&")

	sort.Strings(querySlice)

	data := strings.Join(querySlice, "&")

	// 验证签名
	signData, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return err
	}

	// rsa校验
	var h hash.Hash
	var hType crypto.Hash

	switch alipayConfig.GetAppSignType() {
	case "RSA":
		h = sha1.New()
		hType = crypto.SHA1
	case "RSA2":
		h = sha256.New()
		hType = crypto.SHA256
	default:
		return errors.New("仅支持RSA(SHA1)和RSA2(SHA256)两种签名算法")
	}

	if _, err := h.Write([]byte(data)); err != nil {
		return err
	}
	err = rsa.VerifyPKCS1v15(alipayConfig.GetAlipayPublicKey(), hType, h.Sum(nil), signData)
	if err != nil {
		return err
	}

	return nil
}
