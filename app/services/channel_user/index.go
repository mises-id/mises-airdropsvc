package channel_user

import (
	"context"

	"github.com/mises-id/mises-airdropsvc/app/models"
	"github.com/mises-id/mises-airdropsvc/app/models/search"
	"github.com/mises-id/mises-airdropsvc/lib/codes"
	"github.com/mises-id/mises-airdropsvc/lib/pagination"
	"github.com/mises-id/mises-airdropsvc/lib/utils"
	socialModel "github.com/mises-id/sns-socialsvc/app/models"
)

type (
	PageChannelUserInput struct {
		PageParams *pagination.TraditionalParams
		Misesid    string
	}
	GetCHannelUserInput struct {
		Misesid string
	}
)

//get channel user
func GetChannelUser(ctx context.Context, in *GetCHannelUserInput) (*models.ChannelUser, error) {

	misesid := in.Misesid
	if misesid == "" {
		return nil, codes.ErrInvalidArgument.Newf("invalid misesid")
	}
	user, err := socialModel.FindUserByMisesid(ctx, utils.AddMisesidProfix(misesid))
	if err != nil {
		return nil, codes.ErrInvalidArgument.Newf(err.Error())
	}
	params := &search.ChannelUserSearch{
		UID: user.UID,
	}
	channel_user, err := models.FindChannelUser(ctx, params)
	if err != nil {
		return nil, codes.ErrNotFound.Newf(err.Error())
	}
	channel_user.User = user
	return channel_user, nil

}

//page channel user
func PageChannelUser(ctx context.Context, in *PageChannelUserInput) ([]*models.ChannelUser, pagination.Pagination, error) {
	if in.Misesid == "" {
		return []*models.ChannelUser{}, &pagination.TraditionalPagination{}, nil
	}
	params := &models.PageChannelUserInput{
		PageParams: in.PageParams,
		Misesid:    in.Misesid,
	}
	res, page, err := models.PageChannelUser(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	return res, page, nil
}
