package ginx

/**
 * @Description
 * @Date 2024/3/11 21:49
 **/

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
