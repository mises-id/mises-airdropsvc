package user_twitter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mises-id/mises-airdropsvc/app/models"
	"github.com/mises-id/mises-airdropsvc/app/models/enum"
	"github.com/mises-id/mises-airdropsvc/app/models/search"
	"github.com/mises-id/mises-airdropsvc/config/env"
	"github.com/mises-id/mises-airdropsvc/lib/utils"
)

var (
	lookupUserNum       = 10
	sendTweetNum        = 3
	replyTweetNum       = 6
	likeTweetNum        = 10
	followTwitterNum    = 5
	checkTwitterUserNum = 5
)

func init() {
	sendTweetNum = env.Envs.SendTweetNum
	followTwitterNum = env.Envs.FollowTwitterNum
	checkTwitterUserNum = env.Envs.CheckTwitterUserNum
}

func PlanLookupTwitterUser(ctx context.Context) error {
	fmt.Printf("[%s] RunLookupTwitterUser Start\n", time.Now().Local().String())
	err := runLookupTwitterUser(ctx)
	fmt.Printf("[%s] RunLookupTwitterUser End\n", time.Now().Local().String())
	return err
}

func runLookupTwitterUser(ctx context.Context) error {

	//get list
	params := &search.UserTwitterAuthSearch{
		FindTwitterUserState: 1,
		SortType:             enum.SortAsc,
		SortKey:              "_id",
		ListNum:              int64(lookupUserNum),
	}
	user_twitter_list, err := models.ListUserTwitterAuth(ctx, params)
	if err != nil {
		return err
	}
	num := len(user_twitter_list)
	if num <= 0 {
		return nil
	}
	fmt.Printf("[%s] RunLookupTwitterUser %d \n", time.Now().Local().String(), num)
	//do list
	for _, user_twitter := range user_twitter_list {
		if user_twitter.IsAirdrop == true || user_twitter.ValidState == 2 {
			continue
		}
		uid := user_twitter.UID
		var err error
		var twitter_user *models.TwitterUser
		auth_app_name := user_twitter.AuthAppName
		if auth_app_name == "" || auth_app_name == defaultAuthAppName {
			//v2
			twitter_user, err = lookupTwitterUserV2(ctx, user_twitter)
		} else {
			//v1
			twitter_user, err = lookupTwitterUserV1(ctx, user_twitter)
		}
		if err != nil {
			fmt.Printf("[%s] uid[%d] RunLookupTwitterUser AuthAppName[%s] GetTwitterUserById Error:%s \n", time.Now().Local().String(), uid, auth_app_name, err.Error())
			user_twitter.FindTwitterUserState = 3
			if strings.Contains(err.Error(), "httpStatusCode=401") || strings.Contains(err.Error(), "Invalid or expired token") {
				//user_twitter.FindTwitterUserState = 4
				//delete
				models.DeleteUserTwitterAuthByID(ctx, user_twitter.ID)
				continue
			}
			models.UpdateUserTwitterAuthFindState(ctx, user_twitter)
			continue
		}
		user_twitter.TwitterUser = twitter_user
		//follow
		user_twitter.FollowState = 1
		channel_user, err := models.FindChannelUserByUID(ctx, uid)
		var amount int64
		var do_channeluser bool
		var valid_state enum.UserValidState
		user_twitter.ValidState = 3
		valid_state = enum.UserValidFailed
		if channel_user != nil && (channel_user.ValidState == enum.UserValidDefalut || channel_user.ValidState == enum.UserValidFailed) {
			do_channeluser = true
			fmt.Printf("[%s] RunLookupTwitterUser DoChannelUser True UID[%d]\n", time.Now().Local().String(), uid)
		}
		//check user_agent
		if user_twitter.UserAgent != nil {
			//device_id
			if user_twitter.UserAgent.DeviceId != "" {
				device_id := user_twitter.UserAgent.DeviceId
				deviceIDParams := &search.UserTwitterAuthSearch{
					DeviceId: device_id,
				}
				device_id_num, err := models.CountUserTwitterAuth(ctx, deviceIDParams)
				if err == nil {
					user_twitter.DeviceIdNum = device_id_num
					if device_id_num > 1 {
						fmt.Printf("[%s] RunLookupTwitterUser DeviceId:%s, DeviceIdNum: %d\n", time.Now().Local().String(), device_id, device_id_num)
					}
				}
			}
		}
		//check
		followers_count := user_twitter.TwitterUser.FollowersCount
		//is_valid
		if IsValidTwitterUser(user_twitter.TwitterUser) {
			min_check_followers := env.Envs.MinCheckFollowers
			max_check_followers := env.Envs.MaxCheckFollowers
			user_twitter.SendTweeState = 1
			user_twitter.LikeTweeState = 1
			if min_check_followers > 0 && max_check_followers > 0 && followers_count >= min_check_followers && followers_count <= max_check_followers {
				user_twitter.ValidState = 4
				fmt.Printf("[%s] uid[%d] RunLookupTwitterUser CheckValidState FollowersCount[%d]", time.Now().Local().String(), uid, followers_count)
			} else {
				airdropData, err := createAirdrop(ctx, user_twitter)
				if err != nil {
					fmt.Printf("[%s] uid[%d] RunLookupTwitterUser CreateAirdrop Error:%s \n", time.Now().Local().String(), uid, err.Error())
					user_twitter.FindTwitterUserState = 3
					models.UpdateUserTwitterAuthFindState(ctx, user_twitter)
					continue
				}
				user_twitter.Amount = airdropData.Coin
				user_twitter.IsAirdrop = true
				user_twitter.ValidState = 2
				//channel_user
				if do_channeluser {
					amount = user_twitter.Amount / 10
					valid_state = enum.UserValidSucessed
				}
			}
		}
		user_twitter.FindTwitterUserState = 2
		//update
		err = models.UpdateUserTwitterAuthTwitterUser(ctx, user_twitter)
		if err != nil {
			fmt.Printf("[%s] uid[%d] RunLookupTwitterUser UpdateUserTwitterAuthTwitterUser Error:%s \n", time.Now().Local().String(), uid, err.Error())
			continue
		}
		fmt.Printf("[%s] uid[%d] RunLookupTwitterUser AuthAppName[%s] Success \n", time.Now().Local().String(), uid, auth_app_name)
		//do channel_user
		if do_channeluser {
			if err := channel_user.UpdateCreateAirdrop(ctx, valid_state, amount); err != nil {
				fmt.Printf("[%s] RunLookupTwitterUser UpdateChannelUser UID[%d] Error:%s\n", time.Now().Local().String(), uid, err.Error())
			} else {
				fmt.Printf("[%s] RunLookupTwitterUser UpdateChannelUser UID[%d] Success\n", time.Now().Local().String(), uid)
			}
		}
	}
	return nil
}

