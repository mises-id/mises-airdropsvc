package user_twitter

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	twitterV1 "github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/resources"
	"github.com/michimani/gotwi/tweets"
	tweetsType "github.com/michimani/gotwi/tweets/types"
	"github.com/michimani/gotwi/users"
	"github.com/michimani/gotwi/users/types"
	usersType "github.com/michimani/gotwi/users/types"
	"github.com/mises-id/mises-airdropsvc/app/models"
	"github.com/mises-id/mises-airdropsvc/app/models/enum"
	"github.com/mises-id/mises-airdropsvc/config/env"
	"github.com/mises-id/mises-airdropsvc/lib/codes"
	"github.com/mises-id/mises-airdropsvc/lib/utils"
	socialModel "github.com/mises-id/sns-socialsvc/app/models"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	CallbackStateFlag            = "mises&=mises"
	callbackBase                 = "https://api.alb.mises.site"
	callbackPath                 = "api/v1/twitter/callback"
	FollowV1Endpoint             = "https://api.twitter.com/1.1/friendships/create.json"
	RequestTokenEndpoint         = "https://api.twitter.com/oauth/request_token"
	AccessTokenEndpoint          = "https://api.twitter.com/oauth/access_token"
	AuthEndpoint                 = "https://api.twitter.com/oauth/authorize"
	OAuthVersion10               = "1.0"
	OAuthSignatureMethodHMACSHA1 = "HMAC-SHA1"
	oauth1header                 = `OAuth oauth_callback="%s",oauth_consumer_key="%s",oauth_nonce="%s",oauth_signature="%s",oauth_signature_method="%s",oauth_timestamp="%s",oauth_token="%s",oauth_version="%s"`
	oauth1Userheader             = `OAuth oauth_consumer_key="%s",oauth_nonce="%s",oauth_signature="%s",oauth_signature_method="%s",oauth_timestamp="%s",oauth_token="%s",oauth_version="%s"`
)

var (
	OAuthConsumerKey    = ""
	OAuthConsumerSecret = ""
	OAuthToken          = ""
	OAuthTokenSecret    = ""
	targetTwitterId     = "1442753558311424001"
	targetRetweetID     = "1591980699623776256"
	validRegisterDate   string
	defaultAuthAppName  = "v2.network"
)

type (
	CreateOAuthSignatureInput struct {
		HTTPMethod       string
		RawEndpoint      string
		OAuthConsumerKey string
		OAuthToken       string
		SigningKey       string
		ParameterMap     map[string]string
	}
	CreateOAuthSignatureOutput struct {
		OAuthNonce           string
		OAuthSignatureMethod string
		OAuthTimestamp       string
		OAuthVersion         string
		OAuthSignature       string
	}
	Endpoint     string
	EndpointInfo struct {
		Raw                      string
		Base                     string
		EncodedQueryParameterMap map[string]string
	}
	AirdropInfoOutput struct {
		Twitter *models.UserTwitterAuth
		Airdrop *models.Airdrop
	}
	CallbackParams struct {
		OauthToken, OauthVerifier, State string
		UserAgent                        *models.UserAgent
	}
	GetTwitterAuthUrlParams struct {
		UID      uint64
		DeviceId string
	}
	AuthConfigParams struct {
		name, key, secret string
	}
	RequestTokenParams struct {
		callback    string
		auth_config *AuthConfigParams
	}
)

func init() {
	/* OAuthConsumerKey = env.Envs.GOTWI_API_KEY
	OAuthConsumerSecret = env.Envs.GOTWI_API_KEY_SECRET */
	validRegisterDate = env.Envs.VALID_TWITTER_REGISTER_DATE
}

func getAuthConfigByRand() (*AuthConfigParams, error) {
	config_list := getAuthConfigList()
	num := len(config_list)
	if num == 0 {
		return nil, errors.New("no auth config")
	}
	if num == 1 {
		return config_list[0], nil
	}
	index := utils.GetRand(1, 100) % num
	return config_list[index], nil
}

func getAuthConfigByName(name string) (*AuthConfigParams, error) {
	config_list := getAuthConfigList()
	for _, v := range config_list {
		if v.name == name {
			return v, nil
		}
	}
	if name == defaultAuthAppName {
		config := &AuthConfigParams{
			name:   defaultAuthAppName,
			key:    env.Envs.GOTWI_API_KEY,
			secret: env.Envs.GOTWI_API_KEY_SECRET,
		}
		return config, nil
	}
	return nil, errors.New(fmt.Sprintf("can not find config by %s", name))
}

