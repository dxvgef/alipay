package pay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"net/url"
)

// 使用公钥文件生成签名
func (self *Params) SignByCert() error {
	var err error
	if err = self.checkParams(); err != nil {
		return err
	}
	if err = self.checkBizContent(); err != nil {
		return err
	}
	if err = self.checkExtendParams(); err != nil {
		return err
	}
	if err = self.checkExtendUserInfo(); err != nil {
		return err
	}

	// 获得应用公钥SN
	// self.AppCertSN = alipayConfig.GetAppCertPublicKeySN()
	if self.AppCertSN == "" {
		return errors.New("无法获取配置中的应用公钥SN")
	}

	// 获得支付宝根证书SN
	// self.AlipayRootCertSN = alipayConfig.GetAlipayRootCertSN()
	if self.AlipayRootCertSN == "" {
		return errors.New("无法获取配置中的支付宝根证书SN")
	}

	// 构建参数字符串
	if err = self.buildParamsStr(); err != nil {
		return err
	}

	// 解析私钥并生成签名
	self.sign, err = self.makeSign()
	if err != nil {
		return err
	}

	return nil
}

// 构建请求参数字符串
func (r *Params) buildParamsStr() error {
	var err error

	r.urlValues = make(url.Values)

	bizContent, err := json.Marshal(r.BizContent)
	if err != nil {
		return errors.New("biz_content参数值序列化成JSON时失败：" + err.Error())
	}
	bizContentStr := string(bizContent)

	r.urlValues.Add("alipay_root_cert_sn", r.AlipayRootCertSN)
	r.urlValues.Add("app_cert_sn", r.AppCertSN)
	r.urlValues.Add("app_id", r.AppID)
	r.urlValues.Add("biz_content", bizContentStr)
	r.urlValues.Add("charset", r.Charset)

	var s string
	s = "alipay_root_cert_sn=" + r.AlipayRootCertSN + "&app_cert_sn=" + r.AppCertSN + "&app_id=" + r.AppID + "&biz_content=" + bizContentStr + "&charset=" + r.Charset
	if r.Format != "" {
		r.urlValues.Add("format", r.Format)
		s += "&format=" + r.Format
	}
	s += "&method=" + r.Method
	r.urlValues.Add("method", r.Method)
	if r.NotifyURL != "" {
		r.urlValues.Add("notify_url", r.NotifyURL)
		s += "&notify_url=" + r.NotifyURL
	}
	if r.ReturnURL != "" {
		r.urlValues.Add("return_url", r.ReturnURL)
		s += "&return_url=" + r.ReturnURL
	}
	r.urlValues.Add("sign_type", r.SignType)
	r.urlValues.Add("timestamp", r.Timestamp)
	r.urlValues.Add("version", r.Version)
	s += "&sign_type=" + r.SignType + "&timestamp=" + r.Timestamp + "&version=" + r.Version

	r.paramsStr = s
	return nil
}

// 使用密钥生成签名
func (self *Params) makeSign() (string, error) {
	if self.paramsStr == "" {
		return "", errors.New("签名参数未构建")
	}
	var h hash.Hash
	var hType crypto.Hash

	switch self.alipayConfig.GetAppSignType() {
	case "RSA":
		h = sha1.New()
		hType = crypto.SHA1
	case "RSA2":
		h = sha256.New()
		hType = crypto.SHA256
	default:
		return "", errors.New("仅支持RSA(SHA1)和RSA2(SHA256)两种签名算法")
	}

	// 生成签名
	_, err := h.Write([]byte(self.paramsStr))
	if err != nil {
		return "", err
	}
	bytes := h.Sum(nil)
	signBytes, err := rsa.SignPKCS1v15(rand.Reader, self.alipayConfig.GetAppPrivateKey(), hType, bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signBytes), nil
}
