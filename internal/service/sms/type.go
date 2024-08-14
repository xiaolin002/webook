package sms

import "context"

/**
 * @Description
 * @Date 2024/3/10 22:25
 **/

//发送短信的抽象
// 屏蔽不同供应商的区别

type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}

//	其中的appid  和signature是固定的，一个生产线这个东西是固定不动的，所以可以设置到环境变量中
// 这里接口可以做两个的 一个是单一发送一个是批量发送
// 或者可以实现多模版一块发送 (具体可根据不同的供应商来实现)
