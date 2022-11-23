package handlers

import (
	"context"

	"github.com/mises-id/mises-airdropsvc/app/factory"
	"github.com/mises-id/mises-airdropsvc/app/models"
	airdropSVC "github.com/mises-id/mises-airdropsvc/app/services/airdrop"
	channelSVC "github.com/mises-id/mises-airdropsvc/app/services/channel_list"
	channelUserSVC "github.com/mises-id/mises-airdropsvc/app/services/channel_user"
	"github.com/mises-id/mises-airdropsvc/app/services/user_twitter"
	"github.com/mises-id/mises-airdropsvc/lib/pagination"
	pb "github.com/mises-id/mises-airdropsvc/proto"
)

// NewService returns a naïve, stateless implementation of Service.
func NewService() pb.AirdropsvcServer {
	return airdropsvcService{}
}

type airdropsvcService struct{}

func (s airdropsvcService) Test(ctx context.Context, in *pb.TestRequest) (*pb.TestResponse, error) {
	var resp pb.TestResponse
	return &resp, nil
}

func (s airdropsvcService) GetTwitterAuthUrl(ctx context.Context, in *pb.GetTwitterAuthUrlRequest) (*pb.GetTwitterAuthUrlResponse, error) {
	var resp pb.GetTwitterAuthUrlResponse
	url, err := user_twitter.GetTwitterAuthUrl(ctx, in.CurrentUid)
	if err != nil {
		return nil, err
	}
	resp.Code = 0
	resp.Url = url
	return &resp, nil
}

func (s airdropsvcService) GetAirdropInfo(ctx context.Context, in *pb.GetAirdropInfoRequest) (*pb.GetAirdropInfoResponse, error) {
	var resp pb.GetAirdropInfoResponse
	out, err := user_twitter.GetAirdropInfo(ctx, in.CurrentUid)
	if err != nil {
		return nil, err
	}
	resp.Airdrop = factory.NewAirdrop(out.Airdrop)
	resp.Twitter = factory.NewUserTwitterAuth(out.Twitter)
	resp.Code = 0
	return &resp, nil
}

func (s airdropsvcService) ChannelInfo(ctx context.Context, in *pb.ChannelInfoRequest) (*pb.ChannelInfoResponse, error) {
	var resp pb.ChannelInfoResponse
	out, err := channelSVC.ChannelInfo(ctx, &channelSVC.ChannelUrlInput{Misesid: in.Misesid, Type: in.Type, Medium: in.Medium})
	if err != nil {
		return nil, err
	}
	resp.Code = 0
	resp.Url = out.Url
	resp.IosLink = out.IosLink
	resp.IosMediumLink = out.IosMediumLink
	resp.MediumUrl = out.MediumUrl
	resp.AirdropAmount = float32(out.AirdropAmount)
	resp.TotalChannelUser = out.TotalChannelUser
	return &resp, nil
}

func (s airdropsvcService) PageChannelUser(ctx context.Context, in *pb.PageChannelUserRequest) (*pb.PageChannelUserResponse, error) {
	var resp pb.PageChannelUserResponse
	PageParams := &pagination.TraditionalParams{}
	if in.Paginator != nil {
		PageParams = &pagination.TraditionalParams{
			PageNum:  int64(in.Paginator.PageNum),
			PageSize: int64(in.Paginator.PageSize),
		}
	}
	channel_users, page, err := channelUserSVC.PageChannelUser(ctx, &channelUserSVC.PageChannelUserInput{
		Misesid:    in.Misesid,
		PageParams: PageParams,
	})

	if err != nil {
		return nil, err
	}
	resp.Code = 0
	resp.ChannelUsers = factory.NewChannelUserListSlice(channel_users)
	tradpage := page.BuildJSONResult().(*pagination.TraditionalPagination)
	resp.Paginator = &pb.Page{
		PageNum:      uint64(tradpage.PageNum),
		PageSize:     uint64(tradpage.PageSize),
		TotalPage:    uint64(tradpage.TotalPages),
		TotalRecords: uint64(tradpage.TotalRecords),
	}
	return &resp, nil
}

func (s airdropsvcService) AirdropTwitter(ctx context.Context, in *pb.AirdropTwitterRequest) (*pb.AirdropTwitterResponse, error) {
	var resp pb.AirdropTwitterResponse
	airdropSVC.AirdropTwitter(ctx)
	return &resp, nil
}

func (s airdropsvcService) AirdropChannel(ctx context.Context, in *pb.AirdropChannelRequest) (*pb.AirdropChannelResponse, error) {
	var resp pb.AirdropChannelResponse
	channelUserSVC.AirdropChannel(ctx)
	return &resp, nil
}

func (s airdropsvcService) TwitterCallback(ctx context.Context, in *pb.TwitterCallbackRequest) (*pb.TwitterCallbackResponse, error) {
	var resp pb.TwitterCallbackResponse
	params := &user_twitter.CallbackParams{
		OauthToken:    in.OauthToken,
		OauthVerifier: in.OauthVerifier,
	}
	if in.UserAgent != nil {
		user_agent := &models.UserAgent{
			Ua:       in.UserAgent.Ua,
			Ipaddr:   in.UserAgent.Ipaddr,
			Os:       in.UserAgent.Os,
			Platform: in.UserAgent.Platform,
			Browser:  in.UserAgent.Browser,
			DeviceId: in.UserAgent.DeviceId,
		}
		params.UserAgent = user_agent
	}
	url := user_twitter.TwitterCallback(ctx, in.CurrentUid, params)
	resp.Code = 0
	resp.Url = url
	return &resp, nil
}

func (s airdropsvcService) CheckTwitterUser(ctx context.Context, in *pb.CheckTwitterUserRequest) (*pb.CheckTwitterUserResponse, error) {
	var resp pb.CheckTwitterUserResponse
	err := user_twitter.PlanCheckTwitterUser(ctx)
	if err != nil {
		return nil, err
	}
	resp.Code = 0
	return &resp, nil
}

func (s airdropsvcService) GetChannelUser(ctx context.Context, in *pb.GetChannelUserRequest) (*pb.GetChannelUserResponse, error) {
	var resp pb.GetChannelUserResponse
	out, err := channelUserSVC.GetChannelUser(ctx, &channelUserSVC.GetCHannelUserInput{Misesid: in.Misesid})
	if err != nil {
		return nil, err
	}
	resp.ChanelUser = factory.NewChannelUser(out)
	return &resp, nil
}

func (s airdropsvcService) TwitterFollow(ctx context.Context, in *pb.TwitterFollowRequest) (*pb.TwitterFollowResponse, error) {
	var resp pb.TwitterFollowResponse
	err := user_twitter.FollowTwitter(ctx)
	if err != nil {
		return nil, err
	}
	resp.Code = 0
	return &resp, nil
}

func (s airdropsvcService) LookupTwitter(ctx context.Context, in *pb.LookupTwitterRequest) (*pb.LookupTwitterResponse, error) {
	var resp pb.LookupTwitterResponse
	err := user_twitter.PlanLookupTwitterUser(ctx)
	if err != nil {
		return nil, err
	}
	resp.Code = 0
	return &resp, nil
}

func (s airdropsvcService) SendTweet(ctx context.Context, in *pb.SendTweetRequest) (*pb.SendTweetResponse, error) {
	var resp pb.SendTweetResponse
	err := user_twitter.PlanSendTweet(ctx)
	if err != nil {
		return nil, err
	}
	resp.Code = 0
	return &resp, nil
}
