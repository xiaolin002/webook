package ioc

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSMS "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	sms2 "project/internal/service/sms"
	"project/internal/service/sms/huawei"
	"project/internal/service/sms/huawei/core"
	"project/internal/service/sms/tencent"
)

/**
 * @Description
 * @Date 2024/3/12 19:08
 **/

func InitSmsService() sms2.Service {
	return nil

}
func initTencentSMSService() sms2.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("找不到腾讯的secret_id")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("找不到腾讯的secret_key")
	}
	c, err := tencentSMS.NewClient(common.NewCredential(secretId, secretKey), "api-nanjing", profile.NewClientProfile())
	if err != nil {
		panic(err)
	}
	return tencent.NewSmsService(c, "14088888", "测试科技")

}

func initHUAWEISMSService() sms2.Service {
	appKey, ok := os.LookupEnv("SMS_APP_Key")
	if !ok {
		panic("找不到华为的appKey")
	}
	appSecret, ok := os.LookupEnv("SMS_APP_SECRET")
	if !ok {
		panic("找不到华为的appSecret")
	}
	appInfo := core.Signer{
		// 认证用的appKey和appSecret硬编码到代码中或者明文存储都有很大的安全风险，建议在配置文件或者环境变量中密文存放，使用时解密，确保安全；
		Key:    appKey,    //App Key
		Secret: appSecret, //App Secret
	}

	return huawei.NewHuaweiSmsService(&appInfo, "8ff55eac1d0b478ab3c06c3c6a492300", "华为云短信测试")
}