type PlanSendTweetParams struct {
	AuthAppName string
}

func PlanSendTweet(ctx context.Context, in *PlanSendTweetParams) error {
	fmt.Printf("[%s] RunSendTweet Start \n", time.Now().Local().String())
	err := runSendTweet(ctx, in)
	fmt.Printf("[%s] RunSendTweet End\n", time.Now().Local().String())
	return err
}

func runSendTweet(ctx context.Context, in *PlanSendTweetParams) error {
	//get list
	auth_app_name := defaultAuthAppName
	if in != nil && in.AuthAppName != "" {
		auth_app_name = in.AuthAppName
	}
	params := &search.UserTwitterAuthSearch{
		SendTweetState: 1,
		AuthAppName:    auth_app_name,
		SortBy:         followerSortOrIDAsc(),
		ListNum:        int64(sendTweetNum),
	}
	user_twitter_list, err := models.ListUserTwitterAuth(ctx, params)
	if err != nil {
		return err
	}
	num := len(user_twitter_list)
	if num <= 0 {
		return nil
	}
	fmt.Printf("[%s] RunSendTweet: %d,AuthAppName: %s\n", time.Now().Local().String(), num, auth_app_name)
	//do list
	for _, user_twitter := range user_twitter_list {
		uid := user_twitter.UID
		user_twitter.SendTweeState = 2
		var err error
		auth_app_name := user_twitter.AuthAppName
		if auth_app_name == "" || auth_app_name == defaultAuthAppName {
			//v2
			err = reTweetV2(ctx, user_twitter)
		} else {
			//v1
			err = reTweetV1(ctx, user_twitter)
		}
		if err != nil {
			fmt.Printf("[%s] uid[%d] AuthAppName[%s] Send Tweet Error:%s \n", time.Now().Local().String(), uid, auth_app_name, err.Error())
			user_twitter.SendTweeState = 3
			if strings.Contains(err.Error(), "327") {
				user_twitter.SendTweeState = 2
			}
			if strings.Contains(err.Error(), "httpStatusCode=401") || strings.Contains(err.Error(), "Invalid or expired token") {
				user_twitter.SendTweeState = 4
			}
			if strings.Contains(err.Error(), "httpStatusCode=429") {
				user_twitter.SendTweeState = 5
			}
		}
		if err := models.UpdateUserTwitterAuthSendTweet(ctx, user_twitter); err != nil {
			fmt.Printf("[%s] uid[%d] RunSendTweet UpdateUserTwitterAuthSendTweet Error:%s\n ", time.Now().Local().String(), uid, err.Error())
			continue
		}
		if user_twitter.SendTweeState == 2 {
			fmt.Printf("[%s] uid[%d] RunSendTweet Success AuthAppName[%s]\n", time.Now().Local().String(), uid, auth_app_name)
		}
	}
	return nil
}

