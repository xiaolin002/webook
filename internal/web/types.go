package web

import "github.com/gin-gonic/gin"

/**
 * @Description
 * @Date 2025/4/1 20:48
 **/

type Handler interface {
	RegisterRouter(server *gin.Engine)
}
