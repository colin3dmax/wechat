package context

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/colin3dmax/wechat/util"
)

const (
	//AccessTokenURL 获取access_token的接口
  AccessTokenURL = "https://api.weixin.qq.com/cgi-bin/token"
  AccessTokenURL2 = "https://api.weixin.qq.com/sns/oauth2/access_token"
)

//ResAccessToken struct
type ResAccessToken struct {
	util.CommonError

	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
  RefreshToken string `json:"refresh_token"`
	OpenID string `json:"openid"`
	Scope string `json:"scope"`
	UnionId string `json:"unionid"`
}

//GetAccessTokenFunc 获取 access token 的函数签名
type GetAccessTokenFunc func(ctx *Context) (accessToken string, err error)

//SetAccessTokenLock 设置读写锁（一个appID一个读写锁）
func (ctx *Context) SetAccessTokenLock(l *sync.RWMutex) {
	ctx.accessTokenLock = l
}

//SetGetAccessTokenFunc 设置自定义获取accessToken的方式, 需要自己实现缓存
func (ctx *Context) SetGetAccessTokenFunc(f GetAccessTokenFunc) {
	ctx.accessTokenFunc = f
}

//GetAccessToken 获取access_token
func (ctx *Context) GetAccessToken() (accessToken string, err error) {
  ctx.accessTokenLock.Lock()
  defer ctx.accessTokenLock.Unlock()

  if ctx.accessTokenFunc != nil {
    return ctx.accessTokenFunc(ctx)
  }
  accessTokenCacheKey := fmt.Sprintf("access_token_%s", ctx.AppID)
  val := ctx.Cache.Get(accessTokenCacheKey)
  if val != nil {
    accessToken = val.(string)
    return
  }

  //从微信服务器获取
  var resAccessToken ResAccessToken
  resAccessToken, err = ctx.GetAccessTokenFromServer()
  if err != nil {
    return
  }

  accessToken = resAccessToken.AccessToken
  return
}

//GetAccessToken 获取access_token
func (ctx *Context) GetAccessTokenAndOpenId(code string) (accessToken ResAccessToken, err error) {
  ctx.accessTokenLock.Lock()
  defer ctx.accessTokenLock.Unlock()

  accessTokenCacheKey := fmt.Sprintf("access_token_openid_%s", ctx.AppID)
  val := ctx.Cache.Get(accessTokenCacheKey)
  if val != nil {
    accessToken = val.(ResAccessToken)
    return
  }

  //从微信服务器获取
  var resAccessToken ResAccessToken
  resAccessToken, err = ctx.GetAccessTokenAndOpenIdFromServer(code)
  if err != nil {
    return
  }

  accessToken = resAccessToken
  return
}

//GetAccessTokenFromServer 强制从微信服务器获取token
func (ctx *Context) GetAccessTokenFromServer() (resAccessToken ResAccessToken, err error) {
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", AccessTokenURL, ctx.AppID, ctx.AppSecret)
	var body []byte
	body, err = util.HTTPGet(url)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &resAccessToken)
	if err != nil {
		return
	}
	if resAccessToken.ErrMsg != "" {
		err = fmt.Errorf("get access_token error : errcode=%v , errormsg=%v", resAccessToken.ErrCode, resAccessToken.ErrMsg)
		return
	}

	accessTokenCacheKey := fmt.Sprintf("access_token_%s", ctx.AppID)
	expires := resAccessToken.ExpiresIn - 1500
	err = ctx.Cache.Set(accessTokenCacheKey, resAccessToken.AccessToken, time.Duration(expires)*time.Second)
	return
}


//GetAccessTokenFromServer 强制从微信服务器获取token
func (ctx *Context) GetAccessTokenAndOpenIdFromServer(code string) (resAccessToken ResAccessToken, err error) {
  url := fmt.Sprintf("%s?appid=%s&secret=%s&code=%s&grant_type=authorization_code", AccessTokenURL2, ctx.AppID, ctx.AppSecret,code)
  var body []byte
  body, err = util.HTTPGet(url)
  if err != nil {
    return
  }
  err = json.Unmarshal(body, &resAccessToken)
  if err != nil {
    return
  }
  if resAccessToken.ErrMsg != "" {
    err = fmt.Errorf("get access_token error : errcode=%v , errormsg=%v", resAccessToken.ErrCode, resAccessToken.ErrMsg)
    return
  }

  accessTokenCacheKey := fmt.Sprintf("access_token_openid_%s", ctx.AppID)
  expires := resAccessToken.ExpiresIn - 1500
  err = ctx.Cache.Set(accessTokenCacheKey, resAccessToken, time.Duration(expires)*time.Second)
  return
}
