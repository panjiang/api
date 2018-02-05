package wallet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/panjiang/golog"
)

// Http method
const (
	MethodGet  = "GET"
	MethodPost = "POST"
)

// API global instance
var API Client

// InitWalletCli 初始化钱包客户端
func InitWalletCli(conf *Config) {
	API.conf = conf
}

// Config 钱包配置参数
type Config struct {
	Host     string `json:"host"`
	Product  string `json:"product"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Client 钱包
type Client struct {
	conf *Config
}

// Request 请求Exchange-API
func (w *Client) Request(method string, url string, v interface{}) (int, map[string]interface{}) {
	var data map[string]interface{}

	reqBody := new(bytes.Buffer)
	if v != nil {
		json.NewEncoder(reqBody).Encode(v)
	}
	log.Debugf("reqBody: %+v", reqBody)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		log.Panic("http", method, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(w.conf.Username, w.conf.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Panic("http", method, err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, data
	}
	log.Debugf("%s %s\n%d %s", method, url, resp.StatusCode, body)

	err = json.Unmarshal(body, &data)
	if err != nil {
		return resp.StatusCode, data
	}

	return resp.StatusCode, data
}

func (w *Client) fullURL(uri string) string {
	u, err := url.Parse(w.conf.Host)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join("/v1/wallet", uri)
	return u.String()
}

// CreateAddress 创建钱包地址
// PATH: /v2/wallet/:product(\\w+)/:uid(\\d+)/address
// return 200: { account(string), address(string), new(bool) }
func (w *Client) CreateAddress(uid uint) (int, map[string]interface{}) {
	return w.Request(MethodGet, w.fullURL(fmt.Sprintf("/%s/%d/address", w.conf.Product, uid)), nil)
}

// GetBalance 获取余额
// PATH: /:address(\\w{34})/balance
// return 200: { balance(float) }
func (w *Client) GetBalance(address string) (int, map[string]interface{}) {
	return w.Request(MethodGet, w.fullURL(fmt.Sprintf("/%s/balance", address)), nil)
}

// SyncBalance 同步余额
// PATH: /:address(\\w{34})/sync_balance
// return 200: { balance(float), recharge(float) }
func (w *Client) SyncBalance(address string) (int, map[string]interface{}) {
	return w.Request(MethodPost, w.fullURL(fmt.Sprintf("/%s/sync_balance", address)), nil)
}

// Pay 支付
// PATH: /:address(\\w{34})/pay/:to_address(\\w{34})/:amount/:fee
// return 200: { balance(float), bill_id(string) }
// return 600: { balance(float) } 余额不足
func (w *Client) Pay(address string, toAddress string, amount float64, fee float64) (int, map[string]interface{}) {
	return w.Request(MethodPost, w.fullURL(fmt.Sprintf("/%s/pay/%s/%f/%f", address, toAddress, amount, fee)), nil)
}

// Fee 支付小费
// PATH: /:address(\\w{34})/fee/:fee
// return 200: { balance(float) }
// return 600: { balance(float), bill_id(string) } 余额不足
func (w *Client) Fee(address string, fee float64) (int, map[string]interface{}) {
	return w.Request(MethodPost, w.fullURL(fmt.Sprintf("/%s/fee//%f", address, fee)), nil)
}

// AdvanceFee 预付小费
// PATH: /:address(\\w{34})/fee/:fee
// return 200: { balance(float) }
// return 600: { balance(float), bill_id(string) } 余额不足
//
// example:
// wallet.API.AdvanceFee("qcAhh3TBa9QQxePudRXBoVe89pVowoab63", 0.1)
func (w *Client) AdvanceFee(address string, fee float64) (int, map[string]interface{}) {
	return w.Request(MethodPost, w.fullURL(fmt.Sprintf("/advance/%s/fee/%f", address, fee)), nil)
}

// ApproveAdvanceFee 批准预付小费（批量）
// PATH: /advance/approve_fee
// body(json): { "bills": [billId... ] }
// return 200: { fee_sum(float:小费总额), affected(int:成功笔数) }
//
// example:
// wallet.API.ApproveAdvanceFee([]string{"52", "53"})
func (w *Client) ApproveAdvanceFee(billIDs []string) (int, map[string]interface{}) {
	return w.Request(MethodPost, w.fullURL("/advance/approve_fee"), map[string]interface{}{
		"bills": billIDs,
	})
}

// CancelAdvanceFee 取消预付小费（单个）
// PATH: /advance/approve_fee
// body(json): { "bill": billId }
// return 200: { fee(float), affected(int:成功笔数) }
//
// example:
// wallet.API.CancelAdvanceFee("54")
func (w *Client) CancelAdvanceFee(billID string) (int, map[string]interface{}) {
	return w.Request(MethodPost, w.fullURL("/advance/cancel_fee"), map[string]interface{}{
		"bill": billID,
	})
}
