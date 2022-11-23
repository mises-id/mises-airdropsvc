package models

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/mises-id/mises-airdropsvc/app/models/enum"
	"github.com/mises-id/mises-airdropsvc/app/models/search"
	"github.com/mises-id/mises-airdropsvc/lib/codes"
	"github.com/mises-id/mises-airdropsvc/lib/db"
	socialModel "github.com/mises-id/sns-socialsvc/app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ChannelExist = errors.New("channel exist")
)

type (
	ChannelList struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"`
		UID       uint64             `bson:"uid"`
		Misesid   string             `bson:"misesid"`
		State     enum.State         `bson:"state"` //state: open or close
		CreatedAt time.Time          `bson:"created_at"`
		User      *socialModel.User  `bson:"-"`
	}
)

func FindChannelList(ctx context.Context, params ISearchParams) (*ChannelList, error) {

	res := &ChannelList{}
	chain := params.BuildSearchParams(db.ODM(ctx))
	err := chain.Last(res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func FindChannelListByMisesid(ctx context.Context, misesid string) (*ChannelList, error) {

	params := &search.ChannelListSearch{
		Misesid: misesid,
	}
	return FindChannelList(ctx, params)
}
func FindChannelListByID(ctx context.Context, id primitive.ObjectID) (*ChannelList, error) {

	params := &search.ChannelListSearch{
		ID: id,
	}
	return FindChannelList(ctx, params)
}

func ListChannelList(ctx context.Context, params ISearchParams) ([]*ChannelList, error) {

	res := make([]*ChannelList, 0)
	chain := params.BuildSearchParams(db.ODM(ctx))
	err := chain.Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (this *ChannelList) BeforeCreate(ctx context.Context) error {
	var lc sync.Mutex
	lc.Lock()
	this.ID = primitive.NilObjectID
	this.CreatedAt = time.Now()
	query := db.ODM(ctx).Where(bson.M{"uid": this.UID})

	var c int64
	err := query.Model(this).Count(&c).Error
	lc.Unlock()
	if err != nil {
		return err
	}
	if c > 0 {
		return ChannelExist
	}
	return nil
}

func CreateChannelByMisesid(ctx context.Context, misesid string) (*ChannelList, error) {
	// find user by misesid
	user, err := socialModel.FindUserByMisesid(ctx, misesid)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = codes.ErrNotFound.Newf("misesid not exist")
		}
		return nil, err
	}
	return CreateChannelByUser(ctx, user)
}

func CreateChannelByUser(ctx context.Context, user *socialModel.User) (*ChannelList, error) {

	channel := &ChannelList{
		UID:     user.UID,
		Misesid: user.Misesid,
		State:   enum.StateOpen,
	}
	return CreateChannelList(ctx, channel)
}

func CreateChannelList(ctx context.Context, data *ChannelList) (*ChannelList, error) {

	if err := data.BeforeCreate(ctx); err != nil {
		return nil, err
	}
	res, err := db.DB().Collection("channellists").InsertOne(ctx, data)
	data.ID = res.InsertedID.(primitive.ObjectID)
	return data, err
}
