package airdrop

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
	getListNum              = 5
	userTwitterAuthMaxIdKey = "user_twiter_auth_max_id"
	airdropStop             chan int
	airdropDo               bool
	totalAirdropNum         int
)

type FaucetCallback struct {
	ctx context.Context
}

func AirdropTwitter(ctx context.Context) {
	totalAirdropNum = 20
	airdropStop = make(chan int)
	airdropDo = true
	fmt.Printf("[%s] Airdrop Start\n", time.Now().Local().String())
	airdropLib.AirdropClient.SetListener(&FaucetCallback{ctx})
	go airdropTx(ctx)
	select {
	case <-airdropStop:
		fmt.Printf("[%s] Airdrop End\n", time.Now().Local().String())
	}
	return
}

func airdropToStop() {
	airdropDo = false
	airdropStop <- 1
	return
}

func airdropTx(ctx context.Context) {
	airdrops, err := getAirdropList(ctx)
	if err != nil {
		fmt.Printf("[%s] Airdrop GetAirdropList Error:%s\n", time.Now().Local().String(), err.Error())
		airdropToStop()
		return
	}
	if len(airdrops) == 0 {
		airdropToStop()
		return
	}
	for _, airdrop := range airdrops {
		if err := airdropRun(ctx, airdrop); err != nil {
			fmt.Printf("[%s] Airdrop Run Error:%s \n", time.Now().Local().String(), err.Error())
			airdropToStop()
			return
		}
	}
	return
}

func airdropTxOne(ctx context.Context) {
	airdrop, err := getAirdrop(ctx)
	if err != nil {
		fmt.Printf("[%s] Airdrop Run One GetAirdrop Error:%s \n", time.Now().Local().String(), err.Error())
		airdropToStop()
		return
	}
	if err := airdropRun(ctx, airdrop); err != nil {
		fmt.Printf("[%s] Airdrop Run Error:%s \n", time.Now().Local().String(), err.Error())
		airdropToStop()
		return
	}
	return
}

func getAirdropList(ctx context.Context) ([]*models.Airdrop, error) {
	params := &search.AirdropSearch{
		NotTxID:  true,
		SortType: enum.SortAsc,
		SortKey:  "_id",
		Status:   enum.AirdropDefault,
		ListNum:  int64(getListNum),
	}
	return models.ListAirdrop(ctx, params)
}

//get one
func getAirdrop(ctx context.Context) (*models.Airdrop, error) {
	params := &search.AirdropSearch{
		NotTxID:  true,
		SortType: enum.SortAsc,
		SortKey:  "_id",
		Status:   enum.AirdropDefault,
	}
	return models.FindAirdrop(ctx, params)
}

func airdropRun(ctx context.Context, airdrop *models.Airdrop) error {
	if totalAirdropNum <= 0 {
		return errors.New("run end")
	}
	fmt.Printf("[%s] Airdrop Run num:%d,misesid:%s,coin:%d\n", time.Now().Local().String(), totalAirdropNum, airdrop.Misesid, airdrop.Coin)
	err := airdropLib.AirdropClient.RunAsync(airdrop.Misesid, "", airdrop.Coin)
	if err != nil {
		return err
	}
	totalAirdropNum--
	return pendingAfter(ctx, airdrop.ID)
}

func (cb *FaucetCallback) OnTxGenerated(cmd types.MisesAppCmd) {
	misesid := cmd.MisesUID()
	fmt.Printf("[%s] Mises[%s] Airdrop OnTxGenerated %s\n", time.Now().Local().String(), misesid, cmd.TxID())
	txid := cmd.TxID()
	err := txGeneratedAfter(context.Background(), misesid, txid)
	if err != nil {
		fmt.Printf("[%s] Mises[%s] Airdrop OnTxGenerated After Error:%s\n", time.Now().Local().String(), misesid, err.Error())
	}
}

func (cb *FaucetCallback) OnSucceed(cmd types.MisesAppCmd) {
	misesid := cmd.MisesUID()
	fmt.Printf("[%s] Mises[%s] Airdrop OnSucceed\n", time.Now().Local().String(), misesid)
	err := successAfter(context.Background(), misesid)
	if err != nil {
		fmt.Printf("[%s] Mises[%s] Airdrop OnSucceed After Error:%s\n", time.Now().Local().String(), misesid, err.Error())
	}
	if airdropDo {
		airdropTxOne(cb.ctx)
	}
}