func getAuthConfigList() []*AuthConfigParams {
	name_list := env.Envs.TWI_APP_NAME_LIST
	key_list := env.Envs.TWI_APP_KEY_LIST
	secret_list := env.Envs.TWI_APP_SECRET_LIST
	num := len(name_list)
	if num == 0 || len(key_list) != num || len(secret_list) != num {
		res := make([]*AuthConfigParams, 1)
		res[0] = &AuthConfigParams{
			name:   defaultAuthAppName,
			key:    env.Envs.GOTWI_API_KEY,
			secret: env.Envs.GOTWI_API_KEY_SECRET,
		}
		return res
	}
	res := make([]*AuthConfigParams, num)
	for i := 0; i < num; i++ {
		config := &AuthConfigParams{
			name:   name_list[i],
			key:    key_list[i],
			secret: secret_list[i],
		}
		res[i] = config
	}
	return res
}

//get twitter auth url
func GetTwitterAuthUrl(ctx context.Context, in *GetTwitterAuthUrlParams) (string, error) {
	uid := in.UID
	device_id := in.DeviceId
	baseUrl, err := url.Parse(callbackBase)
	if err != nil {
		return "", err
	}
	baseUrl.Path = callbackPath
	v := url.Values{}
	auth_config, err := getAuthConfigByRand()
	if err != nil {
		return "", err
	}
	rqi := &RequestTokenParams{
		auth_config: auth_config,
	}
	v.Add("state", fmt.Sprintf("%d%s%s%s%s", uid, CallbackStateFlag, device_id, CallbackStateFlag, auth_config.name))
	baseUrl.RawQuery = v.Encode()
	rqi.callback = baseUrl.String()
	auth, err := RequestToken(ctx, rqi)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s?%s", AuthEndpoint, auth)
	return url, nil
}

//get airdrop info
func GetAirdropInfo(ctx context.Context, uid uint64) (*AirdropInfoOutput, error) {
	user_twitter, err := models.FindUserTwitterAuthByUid(ctx, uid)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	if user_twitter != nil {
		check_state := "pending"
		if user_twitter.ValidState == 2 {
			user_twitter.IsValid = true
			check_state = "valid"
		}
		if user_twitter.ValidState == 3 {
			user_twitter.IsValid = true
			check_state = "invalid"
		}
		user_twitter.CheckState = check_state
		if user_twitter.CheckState == "invalid" {
			invalid_code := "created"
			reason := "This Account was created after May. 1, 2022"
			if user_twitter.TwitterUser != nil && user_twitter.TwitterUser.FollowersCount == 0 {
				reason = "Insufficient social data in this Twitter account"
				invalid_code = "followers"
			}
			user_twitter.InvalidCode = invalid_code
			user_twitter.Reason = reason
		}
	}
	airdrop, err := models.FindAirdropByUid(ctx, uid)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	res := &AirdropInfoOutput{
		Twitter: user_twitter,
		Airdrop: airdrop,
	}
	return res, nil

}

func IsValidTwitterUser(twitter_user *models.TwitterUser) (is_valid bool) {
	if twitter_user == nil {
		return is_valid
	}
	if twitter_user.FollowersCount == 0 {
		return is_valid
	}
	validRegisterDate = env.Envs.VALID_TWITTER_REGISTER_DATE
	timeFormat := "2006-01-02"
	st, _ := time.Parse(timeFormat, validRegisterDate)
	vt := st.Unix()
	twitterUserCreatedAt := twitter_user.CreatedAt.Unix()
	if vt >= twitterUserCreatedAt {
		is_valid = true
	}
	return is_valid
}

