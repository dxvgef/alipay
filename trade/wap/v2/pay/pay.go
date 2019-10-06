package pay

import (
	"errors"
	"time"

	"github.com/dxvgef/alipay/config"
)

// 生成一个新的默认请求参数
func New(alipayConfig *config.Config) (*Params, error) {
	if alipayConfig.GetAppID() == "" {
		return nil, errors.New("未设置支付宝配置的AppID参数值")
	}
	if alipayConfig.GetAppSignType() == "" {
		return nil, errors.New("未设置支付宝配置的AppSignType参数值")
	}
	if alipayConfig.GetAppCertPublicKeySN() == "" {
		return nil, errors.New("未设置支付宝配置的应用证书")
	}
	if alipayConfig.GetAlipayRootCertSN() == "" {
		return nil, errors.New("未设置支付宝配置的根证书")
	}
	return &Params{
		alipayConfig:     alipayConfig,
		AppID:            alipayConfig.GetAppID(),
		Method:           "alipay.trade.wap.pay",
		Charset:          "utf-8",
		SignType:         alipayConfig.GetAppSignType(),
		Timestamp:        time.Now().Format("2006-01-02 15:04:05"),
		Version:          "1.0",
		AppCertSN:        alipayConfig.GetAppCertPublicKeySN(),
		AlipayRootCertSN: alipayConfig.GetAlipayRootCertSN(),
		BizContent: &BizContent{
			ProductCode:        "QUICK_WAP_WAY",
			PromoParams:        "",
			ExtendParams:       nil,
			MerchantOrderNo:    "",
			EnablePayChannels:  "",
			DisablePayChannels: "",
			StoreID:            "",
			SpecifiedChannel:   "",
			BusinessParams:     "",
			ExtUserInfo:        nil,
		},
	}, nil
}

// GetParamsStr 获得参数字符串
func (r *Params) GetParamsStr() string {
	return r.paramsStr
}

// GetURL 获得URL编码后的参数字符串
func (r *Params) GetURL() string {
	r.urlValues.Add("sign", r.sign)
	return APIURL + "?" + r.urlValues.Encode()
}
