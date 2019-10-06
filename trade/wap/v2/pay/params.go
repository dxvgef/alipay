package pay

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/dxvgef/alipay/config"
)

// API请求地址
const APIURL = "https://openapi.alipay.com/gateway.do"

// Params 公共请求参数
type Params struct {
	alipayConfig     *config.Config // 支付宝应用配置
	AppCertSN        string         // 应用公钥证书SN
	AlipayRootCertSN string         // 支付宝根证书SN
	AppID            string         // 必填，支付宝分配给开发者的应用ID
	Method           string         // 必填，接口名称
	Format           string         // 仅支持"JSON"
	ReturnURL        string         // HTTP/HTTPS开头的URL字符串
	Charset          string         // 必填，请求使用的编码格式，如utf-8,gbk,gb2312等
	SignType         string         // 必填，商户生成签名字符串所使用的签名算法类型，目前支持RSA2和RSA，推荐使用RSA2
	Timestamp        string         // 必填，发送请求的时间，格式"yyyy-MM-dd HH:mm:ss"
	Version          string         // 必填，调用的接口版本，固定为：1.0
	NotifyURL        string         // 支付宝服务器主动通知商户服务器里指定的页面http/https路径。
	AppAuthToken     string         // 详见应用授权概述
	BizContent       *BizContent
	sign             string // 签名字符串
	paramsStr        string // 请求参数拼接成的字符串
	urlValues        url.Values
}

// BizContent 请求参数
type BizContent struct {
	AuthToken          string        `json:"auth_token,omitempty"`           // 针对用户授权接口，获取用户相关数据时，用于标识用户授权关系
	Body               string        `json:"body,omitempty"`                 // 商品说明
	BusinessParams     string        `json:"business_params,omitempty"`      // 商户传入业务信息，具体值要和支付宝约定，应用于安全，营销等参数直传场景，格式为json格式
	DisablePayChannels string        `json:"disable_pay_channels,omitempty"` // 禁用渠道，用户不可用指定渠道支付，当有多个渠道时用“,”分隔，与enable_pay_channels互斥
	EnablePayChannels  string        `json:"enable_pay_channels,omitempty"`  // 可用渠道，用户只能在指定渠道范围内支付，当有多个渠道时用“,”分隔，与disable_pay_channels互斥
	ExtendParams       *ExtendParams `json:"extend_params,omitempty"`        // 业务扩展参数
	ExtUserInfo        *ExtUserInfo  `json:"ext_user_info,omitempty"`        // 外部指定买家
	GoodsType          string        `json:"goods_type,omitempty"`           // 商品主类型 :0-虚拟类商品,1-实物类商品
	MerchantOrderNo    string        `json:"merchant_order_no,omitempty"`    // 商户原始订单号，最大长度限制32位
	OutTradeNo         string        `json:"out_trade_no"`                   // 本地订单号
	PassbackParams     string        `json:"passback_params,omitempty"`      // 公用回传参数，如果请求时传递了该参数，则返回给商户时会回传该参数。支付宝只会在同步返回（包括跳转回商户网站）和异步通知时将该参数原样返回。本参数必须进行UrlEncode之后才可以发送给支付宝。
	ProductCode        string        `json:"product_code"`                   // 销售产品码，商家和支付宝签约的产品码，移动网站支付2.0的值是UICK_WAP_WAY
	PromoParams        string        `json:"promo_params,omitempty"`         // 优惠参数，仅与支付宝协商后可用
	QuitURL            string        `json:"quit_url,omitempty"`             // 用户付款中途退出返回商户网站的地址
	SpecifiedChannel   string        `json:"specified_channel,omitempty"`    // 指定渠道，目前仅支持传入pcredit，若由于用户原因渠道不可用，用户可选择是否用其他渠道支付
	StoreID            string        `json:"store_id,omitempty"`             // 商户门店编号
	Subject            string        `json:"subject"`                        // 商品标题
	TimeExpire         string        `json:"time_expire,omitempty"`          // 绝对超时时间，格式为yyyy-MM-dd HH:mm
	TimeoutExpress     string        `json:"timeout_express,omitempty"`      // 该笔订单允许的最晚付款时间，逾期将关闭交易。取值范围：1m～15d。m-分钟，h-小时，d-天，1c-当天（1c-当天的情况下，无论交易何时创建，都在0点关闭）。 该参数数值不接受小数点， 如 1.5h，可转换为 90m。
	TotalAmount        float32       `json:"total_amount"`                   // 订单总金额，单位为元，精确到小数点后两位
}

