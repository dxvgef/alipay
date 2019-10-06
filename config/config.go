package config

import (
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"strings"

	"github.com/dxvgef/gommon/encrypt"
)

// 支付宝参数配置
type Config struct {
	appID string // 应用ID
	// alipayRootCert   *x509.Certificate // 支付宝根证书
	alipayRootCertSN string         // 根证书SN的MD5值
	alipayPublicKey  *rsa.PublicKey // 支付宝公钥

	appPublicKey       *rsa.PublicKey  // 应用公钥
	appCertPublicKeySN string          // 应用公钥证书SN的MD5值
	appPrivateKey      *rsa.PrivateKey // 应用私钥
	appPrivateKeyType  string          // 应用私钥类型
	appSignType        string          // 应用签名类型RSA/RSA2
}

// 加载支付宝根证书文件
func (obj *Config) LoadAlipayRootCert(filePath string) error {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 解析PEM块
	blocks := encrypt.ParsePEMBlocks(fileData)
	if blocks == nil {
		return errors.New("支付宝根证书数据格式无效")
	}

	// 计算根证书的SN
	var SNSlice []string
	for k := range blocks {
		sn, err := parseCert(blocks[k].Bytes)
		if err != nil {
			return err
		}
		if sn != "" {
			SNSlice = append(SNSlice, sn)
		}
	}

	if len(SNSlice) == 0 {
		return errors.New("支付宝根证书的SN计算失败")
	}
	obj.alipayRootCertSN = strings.Join(SNSlice, "_")

	return nil
}

// 加载支付宝公钥证书文件
func (obj *Config) LoadAlipayCertPublicKey(filePath string) error {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	blocks := encrypt.ParsePEMBlocks(fileData)
	if blocks == nil {
		return errors.New("支付宝公钥证书格式无效")
	}

	publicKey, err := encrypt.ParseRSAPublicKey(blocks[0].Bytes)
	if err != nil {
		return err
	}

	obj.alipayPublicKey = publicKey

	return nil
}

// 加载应用公钥证书文件
func (obj *Config) LoadAppCertPublicKey(filePath string) error {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	blocks := encrypt.ParsePEMBlocks(fileData)
	if blocks == nil {
		return errors.New("支付宝应用公钥证书格式无效")
	}

	publicKey, err := encrypt.ParseRSAPublicKey(blocks[0].Bytes)
	if err != nil {
		return err
	}

	obj.appPublicKey = publicKey

	// 计算应用公钥的SN
	cert, err := x509.ParseCertificate(blocks[0].Bytes)
	if err != nil {
		return err
	}
	sn, err := encrypt.MD5ByStrings([]string{cert.Issuer.String(), cert.SerialNumber.String()})
	if err != nil {
		return err
	}
	obj.appCertPublicKeySN = sn

	return nil
}

// 加载应用私钥文件，如果通过SetAppPrivateKey设置了私钥字符串，则不需要再用此方法设置应用私钥
func (obj *Config) LoadAppPrivateKey(filePath string) error {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	blocks := encrypt.ParsePEMBlocks(fileData)
	if blocks == nil {
		return errors.New("支付宝应用私钥无效")
	}

	privateKey, privateKeyType, err := encrypt.ParseRSAPrivateKey(blocks[0].Bytes)
	if err != nil {
		return err
	}

	obj.appPrivateKey = privateKey
	obj.appPrivateKeyType = privateKeyType

	return nil
}

// 设置应用私钥字符串，如果通过LoadAppPrivateKey加载了私钥文件，则不需要再用此方法设置应用私钥
func (obj *Config) SetAppPrivateKey(data string) error {
	dataBytes := encrypt.FormatPKCS1PrivateKey(data)

	blocks := encrypt.ParsePEMBlocks(dataBytes)
	if blocks == nil {
		return errors.New("支付宝应用私钥无效")
	}

	privateKey, privateKeyType, err := encrypt.ParseRSAPrivateKey(blocks[0].Bytes)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	obj.appPrivateKey = privateKey
	obj.appPrivateKeyType = privateKeyType

	return nil
}

// 设置应用ID
func (obj *Config) SetAppID(value string) error {
	if value == "" {
		return errors.New("AppID不能为空")
	}
	obj.appID = value
	return nil
}

// 设置应用签名类型
func (obj *Config) SetAppSignType(value string) error {
	if value != "RSA" && value != "RSA2" {
		return errors.New("签名类型必须是RSA或RSA2(推荐)")
	}
	obj.appSignType = value
	return nil
}

// 获得支付宝根证书SN
func (obj *Config) GetAlipayRootCertSN() string {
	return obj.alipayRootCertSN
}

// 获得应用公钥证书SN
func (obj *Config) GetAppCertPublicKeySN() string {
	return obj.appCertPublicKeySN
}

// 获得应用签名类型
func (obj *Config) GetAppSignType() string {
	return obj.appSignType
}

// 获得支付宝公钥证书
func (obj *Config) GetAlipayPublicKey() *rsa.PublicKey {
	return obj.alipayPublicKey
}

// 获得应用私钥
func (obj *Config) GetAppPrivateKey() *rsa.PrivateKey {
	return obj.appPrivateKey
}

// 获得应用私钥类型(PKCS1/PKCS8)
func (obj *Config) GetAppPrivateKeyType() string {
	return obj.appPrivateKeyType
}

// 获得应用ID
func (obj *Config) GetAppID() string {
	return obj.appID
}

// 解析证书SN
func parseCert(block []byte) (string, error) {
	cert, err := x509.ParseCertificate(block)
	if err != nil && err.Error() == "x509: unsupported elliptic curve" {
		return "", nil
	} else if err != nil {
		return "", err
	}
	if cert.SignatureAlgorithm == x509.SHA256WithRSA || cert.SignatureAlgorithm == x509.SHA1WithRSA {
		sn, parseErr := encrypt.MD5ByStrings([]string{cert.Issuer.String(), cert.SerialNumber.String()})
		if parseErr != nil {
			return "", parseErr
		}
		return sn, nil
	}
	return "", nil
}
