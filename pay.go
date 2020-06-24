package pay

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"pinpei_go/config"
	"pinpei_go/db"
	"pinpei_go/model"
	"pinpei_go/paymod"
	"pinpei_go/paymod/wechat"
	"pinpei_go/tools"
	"pinpei_go/utils"
	"sort"
	"strconv"
	"strings"
)

type ReqMyAddress struct {
	ID 			int `json:"id" form:"column:id"`
	Uid 		string `json:"uid" form:"column:uid"`
	Name		string `json:"name" form:"column:name"`
	Province 	string `json:"province" form:"column:province"`
	City 		string `json:"city" form:"column:city"`
	County 		string `json:"county" form:"column:county"`
	Address 	string `json:"address" form:"column:address"`
	Tel 		string `json:"tel" form:"column:tel"`
	Status 		int `json:"status" form:"column:status"`
	Createtime 	string `json:"createtime" form:"column:createtime"`
	Updatetime 	string `json:"updatetime" form:"column:updatetime"`
	AreaCode 	int `json:"area_code" form:"column:area_code"`
	Type 		int `json:"type" form:"column:type"`
}

type ReqOrder struct {
	GoodsId 	int `json:"goods_id" form:"column:goods_id"`
	Num 		int `json:"num" form:"column:num"`
	SkuId		int `json:"sku_id" form:"column:sku_id"`
	Message 	string `json:"message" form:"column:message"`
	SkuName 	string `json:"sku_name" form:"column:sku_name"`
	Title 		string `json:"title" form:"column:title"`
	ImgUrl 		string `json:"imgUrl" form:"column:imgUrl"`
	Price 		float64 `json:"price" form:"column:price"`
	CategoryId 		int `json:"category_id" form:"column:category_id"`
	CategoryName 	string `json:"category_name" form:"column:category_name"`
}

type ReqPayment struct {
	OrderSn 	int `json:"order_sn" form:"column:order_sn"`
	Fee 		int `json:"fee" form:"column:fee"`
	Type		int `json:"type" form:"column:type"`
}
var dbIns = db.GetDb("pinpei_go")

type ReqCreateOrder struct {
	Address   ReqMyAddress
	Paytype   int `json:"paytype" form:"paytype"`
	OrderInfo ReqOrder
	TypeText  string `json:"typetext" form:"typetext"`
	Uid       string `json:"uid" form:"uid"`
	Token     string `json:"token" form:"token"`
	Timestamp string `json:"timestamp" form:"timestamp"`
	Sign      string	`json:"sign" form:"sign"`
}
const (
	PayCom_Wx		= "1" //wx
	PayCom_Ali		= "2" //ali
)

func WxMakeOrder(c *gin.Context) {
	var reqGetOrder ReqCreateOrder
	if err := utils.MiddleValidator(c, &reqGetOrder); err != nil {
		tools.FailWithMsg(c, "参数错误")
		return
	}
	var neworder model.Orders
	neworder.AddressId = reqGetOrder.Address.ID
	neworder.Uid = reqGetOrder.Uid
	neworder.ReceiverName = reqGetOrder.Address.Name
	neworder.ReceiverMobile = reqGetOrder.Address.Tel
	neworder.ReceiverAddress = reqGetOrder.Address.Address
	neworder.KdNum = reqGetOrder.Uid
	neworder.Province = reqGetOrder.Uid
	neworder.City = reqGetOrder.Uid
	neworder.Qu = reqGetOrder.Uid
	neworder.Paytype = reqGetOrder.Paytype
	neworder.Status = 0
	neworder.Status2 = 0
	neworder.SumNum = int(reqGetOrder.OrderInfo.Num)
	neworder.SumPrice = decimal.NewFromFloat(reqGetOrder.OrderInfo.Price).Mul(decimal.NewFromInt(int64(reqGetOrder.OrderInfo.Num)))
	//sumPrice := neworder.SumPrice
	wxSet := config.Conf.Common.WxPay
	OrderSn := "Upk2mizW1i5QzR4ylsQhJc9BnpInhoVC"
	//fmt.Println("out_trade_no:", OrderSn)
	neworder.OrderSn = OrderSn
	// 初始化参数Map
	bm := make(paymod.BodyMap)
	nonceStr := "Upk2mizW1i5QzR4ylsQhJc9BnpInhoVC"
	bm.Set("appid", wxSet.AppId)
	bm.Set("mch_id", wxSet.MchId)
	bm.Set("nonce_str", nonceStr)
	//bm.Set("nonce_str", "ASDFAJK2134")
	bm.Set("body", "app支付")
	bm.Set("out_trade_no", OrderSn)
	bm.Set("total_fee", 1)
	//bm.Set("spbill_create_ip", getIpAddress(c))
	bm.Set("spbill_create_ip", "110.87.23.60")
	bm.Set("notify_url", wxSet.NotifyUrl)
	bm.Set("trade_type", wechat.TradeType_App)
	//bm.Set("device_info", "WEB")
	bm.Set("sign_type", wechat.SignType_MD5)

	sceneInfo := make(map[string]map[string]string)
	h5Info := make(map[string]string)
	h5Info["type"] = "Wap"
	h5Info["wap_url"] = "https://app.pinpei.vip"
	h5Info["wap_name"] = "拼配商城"
	sceneInfo["h5_info"] = h5Info
	//bm.Set("scene_info", sceneInfo)

	//商户流水号  商户号+时间+流水号
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//tradeNo := yourMchID + time.Now().Format(formatDate) + strconv.FormatInt(time.Now().Unix(), 10)[4:] + strconv.Itoa(r.Intn(8999)+1000)
	//签名算法
	//sign := WXSign(wxSet.MchKey, wechat.SignType_MD5, bm)
	//bm.Set("sign", sign)
	//fmt.Println("req xml ",bm)

	client := wechat.NewClient(wxSet.AppId, wxSet.MchId, wxSet.MchKey, true)

	// 设置国家，不设置默认就是 China
	client.SetCountry(wechat.China)
	// 请求支付下单，成功后得到结果
	wxRsp, err := client.UnifiedOrder(bm)
	if err != nil {
		fmt.Println("Error:", err)
		tools.FailWithMsg(c, "创建订单失败")
		return
	}
	if wxRsp.ReturnCode=="FAIL" {
		fmt.Println("request wx order Body:", bm)
		fmt.Println("return:", wxRsp)
		tools.FailWithMsg(c, "创建订单失败")
		return
	}

	//timeStamp := strconv.FormatInt(time.Now().Unix(), 10)

	// 获取H5支付需要的paySign
	//pac := "prepay_id=" + wxRsp.PrepayId
	//paySign := wechat.GetAppPaySign(wxSet.AppId, wxSet.MchId, wxRsp.NonceStr, pac, wechat.SignType_MD5, timeStamp, wxSet.AppSecret)
	//fmt.Println("paySign:", paySign)

	//back := dbIns.Create(neworder)
	//todo 是否需要校验
	d := make(map[string]interface{} )
	//d["paySign"] = paySign
	d["prepay_id"] = wxRsp.PrepayId
	d["mweb_url "] = wxRsp.MwebUrl
	data := tools.ResponseObj{
		Data:  d,
	}
	//if back.Error == nil {
		tools.SuccessWithMsg(c, "创建订单成功", data)
	//} else {
	//	tools.FailWithMsg(c, "创建订单失败")
	//}
}