// ExtendParams // 业务扩展参数
type ExtendParams struct {
	HbFqNum              string `json:"hb_fq_num,omitempty"`               // 花呗分期数（目前仅支持3、6、12）注：使用该参数需要仔细阅读“花呗分期接入文档”
	HbFqSellerPercent    string `json:"hb_fq_seller_percent,omitempty"`    // 使用花呗分期卖家承担收费比例，商家承担手续费传入100，用户承担手续费传入0，仅支持传入100、0两种，其他比例暂不支持注：使用该参数需要仔细阅读“花呗分期接入文档”
	NeedBuyerRealnamed   string `json:"need_buyer_realnamed,omitempty"`    // 是否发起实名校验T：发起F：不发起
	SysServiceProviderID string `json:"sys_service_provider_id,omitempty"` // 系统商编号，该参数作为系统商返佣数据提取的依据，请填写系统商签约协议的PID
	TransMemo            string `json:"trans_memo,omitempty"`              // 账务备注：该字段显示在离线账单的账务备注中
}

// ExtUserInfo 外部指定买家
type ExtUserInfo struct {
	CertNo        string `json:"cert_no,omitempty"`         // 证件号，need_check_info=T时该参数才有效
	CertType      string `json:"cert_type,omitempty"`       // need_check_info=T时该参数才有效。身份证：IDENTITY_CARD、护照：PASSPORT、军官证：OFFICER_CARD、士兵证：SOLDIER_CARD、户口本：HOKOU等。如有其它类型需要支持，请与蚂蚁金服工作人员联系
	MinAge        string `json:"min_age,omitempty"`         // 允许的最小买家年龄，买家年龄必须大于等于所传数值，need_check_info=T时该参数才有效，min_age为整数，必须大于等于0
	Mobile        string `json:"mobile,omitempty"`          // 手机号，该参数暂不校验
	Name          string `json:"name,omitempty"`            // 姓名，need_check_info=T时该参数才有效
	NeedCheckInfo string `json:"need_check_info,omitempty"` // 是否强制校验身份信息，T:强制校验 / F：不强制
	FixBuyer      string `json:"fix_buyer,omitempty"`       // 是否强制校验付款人身份信息，T:强制校验 / F：不强制
}

// 检查持续时间参数值
func checkDuration(value string) bool {
	unit := value[len(value)-1:]
	if unit != "d" && unit != "m" && unit != "h" && unit != "c" {
		return false
	}
	v := value[:len(value)-1]
	i, _ := strconv.ParseUint(v, 10, 32)
	return i == 0
}

// 检测公用请求参数
func (r *Params) checkParams() error {
	if r.AppID == "" {
		return errors.New("AppID参数未赋值")
	}
	appID, _ := strconv.ParseInt(r.AppID, 10, 64)
	if appID == 0 {
		return errors.New("AppID参数值只能是数字")
	}
	if r.Format != "" && r.Format != "JSON" {
		return errors.New("Format参数值只能是JSON")
	}
	if r.ReturnURL != "" {
		returnUrlLen := len(r.ReturnURL)
		if returnUrlLen < 8 {
			return errors.New("ReturnURL参数值必须是http://或https://开头")
		}
		if returnUrlLen > 256 {
			return errors.New("ReturnURL参数值长度不能大于256")
		}
		if r.ReturnURL[0:7] != "http://" && r.ReturnURL[0:8] != "https://" {
			return errors.New("ReturnURL参数值必须是http://或https://开头")
		}
	}
	if r.Method == "" {
		r.Method = "alipay.trade.wap.pay"
	}
	if r.Charset == "" {
		return errors.New("Charset参数未赋值")
	}
	if len(r.Charset) > 10 {
		return errors.New("Charset参数值的长度不能大于10")
	}
	if r.SignType != "RSA" && r.SignType != "RSA2" {
		return errors.New("SignType参数值必须是RSA或RSA2")
	}
	_, err := time.Parse("2006-01-02 15:04:05", r.Timestamp)
	if err != nil {
		return errors.New("Timestamp的参数值格式不正确")
	}
	if r.Version != "1" && r.Version != "1.0" {
		return errors.New("Version的参数值必须是1或者1.0")
	}
	if r.NotifyURL != "" {
		notifyUrlLen := len(r.NotifyURL)
		if notifyUrlLen < 8 {
			return errors.New("NotifyURL参数值必须是http://或https://开头")
		}
		if notifyUrlLen > 256 {
			return errors.New("NotifyURL参数值长度不能大于256")
		}
		if r.NotifyURL[0:7] != "http://" && r.NotifyURL[0:8] != "https://" {
			return errors.New("NotifyURL参数值必须是http://或https://开头")
		}
	}
	return nil
}