func GetTwitterAirdropCoin(ctx context.Context, userTwitter *models.UserTwitterAuth) int64 {
	if userTwitter == nil || userTwitter.TwitterUser == nil {
		return 0
	}
	followers_count := userTwitter.TwitterUser.FollowersCount
	following_count := userTwitter.TwitterUser.FollowingCount
	tweet_count := userTwitter.TwitterUser.TweetCount
	if followers_count == 0 {
		return 0
	}
	var max, umises, mises, perFollowerMises, score uint64
	umises = 1
	score = 100
	mises = 1000000 * umises
	//do score
	if tweet_count == 0 || following_count == 0 {
		score = 1
	}
	if tweet_count <= 10 || following_count <= 10 {
		score = 10
	}
	if followers_count >= env.Envs.FollowsMinFollow {
		if following_count*10 >= followers_count*5 {
			score = 10
		}
	}
	if userTwitter.UserAgent != nil && userTwitter.UserAgent.Browser != "" {
		browser := userTwitter.UserAgent.Browser
		if browser != "Chrome 105.0.0.0" && browser != "Chrome 98.0.4745.25" {
			return int64(mises / 10)
		}
	}
	//followers quality
	if userTwitter.CheckResult != nil && userTwitter.CheckResult.CheckNum > 0 {
		checkNum := userTwitter.CheckResult.CheckNum
		/* if userTwitter.CheckResult.LowFollowerNum*2 >= checkNum*1 || userTwitter.CheckResult.ZeroTweetNum*5 >= checkNum*2 || userTwitter.CheckResult.ZeroFollowerNum*5 >= checkNum*2 {
			score = 1
		} */
		if userTwitter.CheckResult.LowFollowerNum*5 >= checkNum*1 || userTwitter.CheckResult.ZeroTweetNum*5 >= checkNum*1 || userTwitter.CheckResult.ZeroFollowerNum*5 >= checkNum*1 {
			score = 1
		}
		if userTwitter.CheckResult.RecentRegisterNum*5 >= checkNum*1 {
			score = 1
		}
		/* if userTwitter.CheckResult.TotalFollowerNum >= 1000*checkNum {
			score = 100
		} */
		if followers_count >= 200 && checkNum < 180 {
			score = 1
		}
	}
	if userTwitter.UserAgent != nil && userTwitter.UserAgent.DeviceId != "" && userTwitter.DeviceIdNum > 1 {
		score = 1
	}
	if score <= 0 {
		score = 1
	}
	if score > 100 {
		score = 100
	}
	perFollowerMises = 1000 * score / 100
	max = 10 * mises
	coin := mises + perFollowerMises*umises*followers_count
	if coin > max {
		coin = max
	}
	return int64(coin)
}

//send tweet
func sendTweetV2(ctx context.Context, user_twitter *models.UserTwitterAuth, tweet string) error {

	if user_twitter.OauthToken == "" || user_twitter.OauthTokenSecret == "" {
		return codes.ErrForbidden.Newf("OAuthToken and OAuthTokenSecret is required")
	}
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return err
	}
	key := auth_config.key
	secret := auth_config.secret
	transport := &http.Transport{Proxy: setProxy()}
	client := &http.Client{Transport: transport}
	in := &gotwi.NewGotwiClientInput{
		HTTPClient:           client,
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           user_twitter.OauthToken,
		OAuthTokenSecret:     user_twitter.OauthTokenSecret,
		OAuthConsumerKey:     key,
		OAuthConsumerSecret:  secret,
	}
	twitter_client, err := gotwi.NewGotwiClient(in)
	if err != nil {
		return err
	}
	params := &tweetsType.ManageTweetsPostParams{
		Text: &tweet,
	}
	_, err = tweets.ManageTweetsPost(ctx, twitter_client, params)

	return err
}

//retweet
func reTweetV2(ctx context.Context, user_twitter *models.UserTwitterAuth) error {

	if user_twitter.OauthToken == "" || user_twitter.OauthTokenSecret == "" {
		return codes.ErrForbidden.Newf("OAuthToken and OAuthTokenSecret is required")
	}
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return err
	}
	key := auth_config.key
	secret := auth_config.secret
	transport := &http.Transport{Proxy: setProxy()}
	client := &http.Client{Transport: transport}
	in := &gotwi.NewGotwiClientInput{
		HTTPClient:           client,
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           user_twitter.OauthToken,
		OAuthTokenSecret:     user_twitter.OauthTokenSecret,
		OAuthConsumerKey:     key,
		OAuthConsumerSecret:  secret,
	}
	twitter_client, err := gotwi.NewGotwiClient(in)
	if err != nil {
		return err
	}
	params := &tweetsType.TweetRetweetsPostParams{
		ID:      user_twitter.TwitterUserId,
		TweetID: &targetRetweetID,
	}
	_, err = tweets.TweetRetweetsPost(ctx, twitter_client, params)

	return err
}

//reply tweet
func replyTweet(ctx context.Context, user_twitter *models.UserTwitterAuth, reply string) error {

	if user_twitter.OauthToken == "" || user_twitter.OauthTokenSecret == "" {
		return codes.ErrForbidden.Newf("OAuthToken and OAuthTokenSecret is required")
	}
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return err
	}
	key := auth_config.key
	secret := auth_config.secret
	transport := &http.Transport{Proxy: setProxy()}
	client := &http.Client{Transport: transport}
	in := &gotwi.NewGotwiClientInput{
		HTTPClient:           client,
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           user_twitter.OauthToken,
		OAuthTokenSecret:     user_twitter.OauthTokenSecret,
		OAuthConsumerKey:     key,
		OAuthConsumerSecret:  secret,
	}
	twitter_client, err := gotwi.NewGotwiClient(in)
	if err != nil {
		return err
	}
	params := &tweetsType.ManageTweetsPostParams{
		Text: &reply,
		Reply: &tweetsType.ManageTweetsPostParamsReply{
			InReplyToTweetID: targetRetweetID,
		},
	}
	_, err = tweets.ManageTweetsPost(ctx, twitter_client, params)

	return err
}

