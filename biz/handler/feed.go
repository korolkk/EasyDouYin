package handler

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/jwt"
	"github.com/qiong-14/EasyDouYin/biz/resp"
	"github.com/qiong-14/EasyDouYin/constants"
	"github.com/qiong-14/EasyDouYin/dal"
	"github.com/qiong-14/EasyDouYin/middleware"
	"github.com/qiong-14/EasyDouYin/service"
)

type FeedResponse struct {
	resp.Response
	VideoList []resp.Video `json:"video_list,omitempty"`
	NextTime  int64        `json:"next_time,omitempty"`
}

func getVideoEntities(ctx context.Context, c *app.RequestContext, videoInfos []dal.VideoInfo) []resp.Video {
	videosList := make([]resp.Video, len(videoInfos))
	var wg sync.WaitGroup
	userId, _ := middleware.GetUserIdRedis(jwt.GetToken(ctx, c))

	wg.Add(len(videoInfos))
	for idx, info := range videoInfos {
		go func(resPos int, videoInfo dal.VideoInfo) {
			defer wg.Done()

			// 查询视频用户
			user, _ := service.GetUserInfo(ctx, videoInfo.OwnerId)

			playUrl, coverUrl, _ := middleware.GetUrlOfVideoAndCover(context.Background(), videoInfo.Title, time.Hour)

			// 增加用户喜欢的查询
			favoriteCount, _ := service.GetVideoFavUserCount(ctx, int64(videoInfo.ID))
			//favoriteCount, _ := dal.GetLikeUserCount(ctx, int64(videoInfo.ID))

			FollowCount, _ := dal.FollowCount(ctx, user.Id)
			FollowerCount, _ := dal.FollowerCount(ctx, user.Id)
			IsFollow, _ := dal.IsFollow(ctx, user.Id, userId)
			CommentCount, _ := dal.GetCommentVideoIdCount(ctx, int64(videoInfo.ID))
			videosList[resPos] = resp.Video{
				Id: int64(videoInfo.ID),
				Author: resp.User{
					Id:              user.Id,
					Name:            user.Name,
					Avatar:          user.Avatar,
					BackgroundImage: user.BackgroundImage,
					Signature:       user.Signature,
					FollowCount:     FollowCount,
					FollowerCount:   FollowerCount,
					IsFollow:        IsFollow,
				},
				PlayUrl:       playUrl.String(),
				CoverUrl:      coverUrl.String(),
				FavoriteCount: favoriteCount,
				CommentCount:  CommentCount,
				IsFavorite:    false,
			}
		}(idx, info)

	}
	wg.Wait()

	return videosList
}

func GetVideoStream(ctx context.Context, c *app.RequestContext, lastTime int64, limit int) ([]resp.Video, int64) {
	videoInfos := dal.GetVideoStreamInfo(ctx, lastTime, limit)

	return getVideoEntities(ctx, c, videoInfos), videoInfos[len(videoInfos)-1].CreatedAt.Unix()
}

func Feed(ctx context.Context, c *app.RequestContext) {
	//fmt.Println(c.Query("NextTime"))
	latestTimeStr := c.Query("latest_time")
	latestTime := int(time.Now().Unix())
	if latestTimeStr != "" {
		latestTime, _ = strconv.Atoi(latestTimeStr)
	}

	videoList, nextTime := GetVideoStream(ctx, c, int64(latestTime), constants.FeedVideosCount)
	c.JSON(consts.StatusOK, FeedResponse{
		Response:  resp.Response{StatusCode: 0},
		VideoList: videoList,
		// note: 需要替换成本次视频最小的时间戳
		NextTime: nextTime,
	})
	hlog.CtxTracef(ctx, "status=%d method=%s full_path=%s client_ip=%s host=%s",
		c.Response.StatusCode(),
		c.Request.Header.Method(), c.Request.URI().PathOriginal(), c.ClientIP(), c.Request.Host())
}
