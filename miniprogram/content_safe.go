package miniprogram

import (
  "encoding/json"
  "fmt"
  "github.com/colin3dmax/wechat/util"
  "io/ioutil"
  "net/http"
  "strconv"
  "time"
)

func (wxa *MiniProgram) postMultipartForm(urlStr string, fields []util.MultipartFormField) (response []byte, err error) {
  var accessToken string
  accessToken, err = wxa.GetAccessToken()
  if err != nil {
    return
  }
  urlStr = fmt.Sprintf(urlStr, accessToken)
  response, err = util.PostMultipartForm(fields,urlStr)
  return response,err
}

const (
  // 访问页面
  imgSecCheckURL = "https://api.weixin.qq.com/wxa/img_sec_check?access_token=%s"
  msgSecCheck = "https://api.weixin.qq.com/wxa/msg_sec_check?access_token=%s"
)

type FileInfo struct {
  FileName string
  FileType string
  Data []byte
}

func (wxa *MiniProgram) downloadFile(uri string) (*FileInfo, error) {
  response, err := http.Get(uri)
  if err != nil {
    return nil, err
  }

  fileInfo := FileInfo{
    FileName: strconv.FormatInt(time.Now().UnixNano(),10),
    FileType: "",
    Data:     nil,
  }

  defer response.Body.Close()
  if response.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("http get error : uri=%v , statusCode=%v", uri, response.StatusCode)
  }
  fmt.Println(response.Header)
  fileInfo.FileType = response.Header.Get("Content-Type")
  fileInfo.Data,err =ioutil.ReadAll(response.Body)
  if err!=nil {
    return nil,err
  }
  return &fileInfo,nil
}

// ResImgSecCheck 小程序访问分布数据返回
type ResImgSecCheck struct {
  util.CommonError
}

// ImgSecCheck 获取小程序页面访问数据
func (wxa *MiniProgram) ImgSecCheck(imgUrl string) (result ResImgSecCheck, err error) {
  fileInfo,err := wxa.downloadFile(imgUrl)
  if err!=nil {
    return
  }
  forms := []util.MultipartFormField{
    {
      IsFile:    false,
      Fieldname: "media",
      Value:    fileInfo.Data,
      Filename:  fileInfo.FileName,
    },
  }
  response, err := wxa.postMultipartForm(imgSecCheckURL, forms)
  if err != nil {
    return
  }
  err = json.Unmarshal(response, &result)
  if err != nil {
    return
  }
  if result.ErrCode != 0 {
    err = fmt.Errorf("ImgSecCheck error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
    return
  }
  return
}


// ImgSecCheck 获取小程序页面访问数据
func (wxa *MiniProgram) MsgSecCheck(msg string) (result ResImgSecCheck, err error) {

  body := map[string]string{
    "content": msg,
  }
  response, err := wxa.fetchData(msgSecCheck, body)
  if err != nil {
    return
  }
  err = json.Unmarshal(response, &result)
  if err != nil {
    return
  }
  if result.ErrCode != 0 {
    err = fmt.Errorf("MsgSecCheck error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
    return
  }
  return
}
