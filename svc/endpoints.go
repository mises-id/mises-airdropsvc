// Code generated by truss. DO NOT EDIT.
// Rerunning truss will overwrite this file.
// Version: 5f7d5bf015
// Version Date: 2021-11-26T09:27:01Z

package svc

// This file contains methods to make individual endpoints from services,
// request and response types to serve those endpoints, as well as encoders and
// decoders for those types, for all of our supported transport serialization
// formats.

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"

	pb "github.com/mises-id/mises-airdropsvc/proto"
)

// Endpoints collects all of the endpoints that compose an add service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
//
// In a server, it's useful for functions that need to operate on a per-endpoint
// basis. For example, you might pass an Endpoints to a function that produces
// an http.Handler, with each method (endpoint) wired up to a specific path. (It
// is probably a mistake in design to invoke the Service methods on the
// Endpoints struct in a server.)
//
// In a client, it's useful to collect individually constructed endpoints into a
// single type that implements the Service interface. For example, you might
// construct individual endpoints using transport/http.NewClient, combine them into an Endpoints, and return it to the caller as a Service.
type Endpoints struct {
	TestEndpoint              endpoint.Endpoint
	GetTwitterAuthUrlEndpoint endpoint.Endpoint
	GetAirdropInfoEndpoint    endpoint.Endpoint
	TwitterCallbackEndpoint   endpoint.Endpoint
	TwitterFollowEndpoint     endpoint.Endpoint
	LookupTwitterEndpoint     endpoint.Endpoint
	SendTweetEndpoint         endpoint.Endpoint
	LikeTweetEndpoint         endpoint.Endpoint
	ReplyTweetEndpoint        endpoint.Endpoint
	CheckTwitterUserEndpoint  endpoint.Endpoint
	ChannelInfoEndpoint       endpoint.Endpoint
	PageChannelUserEndpoint   endpoint.Endpoint
	GetChannelUserEndpoint    endpoint.Endpoint
	AirdropTwitterEndpoint    endpoint.Endpoint
	AirdropChannelEndpoint    endpoint.Endpoint
}

// Endpoints

func (e Endpoints) Test(ctx context.Context, in *pb.TestRequest) (*pb.TestResponse, error) {
	response, err := e.TestEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.TestResponse), nil
}

func (e Endpoints) GetTwitterAuthUrl(ctx context.Context, in *pb.GetTwitterAuthUrlRequest) (*pb.GetTwitterAuthUrlResponse, error) {
	response, err := e.GetTwitterAuthUrlEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.GetTwitterAuthUrlResponse), nil
}

func (e Endpoints) GetAirdropInfo(ctx context.Context, in *pb.GetAirdropInfoRequest) (*pb.GetAirdropInfoResponse, error) {
	response, err := e.GetAirdropInfoEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.GetAirdropInfoResponse), nil
}

func (e Endpoints) TwitterCallback(ctx context.Context, in *pb.TwitterCallbackRequest) (*pb.TwitterCallbackResponse, error) {
	response, err := e.TwitterCallbackEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.TwitterCallbackResponse), nil
}

func (e Endpoints) TwitterFollow(ctx context.Context, in *pb.TwitterFollowRequest) (*pb.TwitterFollowResponse, error) {
	response, err := e.TwitterFollowEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.TwitterFollowResponse), nil
}

func (e Endpoints) LookupTwitter(ctx context.Context, in *pb.LookupTwitterRequest) (*pb.LookupTwitterResponse, error) {
	response, err := e.LookupTwitterEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.LookupTwitterResponse), nil
}

func (e Endpoints) SendTweet(ctx context.Context, in *pb.SendTweetRequest) (*pb.SendTweetResponse, error) {
	response, err := e.SendTweetEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.SendTweetResponse), nil
}

func (e Endpoints) LikeTweet(ctx context.Context, in *pb.LikeTweetRequest) (*pb.LikeTweetResponse, error) {
	response, err := e.LikeTweetEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.LikeTweetResponse), nil
}

func (e Endpoints) ReplyTweet(ctx context.Context, in *pb.ReplyTweetRequest) (*pb.ReplyTweetResponse, error) {
	response, err := e.ReplyTweetEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.ReplyTweetResponse), nil
}

func (e Endpoints) CheckTwitterUser(ctx context.Context, in *pb.CheckTwitterUserRequest) (*pb.CheckTwitterUserResponse, error) {
	response, err := e.CheckTwitterUserEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.CheckTwitterUserResponse), nil
}

func (e Endpoints) ChannelInfo(ctx context.Context, in *pb.ChannelInfoRequest) (*pb.ChannelInfoResponse, error) {
	response, err := e.ChannelInfoEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.ChannelInfoResponse), nil
}

func (e Endpoints) PageChannelUser(ctx context.Context, in *pb.PageChannelUserRequest) (*pb.PageChannelUserResponse, error) {
	response, err := e.PageChannelUserEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.PageChannelUserResponse), nil
}