//like tweet
func likeTweetV2(ctx context.Context, user_twitter *models.UserTwitterAuth) error {

	if user_twitter.OauthToken == "" || user_twitter.OauthTokenSecret == "" {
		return codes.ErrForbidden.Newf("OAuthToken and OAuthTokenSecret is required")
	}
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return err
	}
	key := auth_config.key
	secret := auth_config.secret
	transport := &http.Transport{Proxy: setProxy()}
	client := &http.Client{Transport: transport}
	in := &gotwi.NewGotwiClientInput{
		HTTPClient:           client,
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           user_twitter.OauthToken,
		OAuthTokenSecret:     user_twitter.OauthTokenSecret,
		OAuthConsumerKey:     key,
		OAuthConsumerSecret:  secret,
	}
	twitter_client, err := gotwi.NewGotwiClient(in)
	if err != nil {
		return err
	}
	params := &tweetsType.TweetLikesPostParams{
		ID:      user_twitter.TwitterUserId,
		TweetID: &targetRetweetID,
	}
	_, err = tweets.TweetLikesPost(ctx, twitter_client, params)

	return err
}

//apiFollowTwitterUser
func followTwitterUserV2(ctx context.Context, user_twitter *models.UserTwitterAuth, target_user_id string) error {
	if user_twitter == nil {
		return errors.New("user_twitter is null")
	}
	if user_twitter.OauthToken == "" || user_twitter.OauthTokenSecret == "" {
		return codes.ErrForbidden.Newf("OAuthToken and OAuthTokenSecret is required")
	}
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return err
	}
	key := auth_config.key
	secret := auth_config.secret
	transport := &http.Transport{Proxy: setProxy()}
	client := &http.Client{Transport: transport}
	in := &gotwi.NewGotwiClientInput{
		HTTPClient:           client,
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           user_twitter.OauthToken,
		OAuthTokenSecret:     user_twitter.OauthTokenSecret,
		OAuthConsumerKey:     key,
		OAuthConsumerSecret:  secret,
	}
	twitter_client, err := gotwi.NewGotwiClient(in)
	if err != nil {
		return err
	}
	params := &types.FollowsFollowingPostParams{
		ID:           user_twitter.TwitterUserId,
		TargetUserID: &target_user_id,
	}
	_, err = users.FollowsFollowingPost(ctx, twitter_client, params)

	return err
}

func createAirdrop(ctx context.Context, user_twitter *models.UserTwitterAuth) (*models.Airdrop, error) {
	coin := GetTwitterAirdropCoin(ctx, user_twitter)
	if coin <= 0 {
		return nil, errors.New("coin is zero")
	}
	airdropAdd := &models.Airdrop{
		UID:       user_twitter.UID,
		Misesid:   user_twitter.Misesid,
		Status:    enum.AirdropDefault,
		Type:      enum.AirdropTwitter,
		Coin:      coin,
		TxID:      "",
		CreatedAt: time.Now(),
	}
	return models.CreateAirdrop(ctx, airdropAdd)
}

func getTwitterCallbackUrl(code, username, misesid string) string {
	return env.Envs.TwitterAuthSuccessCallback + "?code=" + code + "&username=" + username + "&misesid=" + misesid
}

