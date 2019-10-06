# dxvgef/alipay
支付宝API For Golang

## 已实现功能：
- 手机网站支付 - 生成支付链接
- 手机网站支付 - 异步通知验证

#### 手机网站支付示例
```go
package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	AlipayConfig "github.com/dxvgef/alipay/config"
	alipayWapNotify "github.com/dxvgef/alipay/trade/wap/v2/notify"
	alipayWapPay "github.com/dxvgef/alipay/trade/wap/v2/pay"
)

var appID = ""  // 应用ID
var alipayRootCertPath = "" // 支付宝根证书文件路径
var alipayCertPublicKeyPath = ""    // 支付宝公钥证书文件路径
var appCertPublicKeyPath = ""   // 应用公钥证书文件路径
var appPKCS1PrivateKey = "" // 应用PKCS1私钥的字符串
var appPKCS8PrivateKey = "" // 应用PKCS8私钥的字符串

// 支付宝配置实例
var alipayConfig AlipayConfig.Config

func main() {
	log.SetFlags(log.Lshortfile)

	// 加载支付宝根证书文件
	if err := alipayConfig.LoadAlipayRootCert(alipayRootCertPath); err != nil {
		log.Println(err.Error())
		return
	}
	// 加载支付宝公钥证书文件
	if err := alipayConfig.LoadAlipayCertPublicKey(alipayCertPublicKeyPath); err != nil {
		log.Println(err.Error())
		return
	}
	// 加载应用公钥证书
	if err := alipayConfig.LoadAppCertPublicKey(appCertPublicKeyPath); err != nil {
		log.Println(err.Error())
		return
	}
	// 设置应用私钥字符串
	if err := alipayConfig.SetAppPrivateKey(appPKCS8PrivateKey); err != nil {
		log.Println(err.Error())
		return
	}
	// 设置签名类型
	if err := alipayConfig.SetAppSignType("RSA2"); err != nil {
		log.Println(err.Error())
		return
	}
	if err := alipayConfig.SetAppID(appID); err != nil {
		log.Println(err.Error())
	}

	log.Println("支付宝网关基本参数设置成功")

	handler()
}

func handler() {
    // 发起支付
	http.HandleFunc("/pay", func(resp http.ResponseWriter, req *http.Request) {

		wapPay, err := alipayWapPay.New(&alipayConfig)
		if err != nil {
			resp.WriteHeader(500)
			resp.Write([]byte(err.Error()))
			return
		}
		wapPay.ReturnURL = "http://yourdomain/return"
		wapPay.NotifyURL = "http://yourdomain/notify"
		wapPay.BizContent.Subject = "商品名称"
		wapPay.BizContent.OutTradeNo = strconv.FormatInt(time.Now().Unix(), 10)
		wapPay.BizContent.TotalAmount = 0.01
		wapPay.BizContent.GoodsType = "0"
		if wapPay.SignByCert() != nil {
			resp.WriteHeader(500)
			resp.Write([]byte(err.Error()))
			return
		}

		html := `<a href="` + wapPay.GetURL() + `">立即支付</a>`
		resp.WriteHeader(200)
		resp.Write([]byte(html))
	})

    // 支付结果同步跳转
	http.HandleFunc("/return", func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(200)
		resp.Write([]byte("充值完成"))
	})

    // 支付结果异步通知
	http.HandleFunc("/notify", func(resp http.ResponseWriter, req *http.Request) {
		_, err := alipayWapNotify.Verity(&alipayConfig, req)
		if err != nil {
			log.Println(err.Error())
			return
		}
		resp.WriteHeader(200)
		resp.Write([]byte("success"))
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Println(err.Error())
		return
    }
}
```