// 检查biz_content参数
func (r *Params) checkBizContent() error {
	if r.BizContent == nil {
		r.BizContent = &BizContent{
			ProductCode: "QUICK_WAP_WAY",
		}
	}
	if len(r.BizContent.Body) > 128 {
		return errors.New("BizContent.Body参数值的长度不能大于128")
	}
	if r.BizContent.Subject == "" {
		return errors.New("BizContent.Subject参数未赋值")
	}
	if len(r.BizContent.Subject) > 256 {
		return errors.New("BizContent.Subject参数值的长度不能大于256")
	}
	if r.BizContent.OutTradeNo == "" {
		return errors.New("BizContent.OutTradeNo参数未赋值")
	}
	if len(r.BizContent.OutTradeNo) > 64 {
		return errors.New("BizContent.OutTradeNo参数值的长度不能大于64")
	}
	if r.BizContent.TimeoutExpress != "" && !checkDuration(r.BizContent.TimeoutExpress) {
		return errors.New("BizContent.TimeoutExpress参数值值的格式不正确")
	}
	if r.BizContent.TimeExpire != "" {
		_, err := time.Parse("2006-01-02 15:04:05", r.BizContent.TimeExpire)
		if err != nil {
			return errors.New("BizContent.TimeExpire的参数值格式不正确")
		}
	}
	if r.BizContent.TotalAmount < 0.01 || r.BizContent.TotalAmount > 100000000 {
		return errors.New("BizContent.TotalAmount参数值的范围必须是0.01-100000000")
	}
	if r.BizContent.ProductCode != "QUICK_WAP_WAY" {
		return errors.New("BizContent.ProductCod参数值的范围必须是QUICK_WAP_WAY")
	}
	if r.BizContent.GoodsType != "0" && r.BizContent.GoodsType != "1" {
		return errors.New("BizContent.GoodsType参数值只能是0或1")
	}
	if len(r.BizContent.PassbackParams) > 512 {
		return errors.New("BizContent.PassbackParams参数值的长度不能大于512")
	}
	if r.BizContent.PromoParams != "" {
		if len(r.BizContent.PromoParams) > 512 {
			return errors.New("BizContent.PromoParams参数值的长度不能大于512")
		}
		var raw json.RawMessage
		if json.Unmarshal([]byte(r.BizContent.PromoParams), &raw) != nil {
			return errors.New("BizContent.PromoParams参数值必须是有效的JSON格式")
		}
		r.BizContent.PassbackParams = url.QueryEscape(r.BizContent.PassbackParams)
	}
	if r.BizContent.EnablePayChannels != "" && r.BizContent.DisablePayChannels != "" {
		return errors.New("BizContent.EnablePayChannels与BizContent.DisablePayChannels参数互斥，只能使用其中一个")
	}
	if r.BizContent.EnablePayChannels != "" {
		if len(r.BizContent.EnablePayChannels) > 128 {
			return errors.New("BizContent.EnablePayChannels参数值的长度不能大于128")
		}
	}
	if r.BizContent.DisablePayChannels != "" {
		if len(r.BizContent.DisablePayChannels) > 128 {
			return errors.New("BizContent.DisablePayChannels参数值的长度不能大于128")
		}
	}
	if r.BizContent.QuitURL != "" {
		if len(r.BizContent.QuitURL) > 400 {
			return errors.New("BizContent.QuitURL参数值的长度不能大于400")
		}
	}
	return nil
}