func PlanLikeTweet(ctx context.Context) error {
	fmt.Printf("[%s] RunLikeTweet Start\n", time.Now().Local().String())
	err := runLikeTweet(ctx)
	fmt.Printf("[%s] RunLikeTweet End\n", time.Now().Local().String())
	return err
}

func runLikeTweet(ctx context.Context) error {
	//get list
	params := &search.UserTwitterAuthSearch{
		LikeTweetState: 1,
		SortBy:         "followers_count_sort",
		ListNum:        int64(likeTweetNum),
	}
	user_twitter_list, err := models.ListUserTwitterAuth(ctx, params)
	if err != nil {
		return err
	}
	num := len(user_twitter_list)
	if num <= 0 {
		return nil
	}
	fmt.Printf("[%s] RunLikeTweet %d \n", time.Now().Local().String(), num)
	//do list
	for _, user_twitter := range user_twitter_list {
		uid := user_twitter.UID
		//like tweet
		user_twitter.LikeTweeState = 2
		var err error
		auth_app_name := user_twitter.AuthAppName
		if auth_app_name == "" || auth_app_name == defaultAuthAppName {
			//v2
			err = likeTweetV2(ctx, user_twitter)
		} else {
			//v1
			err = likeTweetV1(ctx, user_twitter)
		}
		if err != nil {
			fmt.Printf("[%s] uid[%d] Like Tweet Error:%s \n", time.Now().Local().String(), uid, err.Error())
			user_twitter.LikeTweeState = 3
			if strings.Contains(err.Error(), "httpStatusCode=401") || strings.Contains(err.Error(), "Invalid or expired token") {
				user_twitter.LikeTweeState = 4
			}
			if strings.Contains(err.Error(), "httpStatusCode=429") {
				user_twitter.LikeTweeState = 5
			}
			if strings.Contains(err.Error(), "You have already") {
				user_twitter.LikeTweeState = 2
			}
		}
		if err := models.UpdateUserTwitterAuthLikeTweet(ctx, user_twitter); err != nil {
			fmt.Printf("[%s] uid[%d] RunLikeTweet UpdateUserTwitterAuthLikeTweet Error:%s\n ", time.Now().Local().String(), uid, err.Error())
			continue
		}
		if user_twitter.LikeTweeState == 2 {
			fmt.Printf("[%s] uid[%d] LikeTweet Success AuthAppName[%s] \n", time.Now().Local().String(), uid, auth_app_name)
		}
	}
	return nil
}

func PlanReplyTweet(ctx context.Context) error {
	fmt.Printf("[%s] RunReplyTweet Start\n", time.Now().Local().String())
	err := runReplyTweet(ctx)
	fmt.Printf("[%s] RunReplyTweet End\n", time.Now().Local().String())
	return err
}

