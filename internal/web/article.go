package web

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"net/http"
	"project/internal/domain"
	"project/internal/service"
	"project/internal/web/jwt"
	"strconv"
	"time"
)

/**
 * @Description
 * @Date 2025/4/3 16:11
 **/
type ArticleHandler struct {
	svc service.ArticleService
	// 聚合互动服务  点赞 收藏 评论
	intrSvc service.InteractiveService
	biz     string
}

func NewArticleHandler(svc service.ArticleService, intrSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		intrSvc: intrSvc,
		biz:     "article",
	}
}

func (h *ArticleHandler) RegisterRoute(server *gin.Engine) {
	g := server.Group("/articles")

	//g.PUT("/", h.Edit)
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)

	// 创作者接口  查看指定文章内容
	g.GET("/detail/:id", h.Detail)
	// 按照道理来说，这边就是 GET 方法
	// /list?offset=?&limit=?
	// 查看文章列表
	g.POST("/list", h.List)

	//读者接口
	pub := g.Group("/pub")
	pub.GET("/detail/:id", h.PubDetail)
	pub.POST("/like", h.Like)
	pub.POST("/collect", h.Collect)

}

// Edit 接收 Article 输入，返回一个 ID，文章的 ID
func (h *ArticleHandler) Edit(ctx *gin.Context) {

	// 注意新建和编辑的区别
	// 新建的话，ID就不传，编辑的话就需要传id
	type Req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, StatusMsg{
			Msg: "系统错误",
		})

		return
	}
	ctx.JSON(http.StatusOK, StatusMsg{
		Data: id,
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {

	// 发表的流程一定是先保存到制作库后保存到读者库
	type Req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, StatusMsg{
			Msg: "系统错误",
		})
		// 记录日志 发表文章失败

		return
	}
	ctx.JSON(http.StatusOK, StatusMsg{
		Data: id,
	})
}
func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	err := h.svc.Withdraw(ctx, uc.Uid, req.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, StatusMsg{
			Msg:  "系统错误",
			Code: 5,
		})

		return
	}
	ctx.JSON(http.StatusOK, StatusMsg{
		Msg: "OK",
	})
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, StatusMsg{
			Msg:  "id 参数错误",
			Code: 4,
		})
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, StatusMsg{
			Msg:  "系统错误",
			Code: 5,
		})

		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	if art.Author.Id != uc.Uid {
		// 有人在搞鬼
		ctx.JSON(http.StatusOK, StatusMsg{
			Msg:  "系统错误",
			Code: 5,
		})
		// 日志记录
		return
	}

	vo := ArticleVo{
		Id:    art.Id,
		Title: art.Title,
		//Abstract: art.Abstract(),

		Content:  art.Content,
		AuthorId: art.Author.Id,
		// 列表，你不需要
		Status: art.Status.ToUint8(),
		Ctime:  art.Ctime.Format(time.DateTime),
		Utime:  art.Utime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, StatusMsg{Data: vo})
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	// 我要不要检测一下？
	uc := ctx.MustGet("user").(jwt.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, StatusMsg{
			Code: 5,
			Msg:  "系统错误",
		})

		return
	}
	ctx.JSON(http.StatusOK, StatusMsg{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),

				//Content:  src.Content,
				AuthorId: src.Author.Id,
				// 列表，你不需要
				Status: src.Status.ToUint8(),
				Ctime:  src.Ctime.Format(time.DateTime),
				Utime:  src.Utime.Format(time.DateTime),
			}
		}),
	})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, StatusMsg{
			Msg:  "id 参数错误",
			Code: 4,
		})
		return
	}
	art, err := h.svc.GetPubById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, StatusMsg{
			Msg:  "系统错误",
			Code: 5,
		})

		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	if art.Author.Id != uc.Uid {
		// 有人在搞鬼
		ctx.JSON(http.StatusOK, StatusMsg{
			Msg:  "系统错误",
			Code: 5,
		})
		// 日志记录
		return
	}

	go func() {
		newCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := h.intrSvc.IncrReadCnt(newCtx, h.biz, art.Id)
		if er != nil {
			// 记录日志
		}
	}()

	ctx.JSON(http.StatusOK, StatusMsg{
		Data: ArticleVo{
			Id:    art.Id,
			Title: art.Title,

			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			// 列表，你不需要
			Status: art.Status.ToUint8(),
			Ctime:  art.Ctime.Format(time.DateTime),
			Utime:  art.Utime.Format(time.DateTime),
		},
	})
}

func (h *ArticleHandler) Like(c *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
		// true 是点赞，false 是不点赞
		Like bool `json:"like"`
	}
	var req Req
	if err := c.Bind(&req); err != nil {
		return
	}
	uc := c.MustGet("user").(jwt.UserClaims)
	var err error
	if req.Like {
		// 点赞
		err = h.intrSvc.Like(c, h.biz, req.Id, uc.Uid)
	} else {
		// 取消点赞
		err = h.intrSvc.CancelLike(c, h.biz, req.Id, uc.Uid)
	}
	if err != nil {
		c.JSON(http.StatusOK, StatusMsg{
			Code: 5, Msg: "系统错误",
		})
		return
	}
	c.JSON(http.StatusOK, StatusMsg{
		Msg: "OK",
	})
}

func (h *ArticleHandler) Collect(c *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
		// true 是收藏，false 是不收藏
		// 暂时未做 取消收藏功能 但与取消喜欢一样
		// 	Coll bool `json:"collect"`
		Cid int64 `json:"cid"`
	}
	var req Req
	if err := c.Bind(&req); err != nil {
		return
	}
	uc := c.MustGet("user").(jwt.UserClaims)
	err := h.intrSvc.Collect(c, h.biz, req.Id, req.Cid, uc.Uid)
	if err != nil {
		c.JSON(http.StatusOK, StatusMsg{
			Code: 5, Msg: "系统错误",
		})
		return
	}
	c.JSON(http.StatusOK, StatusMsg{
		Msg: "OK",
	})

}