/*
*	getIpAddress 获取IP地址
*	param	r	 *http.Request
*	reply	IP地址
 */
func getIpAddress(r *gin.Context) string {
	realip := r.Request.Header.Get("X-Real-Ip")
	if realip != "" {
		return realip
	}
	remoteAddr := r.Request.RemoteAddr
	if remoteAddr != "" {
		idx := strings.Index(remoteAddr, ":")
		return remoteAddr[:idx]
	}
	ips := r.Request.Header.Get("X-Forwarded-For")
	//glog.Info("X-Forwarded-For", ips)
	if ips != "" {
		iplist := strings.Split(ips, ",")
		return strings.TrimSpace(iplist[0])
	}
	return ""
}

func WxPayment(c *gin.Context) {
	var reqPayment ReqPayment
	if err := utils.MiddleValidator(c, &reqPayment); err != nil {
		tools.FailWithMsg(c, "参数错误")
		return
	}
	tools.SuccessWithMsg(c, "创建订单成功", nil)
}

//本地通过支付参数计算Sign值
func WXSign(apiKey string, signType string, body paymod.BodyMap) (sign string) {
	signStr := sortWeChatSignParams(apiKey, body)
	//fmt.Println("signStr:", signStr)
	var hashSign []byte
	if signType == wechat.SignType_HMAC_SHA256 {
		hash := hmac.New(sha256.New, []byte(apiKey))
		hash.Write([]byte(signStr))
		hashSign = hash.Sum(nil)
	} else {
		hash := md5.New()
		hash.Write([]byte(signStr))
		hashSign = hash.Sum(nil)
	}
	sign = strings.ToUpper(hex.EncodeToString(hashSign))
	return
}

//获取根据Key排序后的请求参数字符串
func sortWeChatSignParams(apiKey string, body paymod.BodyMap) string {
	keyList := make([]string, 0)
	for k := range body {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)
	buffer := new(bytes.Buffer)
	for _, k := range keyList {
		buffer.WriteString(k)
		buffer.WriteString("=")

		valueStr := convert2String(body[k])
		buffer.WriteString(valueStr)

		buffer.WriteString("&")
	}
	buffer.WriteString("key=")
	buffer.WriteString(apiKey)
	return buffer.String()
}

func convert2String(value interface{}) (valueStr string) {
	switch v := value.(type) {
	case int:
		valueStr = Int2String(v)
	case int64:
		valueStr = Int642String(v)
	case float64:
		valueStr = Float64ToString(v)
	case float32:
		valueStr = Float32ToString(v)
	case string:
		valueStr = v
	default:
		valueStr = ""
	}
	return
}

//Int转字符串
func Int2String(intNum int) (intStr string) {
	intStr = strconv.Itoa(intNum)
	return
}

//Int64转字符串
func Int642String(intNum int64) (int64Str string) {
	//10, 代表10进制
	int64Str = strconv.FormatInt(intNum, 10)
	return
}

//Float64转字符串
//    floatNum：float64数字
//    prec：精度位数（不传则默认float数字精度）
func Float64ToString(floatNum float64, prec ...int) (floatStr string) {
	if len(prec) > 0 {
		floatStr = strconv.FormatFloat(floatNum, 'f', prec[0], 64)
		return
	}
	floatStr = strconv.FormatFloat(floatNum, 'f', -1, 64)
	return
}

//Float32转字符串
//    floatNum：float32数字
//    prec：精度位数（不传则默认float数字精度）
func Float32ToString(floatNum float32, prec ...int) (floatStr string) {
	if len(prec) > 0 {
		floatStr = strconv.FormatFloat(float64(floatNum), 'f', prec[0], 32)
		return
	}
	floatStr = strconv.FormatFloat(float64(floatNum), 'f', -1, 32)
	return
}