// 检查extend_params参数
func (r *Params) checkExtendParams() error {
	if r.BizContent.ExtendParams == nil {
		return nil
	}
	if r.BizContent.ExtendParams.SysServiceProviderID != "" {
		if len(r.BizContent.ExtendParams.SysServiceProviderID) > 64 {
			return errors.New("BizContent.ExtendParams.SysServiceProviderID参数值的长度不能大于64")
		}
	}
	if r.BizContent.ExtendParams.NeedBuyerRealnamed != "" {
		if r.BizContent.ExtendParams.NeedBuyerRealnamed != "T" && r.BizContent.ExtendParams.NeedBuyerRealnamed != "F" {
			return errors.New("BizContent.ExtendParams.NeedBuyerRealnamed参数值只能是T或F")
		}
	}
	if r.BizContent.ExtendParams.TransMemo != "" {
		if len(r.BizContent.ExtendParams.TransMemo) > 128 {
			return errors.New("BizContent.ExtendParams.TransMemo参数值的长度不能大于128")
		}
	}
	if r.BizContent.ExtendParams.HbFqNum != "" {
		if r.BizContent.ExtendParams.HbFqNum != "3" && r.BizContent.ExtendParams.HbFqNum != "6" && r.BizContent.ExtendParams.HbFqNum != "12" {
			return errors.New("BizContent.ExtendParams.HbFqNum参数值只能是3、6、12")
		}
	}
	if r.BizContent.ExtendParams.HbFqSellerPercent != "" {
		if r.BizContent.ExtendParams.HbFqSellerPercent != "100" && r.BizContent.ExtendParams.HbFqSellerPercent != "0" {
			return errors.New("BizContent.ExtendParams.HbFqSellerPercent参数值只能是0或199")
		}
	}
	return nil
}

// 检查extend_user_info参数
func (r *Params) checkExtendUserInfo() error {
	if r.BizContent.ExtUserInfo == nil {
		return nil
	}
	if r.BizContent.ExtUserInfo.NeedCheckInfo != "" && r.BizContent.ExtUserInfo.NeedCheckInfo != "T" && r.BizContent.ExtUserInfo.NeedCheckInfo != "F" {
		return errors.New("BizContent.ExtUserInfo.NeedCheckInfo参数值只能是T或F")
	}
	if r.BizContent.ExtUserInfo.Name != "" && r.BizContent.ExtUserInfo.NeedCheckInfo == "T" {
		if len(r.BizContent.ExtUserInfo.Name) > 16 {
			return errors.New("BizContent.ExtendUserInfo.Name参数值的长度不能大于16")
		}
	}
	if r.BizContent.ExtUserInfo.CertType != "" && r.BizContent.ExtUserInfo.NeedCheckInfo == "T" {
		if len(r.BizContent.ExtUserInfo.CertType) > 32 {
			return errors.New("BizContent.ExtendUserInfo.CertType参数值的长度不能大于32")
		}
	}
	if r.BizContent.ExtUserInfo.CertNo != "" && r.BizContent.ExtUserInfo.NeedCheckInfo == "T" {
		if len(r.BizContent.ExtUserInfo.CertNo) > 64 {
			return errors.New("BizContent.ExtendUserInfo.CertNo参数值的长度不能大于64")
		}
	}
	if r.BizContent.ExtUserInfo.MinAge != "" && r.BizContent.ExtUserInfo.NeedCheckInfo == "T" {
		minAge, err := strconv.ParseUint(r.BizContent.ExtUserInfo.MinAge, 10, 32)
		if err != nil {
			return errors.New("BizContent.ExtendUserInfo.MinAge参数值必须是大于或等于0的整数")
		}
		if minAge == 0 {
			return errors.New("BizContent.ExtendUserInfo.MinAge参数值必须是大于或等于0的整数")
		}
	}
	if r.BizContent.ExtUserInfo.FixBuyer != "" {
		if r.BizContent.ExtUserInfo.FixBuyer != "T" && r.BizContent.ExtUserInfo.FixBuyer != "F" {
			return errors.New("BizContent.ExtUserInfo.FixBuyer参数值只能是T或F")
		}
	}
	return nil
}
