package status

import (
	"fmt"
	"net/url"
	"strings"
)

/**
 * @Description
 * @Date 2024/8/13 18:56
 **/

// 短信平台上报状态报告数据样例(urlencode)
//success_body := "sequence=1&total=1&updateTime=2018-10-31T08%3A43%3A41Z&source=2&smsMsgId=2ea20735-f856-4376-afbf-570bd70a46ee_11840135&status=DELIVRD";
//failed_body := "sequence=1&total=1&updateTime=2018-10-31T08%3A43%3A41Z&source=2&smsMsgId=2ea20735-f856-4376-afbf-570bd70a46ee_11840135&status=E200027"

func OnSmsStatusReport(data string) {
	ss, _ := url.QueryUnescape(data)
	params := strings.Split(ss, "&")
	keyValues := make(map[string]string)
	for i := range params {
		temp := strings.Split(params[i], "=")
		keyValues[temp[0]] = temp[1]
	}
	status := keyValues["status"]
	if status == "DELIVRD" {
		fmt.Println("Send sms success. smsMsgId: " + keyValues["smsMsgId"])
	} else {
		fmt.Println("Send sms failed. smsMsgId: " + keyValues["smsMsgId"])
		fmt.Println("Failed status:  " + keyValues["status"])
	}
}
