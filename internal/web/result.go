package web

/**
 * @Description
 * @Date 2024/3/11 21:49
 **/

type StatusMsg struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
