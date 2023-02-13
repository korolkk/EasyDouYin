package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/qiong-14/EasyDouYin/biz/resp"
	"github.com/qiong-14/EasyDouYin/dal"
	"github.com/qiong-14/EasyDouYin/mw"
	"github.com/qiong-14/EasyDouYin/tools"
	"net/http"
	"strconv"
)

// usersLoginInfo use map to store user info, and key is username+password for demo
// user data will be cleared every time the server starts
// test data: username=zhanglei, password=douyin
//var usersLoginInfo = map[string]resp.User{
//	"zhangleidouyin": {
//		Id:            1,
//		Name:          "zhanglei",
//		FollowCount:   10,
//		FollowerCount: 5,
//		IsFollow:      true,
//	},
//}

//var userIdSequence = int64(1)

func Register(ctx context.Context, c *app.RequestContext) {
	username := c.Query("username")
	password := c.Query("password")

	// 查找用户名是否已经注册
	u, err := dal.GetUserByName(ctx, username)
	if u.Name == username {
		c.JSON(consts.StatusOK, resp.UserLoginResponse{
			Response: resp.Response{StatusCode: 1, StatusMsg: "user already exits"},
		})
		return
	}
	// 加密储存
	if err = dal.CreateUser(ctx, &dal.User{Name: username, Password: tools.Encoder(password)}); err != nil {
		c.JSON(consts.StatusOK, resp.UserLoginResponse{
			Response: resp.Response{StatusCode: 1, StatusMsg: "user create failed"},
		})
		return
	}
	mw.JwtMiddleware.LoginHandler(ctx, c)

	hlog.CtxTracef(ctx, "status=%d method=%s full_path=%s client_ip=%s host=%s",
		c.Response.StatusCode(),
		c.Request.Header.Method(), c.Request.URI().PathOriginal(), c.ClientIP(), c.Request.Host())
}

func UserInfo(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if user, err := dal.GetUserById(ctx, id); err == nil {
		c.JSON(http.StatusOK, resp.UserResponse{
			Response: resp.Response{StatusCode: 0},
			User: resp.User{
				Id:   user.Id,
				Name: user.Name,
			},
		})
	} else {
		c.JSON(http.StatusOK, resp.UserResponse{
			Response: resp.Response{StatusCode: 1, StatusMsg: "user doesn't exist"},
		})
	}
	hlog.CtxTracef(ctx, "status=%d method=%s full_path=%s client_ip=%s host=%s",
		c.Response.StatusCode(),
		c.Request.Header.Method(), c.Request.URI().PathOriginal(), c.ClientIP(), c.Request.Host())
}