func runReplyTweet(ctx context.Context) error {
	//get list
	params := &search.UserTwitterAuthSearch{
		SendTweetState: 1,
		MaxFollower:    1999,
		SortBy:         "id_asc",
		ListNum:        int64(replyTweetNum),
	}
	user_twitter_list, err := models.ListUserTwitterAuth(ctx, params)
	if err != nil {
		return err
	}
	num := len(user_twitter_list)
	if num <= 0 {
		return nil
	}
	fmt.Printf("[%s] RunReplyTweet %d \n", time.Now().Local().String(), num)
	//do list
	for _, user_twitter := range user_twitter_list {
		uid := user_twitter.UID
		user_twitter.SendTweeState = 2
		reply, _ := getReplyText(ctx, user_twitter)
		if err := replyTweet(ctx, user_twitter, reply); err != nil {
			fmt.Printf("[%s] uid[%d] Reply Tweet Error:%s \n", time.Now().Local().String(), uid, err.Error())
			user_twitter.SendTweeState = 3
			if strings.Contains(err.Error(), "httpStatusCode=401") || strings.Contains(err.Error(), "Invalid or expired token") {
				user_twitter.SendTweeState = 4
			}
			if strings.Contains(err.Error(), "httpStatusCode=429") {
				user_twitter.SendTweeState = 5
			}
		}
		if err := models.UpdateUserTwitterAuthSendTweet(ctx, user_twitter); err != nil {
			fmt.Printf("[%s] uid[%d] RunReplyTweet UpdateUserTwitterAuthSendTweet Error:%s\n ", time.Now().Local().String(), uid, err.Error())
			continue
		}
		if user_twitter.SendTweeState == 2 {
			fmt.Printf("[%s] uid[%d] RunReplyTweet Success \n", time.Now().Local().String(), uid)
		}
	}
	return nil
}

func getReplyText(ctx context.Context, user_twitter *models.UserTwitterAuth) (string, error) {
	if user_twitter == nil {
		return "", errors.New("Reply TwitterUser null")
	}
	mis := utils.UMisesToMises(uint64(user_twitter.Amount))
	reply := fmt.Sprintf("I have claimed %.2f $MIS airdrop by using Mises Browser @Mises001, which supports Web3 sites and extensions on mobile.", mis)
	return reply, nil
}

type PlanFollowParams struct {
	AuthAppName string
}

//follow twitter
func FollowTwitter(ctx context.Context, in *PlanFollowParams) error {
	fmt.Printf("[%s] RunFollowTwitter Start\n", time.Now().Local().String())
	err := runFollowTwitter(ctx, in)
	fmt.Printf("[%s] RunFollowTwitter End\n", time.Now().Local().String())
	return err
}

func runFollowTwitter(ctx context.Context, in *PlanFollowParams) error {
	auth_app_name := defaultAuthAppName
	if in != nil && in.AuthAppName != "" {
		auth_app_name = in.AuthAppName
	}
	//get list
	params := &search.UserTwitterAuthSearch{
		FollowState: 1,
		AuthAppName: auth_app_name,
		SortBy:      followerSortOrIDAsc(),
		ListNum:     int64(followTwitterNum),
	}
	user_twitter_list, err := models.ListUserTwitterAuth(ctx, params)
	if err != nil {
		return err
	}
	num := len(user_twitter_list)
	if num <= 0 {
		return nil
	}
	fmt.Printf("[%s] RunFollowTwitter: %d,AuthAppName: %s \n", time.Now().Local().String(), num, auth_app_name)
	//do list
	for _, user_twitter := range user_twitter_list {
		uid := user_twitter.UID
		//to follow
		user_twitter.FollowState = 2
		var err error
		auth_app_name := user_twitter.AuthAppName
		if auth_app_name == "" || auth_app_name == defaultAuthAppName {
			//v2
			err = followTwitterUserV2(ctx, user_twitter, targetTwitterId)
		} else {
			//v1
			err = followTwitterUserV1(ctx, user_twitter, targetTwitterId)
		}
		if err != nil {
			fmt.Printf("[%s] uid[%d] AuthAppName[%s] RunFollowTwitter ApiFollowTwitterUser error:%s\n", time.Now().String(), uid, auth_app_name, err.Error())
			user_twitter.FollowState = 3
			if strings.Contains(err.Error(), "httpStatusCode=401") || strings.Contains(err.Error(), "Invalid or expired token") {
				user_twitter.FollowState = 4
			}
			if strings.Contains(err.Error(), "httpStatusCode=429") || strings.Contains(err.Error(), "429") {
				user_twitter.FollowState = 5
			}
		}
		if err = models.UpdateUserTwitterAuthFollow(ctx, user_twitter); err != nil {
			fmt.Printf("[%s] uid[%d],RunFollowTwitter UpdateUserTwitterAuthFollow Error:%s\n", time.Now().String(), uid, err.Error())
			continue
		}
		if user_twitter.FollowState == 2 {
			fmt.Printf("[%s] uid[%d] RunFollowTwitter Success AuthAppName[%s]\n", time.Now().Local().String(), uid, auth_app_name)
		}
	}
	return nil
}