//twitter auth callback
func TwitterCallback(ctx context.Context, in *CallbackParams) string {

	var (
		callback0 string = getTwitterCallbackUrl("0", "", "")
		callback1 string = getTwitterCallbackUrl("1", "", "")
		callback2 string = getTwitterCallbackUrl("2", "", "")
	)
	state := in.State
	oauth_token := in.OauthToken
	oauth_verifier := in.OauthVerifier
	stateArr := strings.Split(in.State, CallbackStateFlag)
	stateArrlen := len(stateArr)
	if stateArrlen < 2 {
		fmt.Printf("[%s] Callback State[%s] Invalid \n", time.Now().Local().String(), state)
		return callback2
	}
	//uid
	uid, _ := strconv.ParseUint(stateArr[0], 10, 64)
	if uid == 0 {
		fmt.Printf("[%s] Callback State[%s] User Invalid \n", time.Now().Local().String(), state)
		return callback2
	}
	//device_id
	device_id := stateArr[1]
	if in.UserAgent != nil {
		in.UserAgent.DeviceId = device_id
	}
	//auth_app_name
	auth_app_name := defaultAuthAppName
	if stateArrlen > 2 {
		state_auth_app_name := stateArr[2]
		_, err := getAuthConfigByName(state_auth_app_name)
		if err != nil {
			fmt.Printf("[%s] Callback State[%s] AppName Invalid \n", time.Now().Local().String(), state)
			return callback2
		}
		auth_app_name = state_auth_app_name
	}
	if oauth_token == "" || oauth_verifier == "" {
		fmt.Printf("[%s] Oauth_token[%s],oauth_verifier[%s] Empty \n", time.Now().Local().String(), oauth_token, oauth_verifier)
		return callback2
	}
	user, err := socialModel.FindUser(ctx, uid)
	if err != nil {
		fmt.Printf("[%s] Twitter callback find user Error: %s \n", time.Now().Local().String(), err.Error())
		return callback2
	}
	userMisesid := user.Misesid
	callback2 = getTwitterCallbackUrl("2", "", userMisesid)
	//find twitter user
	access_token, err := AccessToken(ctx, oauth_token, oauth_verifier)
	if err != nil {
		fmt.Printf("[%s] Twitter callback access token Error:%s \n", time.Now().Local().String(), err.Error())
		return callback2
	}
	params, _ := url.ParseQuery(access_token)
	user_ids, ok := params["user_id"]
	if !ok || len(user_ids) <= 0 {
		fmt.Printf("[%s] Twitter callback user_id Error:%s \n", time.Now().Local().String(), err.Error())
		return callback2
	}
	oauth_tokens, ok := params["oauth_token"]
	oauth_token_secrets, ok := params["oauth_token_secret"]
	twitter_user_id := user_ids[0]
	oauth_token_new := oauth_tokens[0]
	oauth_token_secret := oauth_token_secrets[0]
	//check twitter_user_id
	twitter_auth, err := models.FindUserTwitterAuthByTwitterUserId(ctx, twitter_user_id)

	if twitter_auth != nil && twitter_auth.UID != uid {
		callback1 = getTwitterCallbackUrl("1", twitter_auth.TwitterUser.UserName, userMisesid)
		fmt.Printf("[%s] FindUserTwitterAuthByTwitterUserId exist uid[%d],username[%s]\n ", time.Now().Local().String(), uid, twitter_auth.TwitterUser.UserName)
		return callback1
	}
	//check uid
	user_twitter, err := models.FindUserTwitterAuthByUid(ctx, uid)
	if err != nil && err != mongo.ErrNoDocuments {
		fmt.Printf("[%s] Twitter callback FindUserTwitterAuthByUid Error:%s \n", time.Now().Local().String(), err.Error())
		return callback2
	}
	callback0 = getTwitterCallbackUrl("0", "", userMisesid)
	//check airdrop
	airdrop, err := models.FindAirdropByUid(ctx, uid)

	if user_twitter == nil {
		//create
		if airdrop != nil {
			fmt.Printf("[%s] Twitter callback airdrop exist uid[%d]\n", time.Now().Local().String(), uid)
			return callback0
		}
		add := &models.UserTwitterAuth{
			UID:                  uid,
			Misesid:              user.Misesid,
			TwitterUserId:        twitter_user_id,
			FindTwitterUserState: 1,
			OauthToken:           oauth_token_new,
			OauthTokenSecret:     oauth_token_secret,
			UserAgent:            in.UserAgent,
			AuthAppName:          auth_app_name,
		}
		err = models.CreateUserTwitterAuth(ctx, add)
		if err != nil {
			fmt.Printf("[%s] Twitter Callback Create Error: %s \n", time.Now().Local().String(), err.Error())
		}

	} else {
		//update
		/*user_twitter.OauthToken = oauth_token_new
		user_twitter.OauthTokenSecret = oauth_token_secret
		 if airdrop == nil && user_twitter.ValidState != 3 {
			user_twitter.TwitterUserId = twitter_user_id
			user_twitter.FindTwitterUserState = 1
		}
		err = models.UpdateUserTwitterAuth(ctx, user_twitter)*/
	}
	return callback0
}

func setProxy() func(*http.Request) (*url.URL, error) {
	return func(_ *http.Request) (*url.URL, error) {
		return nil, nil
		return url.Parse("http://127.0.0.1:1087")
	}
}

