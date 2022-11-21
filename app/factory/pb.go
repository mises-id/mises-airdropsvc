package factory

import (
	"github.com/mises-id/mises-airdropsvc/app/models"
	"github.com/mises-id/mises-airdropsvc/lib/utils"
	pb "github.com/mises-id/mises-airdropsvc/proto"
	socialModel "github.com/mises-id/sns-socialsvc/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewAirdrop(in *models.Airdrop) *pb.Airdrop {
	if in == nil {
		return nil
	}
	out := &pb.Airdrop{
		Coin:      float32(utils.UMisesToMises(uint64(in.Coin))),
		Status:    in.Status.String(),
		FinishAt:  uint64(in.FinishAt.Unix()),
		CreatedAt: uint64(in.CreatedAt.Unix()),
	}
	return out
}

func NewUserTwitterAuth(in *models.UserTwitterAuth) *pb.UserTwitterAuth {
	if in == nil {
		return nil
	}
	out := &pb.UserTwitterAuth{
		TwitterUserId: in.TwitterUserId,
		Misesid:       utils.RemoveMisesidProfix(in.Misesid),
		Amount:        float32(utils.UMisesToMises(uint64(in.Amount))),
		CreatedAt:     uint64(in.CreatedAt.Unix()),
	}
	if in.TwitterUser != nil {
		out.Name = in.TwitterUser.Name
		out.Username = in.TwitterUser.UserName
		out.FollowersCount = in.TwitterUser.FollowersCount
		out.TweetCount = in.TwitterUser.TweetCount
		out.TwitterCreatedAt = uint64(in.TwitterUser.CreatedAt.Unix())
	}
	return out
}

func NewChannelUserListSlice(channel_users []*models.ChannelUser) []*pb.ChannelUserInfo {
	result := make([]*pb.ChannelUserInfo, len(channel_users))
	for i, channel_user := range channel_users {
		result[i] = NewChannelUser(channel_user)
	}
	return result
}

func NewUserInfo(user *socialModel.User) *pb.UserInfo {
	if user == nil {
		return nil
	}
	userinfo := pb.UserInfo{
		Uid:             user.UID,
		Username:        user.Username,
		Misesid:         user.Misesid,
		Gender:          user.Gender.String(),
		Mobile:          user.Mobile,
		Email:           user.Email,
		Address:         user.Address,
		Avatar:          user.AvatarUrl,
		Intro:           user.Intro,
		IsFollowed:      user.IsFollowed,
		IsAirdropped:    user.IsAirdropped,
		AirdropStatus:   user.AirdropStatus,
		IsBlocked:       user.IsBlocked,
		FollowingsCount: user.FollowingCount,
		FansCount:       user.FansCount,
		LikedCount:      user.LikedCount,
		NewFansCount:    user.NewFansCount,
		IsLogined:       user.IsLogined,
		HelpMisesid:     user.Misesid,
	}
	if user.NftAvatar != nil {
		userinfo.AvatarUrl = &pb.UserAvatar{
			Small:      user.NftAvatar.ImageThumbnailUrl,
			Medium:     user.NftAvatar.ImagePreviewUrl,
			Large:      user.NftAvatar.ImageURL,
			NftAssetId: user.NftAvatar.NftAssetID.Hex(),
		}
	} else {
		userinfo.AvatarUrl = &pb.UserAvatar{
			Small:      user.AvatarUrl,
			Medium:     user.AvatarUrl,
			Large:      user.AvatarUrl,
			NftAssetId: "",
		}
		if user.Avatar != nil {
			userinfo.AvatarUrl.Small = user.Avatar.Small
			userinfo.AvatarUrl.Medium = user.Avatar.Medium
		}
	}
	return &userinfo
}

func NewChannelUser(channel_user *models.ChannelUser) *pb.ChannelUserInfo {

	return &pb.ChannelUserInfo{
		Id:             channel_user.ID.Hex(),
		ChannelId:      channel_user.ChannelID.Hex(),
		ValidState:     int32(channel_user.ValidState),
		Amount:         uint64(channel_user.Amount),
		TxId:           channel_user.TxID,
		User:           NewUserInfo(channel_user.User),
		AirdropState:   int32(channel_user.AirdropState),
		AirdropTime:    uint64(channel_user.AirdropTime.Unix()),
		CreatedAt:      uint64(channel_user.CreatedAt.Unix()),
		ChannelUid:     channel_user.ChannelUID,
		ChannelMisesid: channel_user.ChannelMisesid,
	}

}

func docID(id primitive.ObjectID) string {
	if id.IsZero() {
		return ""
	}
	return id.Hex()
}
