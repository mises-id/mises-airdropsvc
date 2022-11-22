package channel_user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mises-id/mises-airdropsvc/app/models"
	"github.com/mises-id/mises-airdropsvc/app/models/enum"
	"github.com/mises-id/mises-airdropsvc/app/models/search"
	airdropLib "github.com/mises-id/mises-airdropsvc/lib/airdrop"
	"github.com/mises-id/sdk/types"
	socialModel "github.com/mises-id/sns-socialsvc/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	getListNum                 = 5
	channelTwitterAuthMaxIdKey = "channel_twiter_auth_max_id"
	airdropStop                chan int
	airdropDo                  bool
	totalAirdropNum            int
)

type (
	FaucetCallback struct {
		ctx context.Context
	}
)

//airdrop channel
func AirdropChannel(ctx context.Context) {
	totalAirdropNum = 20
	airdropStop = make(chan int)
	airdropDo = true
	fmt.Printf("[%s] Channel Airdrop Start\n", time.Now().Local().String())
	airdropLib.AirdropClient.SetListener(&FaucetCallback{ctx})
	go airdropTx(ctx)
	select {
	case <-airdropStop:
		fmt.Printf("[%s] Channel Airdrop End\n", time.Now().Local().String())
	}
	return
}

func airdropToStop() {
	airdropDo = false
	airdropStop <- 1
	return
}

func airdropTx(ctx context.Context) {
	airdrops, err := getChannelAirdropList(ctx)
	if err != nil {
		airdropToStop()
		return
	}
	if len(airdrops) == 0 {
		airdropToStop()
		return
	}
	for _, airdrop := range airdrops {
		if err := airdropRun(ctx, airdrop); err != nil {
			fmt.Printf("[%s] Channel Airdrop Run Error:%s \n", time.Now().Local().String(), err.Error())
			airdropToStop()
			return
		}
	}
	return
}

func airdropTxOne(ctx context.Context) {
	airdrop, err := getChannelAirdrop(ctx)
	if err != nil {
		airdropToStop()
		return
	}
	if err := airdropRun(ctx, airdrop); err != nil {
		fmt.Printf("[%s] Channel Airdrop Run One Error:%s \n", time.Now().Local().String(), err.Error())
		airdropToStop()
		return
	}
	return
}

func getChannelAirdropList(ctx context.Context) ([]*models.ChannelUser, error) {
	params := &search.ChannelUserSearch{
		ValidStates:   []enum.UserValidState{enum.UserValidSucessed},
		SortType:      enum.SortAsc,
		SortKey:       "_id",
		AirdropStates: []enum.ChannelAirdropState{enum.ChannelAirdropDefault},
		ListNum:       int64(getListNum),
	}
	return models.ListChannelUser(ctx, params)
}

//get one
func getChannelAirdrop(ctx context.Context) (*models.ChannelUser, error) {
	params := &search.ChannelUserSearch{
		ValidStates:   []enum.UserValidState{enum.UserValidSucessed},
		SortType:      enum.SortAsc,
		SortKey:       "_id",
		AirdropStates: []enum.ChannelAirdropState{enum.ChannelAirdropDefault},
	}
	return models.FindChannelUser(ctx, params)
}

func airdropRun(ctx context.Context, channel_user *models.ChannelUser) error {
	if totalAirdropNum <= 0 {
		return errors.New("run end")
	}
	misesid := channel_user.ChannelMisesid
	amount := channel_user.Amount
	trackid := channel_user.ID.Hex()
	fmt.Printf("[%s] Channel Airdrop num:%d,id:%s,coin:%d\n", time.Now().Local().String(), totalAirdropNum, trackid, amount)
	err := airdropLib.AirdropClient.RunAsync(misesid, "", amount, airdropLib.AirdropClient.SetTrackID(trackid))
	if err != nil {
		return err
	}
	totalAirdropNum--
	return pendingAfter(ctx, channel_user.ID)
}

func trackIDToObjectID(trackid string) primitive.ObjectID {

	id, err := primitive.ObjectIDFromHex(trackid)
	if err != nil {
		fmt.Println("trackid error: ", err.Error())
		id = primitive.NilObjectID
	}
	return id
}