//get twitter auth request_token
func RequestToken(ctx context.Context, rqin *RequestTokenParams) (string, error) {
	if rqin == nil {
		return "", errors.New("RequestTokenParams is null")
	}
	callback := rqin.callback
	key := rqin.auth_config.key
	secret := rqin.auth_config.secret
	api := fmt.Sprintf("%s?oauth_callback=%s", RequestTokenEndpoint, callback)
	transport := &http.Transport{Proxy: setProxy()}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest("POST", api, nil)
	ParameterMap := map[string]string{
		"oauth_callback": callback,
	}
	siginkey := fmt.Sprintf("%s&%s", secret, "")
	in := &CreateOAuthSignatureInput{
		HTTPMethod:       req.Method,
		RawEndpoint:      req.URL.String(),
		OAuthConsumerKey: key,
		OAuthToken:       "",
		SigningKey:       siginkey,
		ParameterMap:     ParameterMap,
	}
	out, err := CreateOAuthSignature(in)
	if err != nil {
		return "", err
	}
	auth := fmt.Sprintf(oauth1header,
		url.QueryEscape(callback),
		url.QueryEscape(key),
		url.QueryEscape(out.OAuthNonce),
		url.QueryEscape(out.OAuthSignature),
		url.QueryEscape(out.OAuthSignatureMethod),
		url.QueryEscape(out.OAuthTimestamp),
		url.QueryEscape(OAuthToken),
		url.QueryEscape(out.OAuthVersion),
	)
	req.Header.Add("Authorization", auth)
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", errors.New(res.Status)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return string(body), nil
}

//v1
func reTweetV1(ctx context.Context, user_twitter *models.UserTwitterAuth) error {
	if user_twitter == nil || user_twitter.TwitterUser == nil {
		return errors.New("Twitter User is null")
	}
	tweet_id, _ := strconv.ParseInt(targetRetweetID, 10, 64)
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return err
	}
	key := auth_config.key
	secret := auth_config.secret
	config := oauth1.NewConfig(key, secret)
	token := oauth1.NewToken(user_twitter.OauthToken, user_twitter.OauthTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitterV1.NewClient(httpClient)
	params := &twitterV1.StatusRetweetParams{
		ID: tweet_id,
	}
	_, _, err = client.Statuses.Retweet(tweet_id, params)
	return err
}

//v1 like
func likeTweetV1(ctx context.Context, user_twitter *models.UserTwitterAuth) error {
	if user_twitter == nil || user_twitter.TwitterUser == nil {
		return errors.New("Twitter User is null")
	}
	tweet_id, _ := strconv.ParseInt(targetRetweetID, 10, 64)
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return err
	}
	key := auth_config.key
	secret := auth_config.secret
	config := oauth1.NewConfig(key, secret)
	token := oauth1.NewToken(user_twitter.OauthToken, user_twitter.OauthTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitterV1.NewClient(httpClient)
	params := &twitterV1.FavoriteCreateParams{
		ID: tweet_id,
	}
	_, _, err = client.Favorites.Create(params)
	return err
}

//user followers
func userFollowersV2(ctx context.Context, user_twitter *models.UserTwitterAuth) ([]*models.TwitterUser, error) {
	if user_twitter.OauthToken == "" || user_twitter.OauthTokenSecret == "" {
		return nil, codes.ErrForbidden.Newf("OAuthToken and OAuthTokenSecret is required")
	}
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return nil, err
	}
	key := auth_config.key
	secret := auth_config.secret
	transport := &http.Transport{Proxy: setProxy()}
	client := &http.Client{Transport: transport}
	in := &gotwi.NewGotwiClientInput{
		HTTPClient:           client,
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           user_twitter.OauthToken,
		OAuthTokenSecret:     user_twitter.OauthTokenSecret,
		OAuthConsumerKey:     key,
		OAuthConsumerSecret:  secret,
	}
	twitter_client, err := gotwi.NewGotwiClient(in)
	if err != nil {
		return nil, err
	}
	params := &usersType.FollowsFollowersParams{
		ID:         user_twitter.TwitterUserId,
		MaxResults: usersType.FollowsMaxResults(env.Envs.FollowsMaxResults),
		UserFields: fields.UserFieldList{
			fields.UserFieldCreatedAt,
			fields.UserFieldPublicMetrics,
		},
	}
	followers, err := users.FollowsFollowers(ctx, twitter_client, params)
	if err != nil {
		return nil, err
	}
	res := make([]*models.TwitterUser, len(followers.Data))
	for i, follower := range followers.Data {
		res[i] = buildV2User(&follower)
	}
	return res, nil
}

