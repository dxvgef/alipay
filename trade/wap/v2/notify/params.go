package notify

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

// 异步通知参数
type Params struct {
	NotifyTime        string              // 通知的发送时间，格式为yyyy-MM-dd HH:mm:ss
	NotifyType        string              // 通知的类型
	NotifyID          string              // 通知校验ID
	AppID             string              // 支付宝分配给开发者的应用ID
	Charset           string              // 编码格式，如utf-8、gbk、gb2312等
	Version           string              // 接口版本，固定为1.0，与接口版本号不同步，目前接口版本号是2.0
	SignType          string              // 商户生成签名字符串所使用的签名算法类型，目前支持RSA2和RSA，推荐使用RSA2
	Sign              string              // 签名
	TradeNo           string              // 支付宝交易凭证号
	OutTradeNo        string              // 商户订单号
	OutBizNo          string              // 商户业务ID，主要是退款通知中返回退款申请的流水号
	BuyerID           string              // 买家支付宝账号对应的支付宝唯一用户号，以2088开头的纯16位数字
	BuyerLogonID      string              // 买家支付宝账号
	SellerID          string              // 卖家支付宝用户号
	SellerEmail       string              // 卖家支付宝账号
	TradeStatus       string              // 交易目前所处的状态
	TotalAmount       float64             // 订单金额，单位为元，取值范围为[0.01，100000000.00]，精确到小数点后两位
	ReceiptAmount     float64             // 商家在交易中实际收到的款项，单位为元，取值范围为[0.01，100000000.00]，精确到小数点后两位
	InvoiceAmount     float64             // 用户在交易中支付的可开发票的金额，单位为元，取值范围为[0.01，100000000.00]，精确到小数点后两位
	BuyerPayAmount    float64             // 用户在交易中支付的金额，单位为元，取值范围为[0.01，100000000.00]，精确到小数点后两位
	PointAmount       float64             // 使用集分宝支付的金额，单位为元，取值范围为[0.01，100000000.00]，精确到小数点后两位
	RefundFee         float64             // 退款通知中，返回总退款金额，单位为元，取值范围为[0.01，100000000.00]，精确到小数点后两位
	Subject           string              // 商品的标题/交易标题/订单标题/订单关键字等，是请求时对应的参数，原样通知回来
	Body              string              // 该订单的备注、描述、明细等。对应请求时的body参数，原样通知回来
	GmtCreate         string              // 该笔交易创建的时间。格式为yyyy-MM-dd HH:mm:ss
	GmtPayment        string              // 该笔交易的买家付款时间。格式为yyyy-MM-dd HH:mm:ss
	GmtRefund         string              // 该笔交易的退款时间。格式为yyyy-MM-dd HH:mm:ss
	GmtClose          string              // 该笔交易结束时间。格式为yyyy-MM-dd HH:mm:ss
	FundBillList      []FundBillList      // 支付成功的各个渠道金额信息，详见资金明细信息说明
	PassbackParams    string              // 公共回传参数，如果请求时传递了该参数，则返回给商户时会在异步通知时将该参数原样返回，本参数必须进行UrlEncode之后才可以发送给支付宝
	VoucherDetailList []VoucherDetailList // 本交易支付时所使用的所有优惠券信息，详见优惠券信息说明
}

// 支付渠道信息
type FundBillList struct {
	FundChannel string `json:"fund_channel,omitempty"` // 支付渠道
	Amount      string `json:"amount,omitempty"`       // 使用指定支付渠道支付的金额，单位为元，取值范围为[0.01，100000000.00]，精确到小数点后两位
}

