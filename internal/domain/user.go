package domain

import (
	"time"
)

/**
 * @Description
 * @Date 2024/3/1 17:49
 **/

type User struct {
	Id       int64
	Email    string
	Password string

	NickName   string
	Phone      string
	AboutMe    string
	Ctime      time.Time
	Birthday   time.Time
	WechatInfo WechatInfo
}