func buildV2User(twitter_user *resources.User) *models.TwitterUser {
	if twitter_user == nil {
		return nil
	}
	user := &models.TwitterUser{
		TwitterUserId:  *twitter_user.ID,
		UserName:       *twitter_user.Username,
		Name:           *twitter_user.Name,
		CreatedAt:      *twitter_user.CreatedAt,
		FollowersCount: uint64(*twitter_user.PublicMetrics.FollowersCount),
		FollowingCount: uint64(*twitter_user.PublicMetrics.FollowingCount),
		TweetCount:     uint64(*twitter_user.PublicMetrics.TweetCount),
	}
	return user
}

//v1 followers
func userFollowersV1(ctx context.Context, user_twitter *models.UserTwitterAuth) ([]*models.TwitterUser, error) {
	if user_twitter == nil || user_twitter.TwitterUser == nil {
		return nil, errors.New("Twitter User is null")
	}
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return nil, err
	}
	key := auth_config.key
	secret := auth_config.secret
	config := oauth1.NewConfig(key, secret)
	token := oauth1.NewToken(user_twitter.OauthToken, user_twitter.OauthTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitterV1.NewClient(httpClient)
	user_id, _ := strconv.ParseInt(user_twitter.TwitterUserId, 10, 64)
	countF := int(env.Envs.FollowsMaxResults)
	if countF > 200 {
		countF = 200
	}
	params := &twitterV1.FollowerListParams{
		UserID: user_id,
		Count:  countF,
	}
	followers, _, err := client.Followers.List(params)
	if err != nil {
		return nil, err
	}
	res := make([]*models.TwitterUser, len(followers.Users))
	for i, twitter_user := range followers.Users {
		res[i] = buildV1User(&twitter_user)
	}
	return res, nil
}

func buildV1User(twitter_user *twitterV1.User) *models.TwitterUser {
	if twitter_user == nil {
		return nil
	}
	ct, _ := time.Parse(time.RubyDate, twitter_user.CreatedAt)
	user := &models.TwitterUser{
		TwitterUserId:  strconv.Itoa(int(twitter_user.ID)),
		UserName:       twitter_user.ScreenName,
		Name:           twitter_user.Name,
		CreatedAt:      ct,
		FollowersCount: uint64(twitter_user.FollowersCount),
		FollowingCount: uint64(*&twitter_user.FriendsCount),
		TweetCount:     uint64(twitter_user.StatusesCount),
	}
	return user
}

func lookupTwitterUserV2(ctx context.Context, user_twitter *models.UserTwitterAuth) (*models.TwitterUser, error) {
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return nil, err
	}
	key := auth_config.key
	secret := auth_config.secret
	transport := &http.Transport{Proxy: setProxy()}
	client := &http.Client{Transport: transport}
	in := &gotwi.NewGotwiClientInput{
		HTTPClient:           client,
		AuthenticationMethod: gotwi.AuthenMethodOAuth2BearerToken,
		OAuthConsumerKey:     key,
		OAuthConsumerSecret:  secret,
	}
	twitter_client, err := gotwi.NewGotwiClient(in)
	if err != nil {
		return nil, err
	}
	params := &types.UserLookupIDParams{
		ID: user_twitter.TwitterUserId,
		UserFields: fields.UserFieldList{
			fields.UserFieldCreatedAt,
			fields.UserFieldPublicMetrics,
		},
	}
	tr, err := users.UserLookupID(ctx, twitter_client, params)
	if err != nil {
		return nil, err
	}
	return buildV2User(&tr.Data), nil
}

//v1 lookup
func lookupTwitterUserV1(ctx context.Context, user_twitter *models.UserTwitterAuth) (*models.TwitterUser, error) {
	user_id, _ := strconv.ParseInt(user_twitter.TwitterUserId, 10, 64)
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return nil, err
	}
	key := auth_config.key
	secret := auth_config.secret
	config := oauth1.NewConfig(key, secret)
	token := oauth1.NewToken("", "")
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitterV1.NewClient(httpClient)
	params := &twitterV1.UserLookupParams{
		UserID: []int64{user_id},
	}
	users, _, err := client.Users.Lookup(params)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.New("Lookup User is empty")
	}
	twitter_user := users[0]
	return buildV1User(&twitter_user), nil
}