func (e Endpoints) GetChannelUser(ctx context.Context, in *pb.GetChannelUserRequest) (*pb.GetChannelUserResponse, error) {
	response, err := e.GetChannelUserEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.GetChannelUserResponse), nil
}

func (e Endpoints) AirdropTwitter(ctx context.Context, in *pb.AirdropTwitterRequest) (*pb.AirdropTwitterResponse, error) {
	response, err := e.AirdropTwitterEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.AirdropTwitterResponse), nil
}

func (e Endpoints) AirdropChannel(ctx context.Context, in *pb.AirdropChannelRequest) (*pb.AirdropChannelResponse, error) {
	response, err := e.AirdropChannelEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return response.(*pb.AirdropChannelResponse), nil
}

// Make Endpoints

func MakeTestEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.TestRequest)
		v, err := s.Test(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeGetTwitterAuthUrlEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.GetTwitterAuthUrlRequest)
		v, err := s.GetTwitterAuthUrl(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeGetAirdropInfoEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.GetAirdropInfoRequest)
		v, err := s.GetAirdropInfo(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeTwitterCallbackEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.TwitterCallbackRequest)
		v, err := s.TwitterCallback(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeTwitterFollowEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.TwitterFollowRequest)
		v, err := s.TwitterFollow(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeLookupTwitterEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.LookupTwitterRequest)
		v, err := s.LookupTwitter(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeSendTweetEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.SendTweetRequest)
		v, err := s.SendTweet(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeLikeTweetEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.LikeTweetRequest)
		v, err := s.LikeTweet(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeReplyTweetEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.ReplyTweetRequest)
		v, err := s.ReplyTweet(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeCheckTwitterUserEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.CheckTwitterUserRequest)
		v, err := s.CheckTwitterUser(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeChannelInfoEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.ChannelInfoRequest)
		v, err := s.ChannelInfo(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakePageChannelUserEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.PageChannelUserRequest)
		v, err := s.PageChannelUser(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeGetChannelUserEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.GetChannelUserRequest)
		v, err := s.GetChannelUser(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeAirdropTwitterEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.AirdropTwitterRequest)
		v, err := s.AirdropTwitter(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeAirdropChannelEndpoint(s pb.AirdropsvcServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*pb.AirdropChannelRequest)
		v, err := s.AirdropChannel(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

// WrapAllExcept wraps each Endpoint field of struct Endpoints with a
// go-kit/kit/endpoint.Middleware.
// Use this for applying a set of middlewares to every endpoint in the service.
// Optionally, endpoints can be passed in by name to be excluded from being wrapped.
// WrapAllExcept(middleware, "Status", "Ping")
func (e *Endpoints) WrapAllExcept(middleware endpoint.Middleware, excluded ...string) {
	included := map[string]struct{}{
		"Test":              {},
		"GetTwitterAuthUrl": {},
		"GetAirdropInfo":    {},
		"TwitterCallback":   {},
		"TwitterFollow":     {},
		"LookupTwitter":     {},
		"SendTweet":         {},
		"LikeTweet":         {},
		"ReplyTweet":        {},
		"CheckTwitterUser":  {},
		"ChannelInfo":       {},
		"PageChannelUser":   {},
		"GetChannelUser":    {},
		"AirdropTwitter":    {},
		"AirdropChannel":    {},
	}

	for _, ex := range excluded {
		if _, ok := included[ex]; !ok {
			panic(fmt.Sprintf("Excluded endpoint '%s' does not exist; see middlewares/endpoints.go", ex))
		}
		delete(included, ex)
	}

	for inc := range included {
		if inc == "Test" {
			e.TestEndpoint = middleware(e.TestEndpoint)
		}
		if inc == "GetTwitterAuthUrl" {
			e.GetTwitterAuthUrlEndpoint = middleware(e.GetTwitterAuthUrlEndpoint)
		}
		if inc == "GetAirdropInfo" {
			e.GetAirdropInfoEndpoint = middleware(e.GetAirdropInfoEndpoint)
		}
		if inc == "TwitterCallback" {
			e.TwitterCallbackEndpoint = middleware(e.TwitterCallbackEndpoint)
		}
		if inc == "TwitterFollow" {
			e.TwitterFollowEndpoint = middleware(e.TwitterFollowEndpoint)
		}
		if inc == "LookupTwitter" {
			e.LookupTwitterEndpoint = middleware(e.LookupTwitterEndpoint)
		}
		if inc == "SendTweet" {
			e.SendTweetEndpoint = middleware(e.SendTweetEndpoint)
		}
		if inc == "LikeTweet" {
			e.LikeTweetEndpoint = middleware(e.LikeTweetEndpoint)
		}
		if inc == "ReplyTweet" {
			e.ReplyTweetEndpoint = middleware(e.ReplyTweetEndpoint)
		}
		if inc == "CheckTwitterUser" {
			e.CheckTwitterUserEndpoint = middleware(e.CheckTwitterUserEndpoint)
		}
		if inc == "ChannelInfo" {
			e.ChannelInfoEndpoint = middleware(e.ChannelInfoEndpoint)
		}
		if inc == "PageChannelUser" {
			e.PageChannelUserEndpoint = middleware(e.PageChannelUserEndpoint)
		}
		if inc == "GetChannelUser" {
			e.GetChannelUserEndpoint = middleware(e.GetChannelUserEndpoint)
		}
		if inc == "AirdropTwitter" {
			e.AirdropTwitterEndpoint = middleware(e.AirdropTwitterEndpoint)
		}
		if inc == "AirdropChannel" {
			e.AirdropChannelEndpoint = middleware(e.AirdropChannelEndpoint)
		}
	}
}

// LabeledMiddleware will get passed the endpoint name when passed to
// WrapAllLabeledExcept, this can be used to write a generic metrics
// middleware which can send the endpoint name to the metrics collector.
type LabeledMiddleware func(string, endpoint.Endpoint) endpoint.Endpoint

// WrapAllLabeledExcept wraps each Endpoint field of struct Endpoints with a
// LabeledMiddleware, which will receive the name of the endpoint. See
// LabeldMiddleware. See method WrapAllExept for details on excluded
// functionality.
func (e *Endpoints) WrapAllLabeledExcept(middleware func(string, endpoint.Endpoint) endpoint.Endpoint, excluded ...string) {
	included := map[string]struct{}{
		"Test":              {},
		"GetTwitterAuthUrl": {},
		"GetAirdropInfo":    {},
		"TwitterCallback":   {},
		"TwitterFollow":     {},
		"LookupTwitter":     {},
		"SendTweet":         {},
		"LikeTweet":         {},
		"ReplyTweet":        {},
		"CheckTwitterUser":  {},
		"ChannelInfo":       {},
		"PageChannelUser":   {},
		"GetChannelUser":    {},
		"AirdropTwitter":    {},
		"AirdropChannel":    {},
	}

	for _, ex := range excluded {
		if _, ok := included[ex]; !ok {
			panic(fmt.Sprintf("Excluded endpoint '%s' does not exist; see middlewares/endpoints.go", ex))
		}
		delete(included, ex)
	}

	for inc := range included {
		if inc == "Test" {
			e.TestEndpoint = middleware("Test", e.TestEndpoint)
		}
		if inc == "GetTwitterAuthUrl" {
			e.GetTwitterAuthUrlEndpoint = middleware("GetTwitterAuthUrl", e.GetTwitterAuthUrlEndpoint)
		}
		if inc == "GetAirdropInfo" {
			e.GetAirdropInfoEndpoint = middleware("GetAirdropInfo", e.GetAirdropInfoEndpoint)
		}
		if inc == "TwitterCallback" {
			e.TwitterCallbackEndpoint = middleware("TwitterCallback", e.TwitterCallbackEndpoint)
		}
		if inc == "TwitterFollow" {
			e.TwitterFollowEndpoint = middleware("TwitterFollow", e.TwitterFollowEndpoint)
		}
		if inc == "LookupTwitter" {
			e.LookupTwitterEndpoint = middleware("LookupTwitter", e.LookupTwitterEndpoint)
		}
		if inc == "SendTweet" {
			e.SendTweetEndpoint = middleware("SendTweet", e.SendTweetEndpoint)
		}
		if inc == "LikeTweet" {
			e.LikeTweetEndpoint = middleware("LikeTweet", e.LikeTweetEndpoint)
		}
		if inc == "ReplyTweet" {
			e.ReplyTweetEndpoint = middleware("ReplyTweet", e.ReplyTweetEndpoint)
		}
		if inc == "CheckTwitterUser" {
			e.CheckTwitterUserEndpoint = middleware("CheckTwitterUser", e.CheckTwitterUserEndpoint)
		}
		if inc == "ChannelInfo" {
			e.ChannelInfoEndpoint = middleware("ChannelInfo", e.ChannelInfoEndpoint)
		}
		if inc == "PageChannelUser" {
			e.PageChannelUserEndpoint = middleware("PageChannelUser", e.PageChannelUserEndpoint)
		}
		if inc == "GetChannelUser" {
			e.GetChannelUserEndpoint = middleware("GetChannelUser", e.GetChannelUserEndpoint)
		}
		if inc == "AirdropTwitter" {
			e.AirdropTwitterEndpoint = middleware("AirdropTwitter", e.AirdropTwitterEndpoint)
		}
		if inc == "AirdropChannel" {
			e.AirdropChannelEndpoint = middleware("AirdropChannel", e.AirdropChannelEndpoint)
		}
	}
}
