package send_many

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"project/internal/service/sms/huawei/core"
)

/**
 * @Description
 * @Date 2024/8/13 17:40
 **/

func BuildRequestBody(sender string, item []map[string]interface{}, statusCallBack string) []byte {
	body := make(map[string]interface{})
	body["smsContent"] = item
	body["from"] = sender
	if statusCallBack != "" {
		body["statusCallback"] = statusCallBack
	}
	res, _ := json.Marshal(body)
	return res
}

func InitDiffSms(reveiver []string, templateId string, templateParas []string, signature string) map[string]interface{} {
	diffSms := make(map[string]interface{})
	diffSms["to"] = reveiver
	diffSms["templateId"] = templateId
	if templateParas != nil && len(templateParas) > 0 {
		diffSms["templateParas"] = templateParas
	}
	if signature != "" {
		diffSms["signature"] = signature
	}
	return diffSms
}

func Post(url string, param []byte, appInfo core.Signer) (string, error) {
	if param == nil || appInfo == (core.Signer{}) {
		return "", nil
	}

	// 代码样例为了简便，设置了不进行证书校验，请在商用环境自行开启证书校验。
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(param))
	if err != nil {
		return "", err
	}

	// 对请求增加内容格式，固定头域
	req.Header.Add("Content-Type", "application/json")
	// 对请求进行HMAC算法签名，并将签名结果设置到Authorization头域。
	appInfo.Sign(req)

	fmt.Println(req.Header)
	// 发送短信请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	// 获取短信响应
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