// 优惠券信息
type VoucherDetailList struct {
	Name               string  `json:"name"`                          // 券名称
	Type               string  `json:"type"`                          // 券类型
	Amount             float64 `json:"amount"`                        // 优惠券面额，它应该等于商家出资加上其他出资方出资
	MerchantContribute float64 `json:"merchant_contribute,omitempty"` // 商家出资，特指发起交易的商家出资金额，单位为元，取值范围为[0.01，100000000.00]，精确到小数点后两位
	OtherContribute    float64 `json:"other_contribute,omitempty"`    // 其他出资方出资金额，可能是支付宝，可能是品牌商，或者其他方，也可能是他们的共同出资
	Memo               string  `json:"memo,omitempty"`                // 备注信息
}

// 解析异步通知参数到结构体
func parseNotifyParams(req *http.Request) (*Params, error) {
	var err error
	var params Params
	params.NotifyTime = req.PostFormValue("notify_time")
	params.NotifyType = req.PostFormValue("notify_type")
	params.NotifyID = req.PostFormValue("notify_id")
	params.AppID = req.PostFormValue("app_id")
	params.Charset = req.PostFormValue("charset")
	params.Version = req.PostFormValue("version")
	params.SignType = req.PostFormValue("sign_type")
	params.Sign = req.PostFormValue("sign")
	params.TradeNo = req.PostFormValue("trade_no")
	params.OutTradeNo = req.PostFormValue("out_trade_no")
	params.OutBizNo = req.PostFormValue("out_biz_no")
	params.BuyerID = req.PostFormValue("buyer_id")
	params.BuyerLogonID = req.PostFormValue("buyer_logon_id")
	params.SellerID = req.PostFormValue("seller_id")
	params.SellerEmail = req.PostFormValue("seller_email")
	params.TradeStatus = req.PostFormValue("trade_status")
	if req.PostFormValue("total_amount") != "" {
		params.TotalAmount, err = strconv.ParseFloat(req.PostFormValue("total_amount"), 64)
		if err != nil {
			return nil, err
		}
	}
	if req.PostFormValue("receipt_amount") != "" {
		params.ReceiptAmount, err = strconv.ParseFloat(req.PostFormValue("receipt_amount"), 64)
		if err != nil {
			return nil, err
		}
	}
	if req.PostFormValue("invoice_amount") != "" {
		params.InvoiceAmount, err = strconv.ParseFloat(req.PostFormValue("invoice_amount"), 64)
		if err != nil {
			return nil, err
		}
	}
	if req.PostFormValue("buyer_pay_amount") != "" {
		params.BuyerPayAmount, err = strconv.ParseFloat(req.PostFormValue("buyer_pay_amount"), 64)
		if err != nil {
			return nil, err
		}
	}
	if req.PostFormValue("point_amount") != "" {
		params.PointAmount, err = strconv.ParseFloat(req.PostFormValue("point_amount"), 64)
		if err != nil {
			return nil, err
		}
	}
	if req.PostFormValue("refund_fee") != "" {
		params.RefundFee, err = strconv.ParseFloat(req.PostFormValue("refund_fee"), 64)
		if err != nil {
			return nil, err
		}
	}
	params.Subject = req.PostFormValue("subject")
	params.Body = req.PostFormValue("body")
	params.GmtCreate = req.PostFormValue("gmt_create")
	params.GmtPayment = req.PostFormValue("gmt_payment")
	params.GmtRefund = req.PostFormValue("gmt_refund")
	params.GmtClose = req.PostFormValue("gmt_close")
	if req.PostFormValue("fund_bill_list") != "" {
		if err := json.Unmarshal([]byte(req.PostFormValue("fund_bill_list")), &params.FundBillList); err != nil {
			return nil, err
		}
	}
	if req.PostFormValue("passback_params") != "" {
		if params.PassbackParams, err = url.QueryUnescape(req.PostFormValue("passback_params")); err != nil {
			return nil, err
		}
	}
	if req.PostFormValue("voucher_detail_list") != "" {
		if err := json.Unmarshal([]byte(req.PostFormValue("voucher_detail_list")), &params.VoucherDetailList); err != nil {
			return nil, err
		}
	}

	return &params, nil
}