func (cb *FaucetCallback) OnTxGenerated(cmd types.MisesAppCmd) {
	trackid := cmd.TrackID()
	id := trackIDToObjectID(trackid)
	fmt.Printf("[%s] ID:%s Channel Airdrop OnTxGenerated %s\n", time.Now().Local().String(), trackid, cmd.TxID())
	txid := cmd.TxID()
	err := txGeneratedAfter(context.Background(), id, txid)
	if err != nil {
		fmt.Printf("[%s] ID:%s Channel Airdrop tx generated after Error:%s \n ", time.Now().Local().String(), trackid, err.Error())
	}
}
func (cb *FaucetCallback) OnSucceed(cmd types.MisesAppCmd) {
	txid := cmd.TxID()
	trackid := cmd.TrackID()
	id := trackIDToObjectID(trackid)
	fmt.Printf("[%s] ID:%s Channel Airdrop OnSucceed %s\n", time.Now().Local().String(), trackid, cmd.TxID())
	err := successAfter(context.Background(), id)
	if err != nil {
		fmt.Printf("[%s] ID:%s,TxID:%s Channel Airdrop tx success after Error:%s \n", time.Now().Local().String(), trackid, txid, err.Error())
	}
	if airdropDo {
		airdropTxOne(cb.ctx)
	}
}

func (cb *FaucetCallback) OnFailed(cmd types.MisesAppCmd, err error) {
	txid := cmd.TxID()
	trackid := cmd.TrackID()
	id := trackIDToObjectID(trackid)
	if err != nil {
		fmt.Printf("[%s] ID:%s,TxID:%s Channel Airdrop OnFailed: %s\n", time.Now().Local().String(), trackid, txid, err.Error())
	}
	err = failedAfter(context.Background(), id, err.Error())
	if err != nil {
		fmt.Printf("[%s] ID:%s,TxID:%s Channel Airdrop tx failed after Error:%s \n", time.Now().Local().String(), trackid, txid, err.Error())
	}
	if airdropDo {
		airdropTxOne(cb.ctx)
	}
}

func successAfter(ctx context.Context, id primitive.ObjectID) error {
	channel_user, err := models.FindChannelUserByID(ctx, id)
	if err != nil {
		fmt.Printf("[%s] Channel Airdrop find channel user Error:%s \n", time.Now().Local().String(), err.Error())
		return err
	}
	if channel_user.AirdropState != enum.ChannelAirdropPending {
		return errors.New("state error")
	}
	if err = channel_user.UpdateStatusSuccess(ctx); err != nil {
		fmt.Println("Channel Airdrop success update Error: ", err.Error())
		return err
	}
	//update user airdrop coin
	if err = updateUserAirdrop(ctx, channel_user.ChannelUID, channel_user.Amount); err != nil {
		fmt.Println("Channel Airdrop success update user ext Error: ", err.Error())
		return err
	}
	return nil
}
func failedAfter(ctx context.Context, id primitive.ObjectID, airdrop_err string) error {
	channel_user, err := models.FindChannelUserByID(ctx, id)
	if err != nil {
		fmt.Println("Channel Airdrop find channel user Error: ", err.Error())
		return err
	}
	if channel_user.AirdropState != enum.ChannelAirdropPending {
		return errors.New("Channel Airdrop state Error")
	}
	if err = channel_user.UpdateStatusFailed(ctx, airdrop_err); err != nil {
		fmt.Println("Channel Airdrop failed update Error: ", err.Error())
		return err
	}
	return nil
}

func updateUserAirdrop(ctx context.Context, uid uint64, coin int64) error {
	user_ext, err := socialModel.FindOrCreateUserExt(ctx, uid)
	if err != nil {
		return err
	}
	user_ext.ChannelAirdropCoin += uint64(coin)

	return user_ext.UpdateChannelAirdrop(ctx)
}

func pendingAfter(ctx context.Context, id primitive.ObjectID) error {
	channel_user, err := models.FindChannelUserByID(ctx, id)
	if err != nil {
		fmt.Printf("id[%s],Channel Airdrop pending after find channel user Error: %s \n", id.Hex(), err.Error())
		return err
	}
	if channel_user.TxID != "" && channel_user.AirdropState != enum.ChannelAirdropDefault {
		return errors.New("Channel Airdrop pending state tx_id exists")
	}
	err = channel_user.UpdateStatusPending(ctx)
	if err != nil {
		fmt.Println("Channel Airdrop UpdateStatusPending Error: ", err.Error())
		return err
	}
	return err
}

func txGeneratedAfter(ctx context.Context, id primitive.ObjectID, tx_id string) error {
	channel_user, err := models.FindChannelUserByID(ctx, id)
	if err != nil {
		fmt.Println("Channel Airdrop find  error: ", err.Error())
		return err
	}
	if channel_user.TxID != "" || channel_user.AirdropState != enum.ChannelAirdropPending {
		return errors.New("tx_id exists")
	}
	//update
	return channel_user.UpdateTxID(ctx, tx_id)
}
