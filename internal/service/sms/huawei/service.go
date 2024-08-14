package huawei

import (
	"context"
	"project/internal/service/sms/huawei/core"
	sm "project/internal/service/sms/huawei/send_many"
	"project/internal/service/sms/huawei/status"
)

/**
 * @Description
 * @Date 2024/8/13 16:53
 **/

var (
	apiAddress = "https://smsapi.cn-north-4.myhuaweicloud.com:443/sms/batchSendSms/v1"
	sender     = "csms12345678"
)

// 这里采用的事华为云的特殊Ak/Sk 认证
// 其中这里采用的是发送分批短信（单模版）如果用多模版的话需要改造接口

type SmsService struct {
	appInfo *core.Signer
	// 模版id
	templateId string
	// 签名名称
	signature string
	// 选填 ,短信状态报告接收地址
	statusCallBack string
}

func NewHuaweiSmsService(appInfo *core.Signer, templateId string, signature string) *SmsService {
	return &SmsService{
		appInfo:        appInfo,
		templateId:     templateId,
		signature:      signature,
		statusCallBack: "",
	}
}

func (s *SmsService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	item1 := sm.InitDiffSms(numbers, tplId, args, s.signature)
	item := []map[string]interface{}{item1}
	body := sm.BuildRequestBody(sender, item, s.statusCallBack)
	resp, err := sm.Post(apiAddress, []byte(body), *s.appInfo)
	if err != nil {
		return err
	}
	status.OnSmsStatusReport(resp)
	return nil
}
