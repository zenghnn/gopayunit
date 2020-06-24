package qq

import (
	"fmt"
	"os"
	"testing"

	"pinpei_go/paymod"
)

var (
	client *Client
	mchId  = "1368139502"
	apiKey = "GFDS8j98rewnmgl45wHTt980jg543abc"
)

func TestMain(m *testing.M) {

	// 初始化QQ客户端
	//    mchId：商户ID
	//    apiKey：API秘钥值
	client = NewClient(mchId, apiKey)

	//err := client.AddCertFilePath(nil, nil, nil)
	//if err != nil {
	//	panic(err)
	//}

	os.Exit(m.Run())
}

func TestClient_MicroPay(t *testing.T) {
	bm := make(paymod.BodyMap)
	bm.Set("nonce_str", paymod.GetRandomString(32))

	qqRsp, err := client.MicroPay(bm)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println("qqRsp:", *qqRsp)
}

func TestNotifyResponse_ToXmlString(t *testing.T) {
	n := new(NotifyResponse)
	n.ReturnCode = "SUCCESS"
	fmt.Println(n.ToXmlString())

	n.ReturnCode = "FAIL"
	n.ReturnMsg = "abc"
	fmt.Println(n.ToXmlString())

}
