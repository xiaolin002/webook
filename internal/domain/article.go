package domain

import "time"

/**
 * @Description
 * @Date 2025/4/3 16:13
 **/
type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Ctime   time.Time
	Utime   time.Time
}
type Author struct {
	Id   int64
	Name string
}