//check TwitterUser
func PlanCheckTwitterUser(ctx context.Context) error {
	fmt.Printf("[%s] PlanCheckTwitterUser Start\n", time.Now().Local().String())
	err := runCheckTwitterUser(ctx)
	fmt.Printf("[%s] PlanCheckTwitterUser End\n", time.Now().Local().String())
	return err
}

func followerSortOrIDAsc() string {
	sort := "followers_count_sort"
	m := utils.GetRand(1, 100) % 3
	if m == 0 {
		sort = "id_asc"
	}
	return sort
}

func runCheckTwitterUser(ctx context.Context) error {
	//get list

	params := &search.UserTwitterAuthSearch{
		ValidState: 4,
		SortBy:     followerSortOrIDAsc(),
		ListNum:    int64(checkTwitterUserNum),
	}
	user_twitter_list, err := models.ListUserTwitterAuth(ctx, params)
	if err != nil {
		return err
	}
	num := len(user_twitter_list)
	if num <= 0 {
		return nil
	}
	fmt.Printf("[%s] PlanCheckTwitterUser %d \n", time.Now().Local().String(), num)
	//do list
	for _, user_twitter := range user_twitter_list {
		if user_twitter.ValidState != 4 {
			continue
		}
		uid := user_twitter.UID
		if user_twitter.TwitterUser == nil {
			fmt.Printf("[%s] uid[%d],Error PlanCheckTwitterUser TwitterUser is Null\n", time.Now().String(), uid)
			continue
		}
		var err error
		var followerUsers []*models.TwitterUser
		auth_app_name := user_twitter.AuthAppName
		if auth_app_name == "" || auth_app_name == defaultAuthAppName {
			//v2
			followerUsers, err = userFollowersV2(ctx, user_twitter)
		} else {
			//v1
			followerUsers, err = userFollowersV1(ctx, user_twitter)
		}
		if err != nil {
			fmt.Printf("[%s] uid[%d] AuthAppName[%s] PlanCheckTwitterUser UserFollowers Error:%s\n", time.Now().String(), uid, auth_app_name, err.Error())
			user_twitter.ValidState = 6 //check failed
			if strings.Contains(err.Error(), "httpStatusCode=429") || strings.Contains(err.Error(), "429") {
				user_twitter.ValidState = 7 //check 429
			}
			updateUserTwitterAuthTwitterUser(ctx, user_twitter)
			continue
		}
		//check followers
		followersNum := len(followerUsers)
		if followerUsers == nil || followersNum == 0 {
			user_twitter.ValidState = 3 //invalid
			updateUserTwitterAuthTwitterUser(ctx, user_twitter)
			continue
		}
		et := time.Now().UTC().AddDate(0, -3, 0)
		fmt.Printf("[%s] uid[%d],PlanCheckTwitterUser Check ET[%s]\n", time.Now().String(), uid, et.String())
		var zeroTweetNum, zeroFollowerNum, lowFollowerNum, recentRegisterNum, totalFollowerNum int
		for _, followerUser := range followerUsers {

			totalFollowerNum += int(followerUser.FollowersCount)
			if followerUser.TweetCount == 0 {
				zeroTweetNum++
			}
			if followerUser.FollowersCount == 0 {
				zeroFollowerNum++
			}
			if followerUser.FollowersCount <= 5 {
				lowFollowerNum++
			}
			if et.UTC().Unix() < followerUser.CreatedAt.UTC().Unix() {
				recentRegisterNum++
			}
		}
		checkResult := &models.CheckResult{
			CheckNum:          followersNum,
			ZeroTweetNum:      zeroTweetNum,
			ZeroFollowerNum:   zeroFollowerNum,
			LowFollowerNum:    lowFollowerNum,
			RecentRegisterNum: recentRegisterNum,
			TotalFollowerNum:  totalFollowerNum,
		}
		user_twitter.CheckResult = checkResult
		fmt.Printf("[%s] uid[%d] AuthAppName[%s] PlanCheckTwitterUser Check FollowersNum[%d],zeroTweetNum[%d],zeroFollowerNum[%d],lowFollowerNum[%d],recentRegisterNum[%d],totalFollowerNum[%d]\n", time.Now().String(), uid, auth_app_name, followersNum, zeroTweetNum, zeroFollowerNum, lowFollowerNum, recentRegisterNum, totalFollowerNum)
		channel_user, err := models.FindChannelUserByUID(ctx, uid)
		var amount int64
		var do_channeluser bool
		valid_state := enum.UserValidFailed
		if channel_user != nil && (channel_user.ValidState == enum.UserValidDefalut || channel_user.ValidState == enum.UserValidFailed) {
			do_channeluser = true
			fmt.Printf("[%s] PlanCheckTwitterUser DoChannelUser True UID[%d]\n", time.Now().Local().String(), uid)
		}
		airdropData, err := createAirdrop(ctx, user_twitter)
		if err != nil {
			fmt.Printf("[%s] uid[%d] PlanCheckTwitterUser CreateAirdrop Error:%s \n", time.Now().Local().String(), uid, err.Error())
			user_twitter.ValidState = 5
			updateUserTwitterAuthTwitterUser(ctx, user_twitter)
			continue
		}
		user_twitter.Amount = airdropData.Coin
		user_twitter.IsAirdrop = true
		user_twitter.ValidState = 2
		if user_twitter.SendTweeState == 0 {
			user_twitter.SendTweeState = 1
		}
		if user_twitter.LikeTweeState == 0 {
			user_twitter.LikeTweeState = 1
		}
		//channel_user
		if do_channeluser {
			amount = user_twitter.Amount / 10
			valid_state = enum.UserValidSucessed
		}
		//update
		err = updateUserTwitterAuthTwitterUser(ctx, user_twitter)
		if err == nil {
			fmt.Printf("[%s] uid[%d],coin[%d] PlanCheckTwitterUser Success AuthAppName[%s] \n", time.Now().Local().String(), uid, user_twitter.Amount, auth_app_name)
		}
		//do channel_user
		if do_channeluser {
			if err := channel_user.UpdateCreateAirdrop(ctx, valid_state, amount); err != nil {
				fmt.Printf("[%s] PlanCheckTwitterUser UpdateChannelUser UID[%d] Error:%s\n", time.Now().Local().String(), uid, err.Error())
			} else {
				fmt.Printf("[%s] PlanCheckTwitterUser UpdateChannelUser UID[%d] Success\n", time.Now().Local().String(), uid)
			}
		}
	}
	return nil
}

func updateUserTwitterAuthTwitterUser(ctx context.Context, user_twitter *models.UserTwitterAuth) error {
	err := models.UpdateUserTwitterAuthTwitterUser(ctx, user_twitter)
	if err != nil {
		fmt.Printf("[%s] uid[%d] PlanCheckTwitterUser UpdateUserTwitterAuthTwitterUser Error:%s \n", time.Now().Local().String(), user_twitter.UID, err.Error())
	}
	return err
}