func (cb *FaucetCallback) OnFailed(cmd types.MisesAppCmd, err error) {
	misesid := cmd.MisesUID()
	if err != nil {
		fmt.Printf("[%s] Mises[%s] Airdrop OnFailed: %s\n", time.Now().Local().String(), misesid, err.Error())
	}
	err = failedAfter(context.Background(), misesid)
	if err != nil {
		fmt.Printf("[%s] Mises[%s] Airdrop Onfailed After Error:%s\n", time.Now().Local().String(), misesid, err.Error())
	}
	if airdropDo {
		airdropTxOne(cb.ctx)
	}
}

func successAfter(ctx context.Context, misesid string) error {
	//airdrop update
	params := &search.AirdropSearch{
		Misesid: misesid,
		Type:    enum.AirdropTwitter,
		Status:  enum.AirdropPending,
	}
	airdrop, err := models.FindAirdrop(ctx, params)
	if err != nil {
		fmt.Printf("[%s] Airdrop SuccessAfter FindAirdrop Error:%s \n", time.Now().Local().String(), err.Error())
		return err
	}
	if airdrop.Status != enum.AirdropPending {
		return errors.New("misesid finished")
	}
	if err = airdrop.UpdateStatus(ctx, enum.AirdropSuccess); err != nil {
		return err
	}
	//update user airdrop coin
	if err = updateUserAirdrop(ctx, airdrop.UID, uint64(airdrop.Coin)); err != nil {
		return err
	}
	return nil
}
func failedAfter(ctx context.Context, misesid string) error {
	//airdrop update
	params := &search.AirdropSearch{
		Misesid:  misesid,
		Type:     enum.AirdropTwitter,
		Statuses: []enum.AirdropStatus{enum.AirdropDefault, enum.AirdropPending},
	}
	airdrop, err := models.FindAirdrop(ctx, params)
	if err != nil {
		fmt.Printf("[%s] Airdrop FailedAfter FindAirdrop Error:%s \n", time.Now().Local().String(), err.Error())
		return err
	}
	if airdrop.Status != enum.AirdropPending && airdrop.Status != enum.AirdropDefault {
		return errors.New("airdrop status error")
	}
	if err = airdrop.UpdateStatus(ctx, enum.AirdropFailed); err != nil {
		return err
	}
	return nil
}

func updateUserAirdrop(ctx context.Context, uid uint64, coin uint64) error {
	user_ext, err := socialModel.FindOrCreateUserExt(ctx, uid)
	if err != nil {
		return err
	}
	user_ext.AirdropCoin += coin
	return user_ext.UpdateAirdrop(ctx)
}

func pendingAfter(ctx context.Context, id primitive.ObjectID) error {
	params := &search.AirdropSearch{
		ID:     id,
		Type:   enum.AirdropTwitter,
		Status: enum.AirdropDefault,
	}
	airdrop, err := models.FindAirdrop(ctx, params)
	if err != nil {
		fmt.Printf("[%s] Airdrop PendingAfter FindAirdrop Error:%s \n", time.Now().Local().String(), err.Error())
		return err
	}
	if airdrop.TxID != "" && airdrop.Status != enum.AirdropDefault {
		return errors.New("pending status tx_id exists")
	}
	return airdrop.UpdateStatusPending(ctx)
}

func txGeneratedAfter(ctx context.Context, misesid string, tx_id string) error {
	//update
	params := &search.AirdropSearch{
		Misesid: misesid,
		Type:    enum.AirdropTwitter,
		Status:  enum.AirdropPending,
	}
	airdrop, err := models.FindAirdrop(ctx, params)
	if err != nil {
		fmt.Printf("[%s] Airdrop TxGeneratedAfter FindAirdrop Error:%s \n", time.Now().Local().String(), err.Error())
		return err
	}
	if airdrop.TxID != "" || airdrop.Status != enum.AirdropPending {
		return errors.New("tx_id exists")
	}
	//update
	return airdrop.UpdateTxID(ctx, tx_id)
}