//v1
func followTwitterUserV1(ctx context.Context, user_twitter *models.UserTwitterAuth, twitter_user_id string) error {
	if user_twitter == nil || user_twitter.TwitterUser == nil {
		return errors.New("Twitter User is null")
	}
	user_id, _ := strconv.ParseInt(twitter_user_id, 10, 64)
	auth_app_name := user_twitter.AuthAppName
	if auth_app_name == "" {
		auth_app_name = defaultAuthAppName
	}
	auth_config, err := getAuthConfigByName(auth_app_name)
	if err != nil {
		return err
	}
	key := auth_config.key
	secret := auth_config.secret
	config := oauth1.NewConfig(key, secret)
	token := oauth1.NewToken(user_twitter.OauthToken, user_twitter.OauthTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitterV1.NewClient(httpClient)
	params := &twitterV1.FriendshipCreateParams{
		UserID: user_id,
	}
	_, _, err = client.Friendships.Create(params)
	return err
}

func AccessToken(ctx context.Context, oauth_token, oauth_verifier string) (string, error) {

	api := fmt.Sprintf("%s?oauth_token=%s&oauth_verifier=%s", AccessTokenEndpoint, oauth_token, oauth_verifier)

	transport := &http.Transport{Proxy: setProxy()}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest("POST", api, nil)

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", errors.New(res.Status)
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	return string(body), nil
}

func CreateOAuthSignature(in *CreateOAuthSignatureInput) (*CreateOAuthSignatureOutput, error) {
	out := CreateOAuthSignatureOutput{
		OAuthSignatureMethod: OAuthSignatureMethodHMACSHA1,
		OAuthVersion:         OAuthVersion10,
	}
	nonce, err := generateOAthNonce()
	if err != nil {
		return nil, err
	}
	out.OAuthNonce = nonce

	ts := fmt.Sprintf("%d", time.Now().Unix())
	out.OAuthTimestamp = ts
	endpointBase := endpointBase(in.RawEndpoint)

	parameterString := createParameterString(nonce, ts, in)
	sigBase := createSignatureBase(in.HTTPMethod, endpointBase, parameterString)
	sig, err := calculateSignature(sigBase, in.SigningKey)
	if err != nil {
		return nil, err
	}
	out.OAuthSignature = sig

	return &out, nil
}

func generateOAthNonce() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	nonce := base64.StdEncoding.EncodeToString(key)
	symbols := []string{"+", "/", "="}
	for _, s := range symbols {
		nonce = strings.Replace(nonce, s, "", -1)
	}
	return nonce, nil
}

func endpointBase(e string) string {
	queryIdx := strings.Index(e, "?")
	if queryIdx < 0 {
		return e
	}

	return e[:queryIdx]
}

func (e Endpoint) String() string {
	return string(e)
}

func (e Endpoint) Detail() (*EndpointInfo, error) {
	d := EndpointInfo{
		Raw:                      e.String(),
		EncodedQueryParameterMap: map[string]string{},
	}

	queryIdx := strings.Index(e.String(), "?")
	if queryIdx < 0 {
		d.Base = string(e)
		return &d, nil
	}

	d.Base = e.String()[:queryIdx]
	queryPart := e.String()[queryIdx+1:]
	paramsPairs := strings.Split(queryPart, "&")
	for _, pp := range paramsPairs {
		keyValue := strings.Split(pp, "=")
		var err error
		v := ""
		if len(keyValue) == 2 {
			v, err = url.QueryUnescape(keyValue[1])
			if err != nil {
				return nil, err
			}
		}
		d.EncodedQueryParameterMap[keyValue[0]] = v
	}

	return &d, nil
}

func createParameterString(nonce, ts string, in *CreateOAuthSignatureInput) string {
	qv := url.Values{}
	for k, v := range in.ParameterMap {
		qv.Add(k, v)
	}

	qv.Add("oauth_consumer_key", in.OAuthConsumerKey)
	qv.Add("oauth_nonce", nonce)
	qv.Add("oauth_signature_method", OAuthSignatureMethodHMACSHA1)
	qv.Add("oauth_timestamp", ts)
	qv.Add("oauth_token", in.OAuthToken)
	qv.Add("oauth_version", OAuthVersion10)

	encoded := qv.Encode()
	encoded = regexp.MustCompile(`([^%])(\+)`).ReplaceAllString(encoded, "$1%20")
	return encoded
}

func createSignatureBase(method, endpointBase, parameterString string) string {
	return fmt.Sprintf(
		"%s&%s&%s",
		url.QueryEscape(strings.ToUpper(method)),
		url.QueryEscape(endpointBase),
		url.QueryEscape(parameterString),
	)
}

func calculateSignature(base, key string) (string, error) {
	b := []byte(key)
	h := hmac.New(sha1.New, b)
	_, err := io.WriteString(h, base)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
