package wxext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// 文档地址：https://www.wxext.cn/home/developer.html

const (
	TextMessage                  = 1   // 文本消息
	ImgMessage                   = 3   // 图片消息
	VoiceMessage                 = 34  // 语音消息
	FriendMessage                = 37  // 好友确认消息
	VideoMessage                 = 43  // 视频消息
	EmotionMessage               = 47  // 表情消息
	LocationMessage              = 48  // 位置消息
	LinkShareMessage             = 49  // 分享链接
	GroupMemberInfoUpdateMessage = 701 // 群成员信息更新
	GroupMemberIncreaseMessage   = 702 // 群成员增加
	GroupMemberDecreaseMessage   = 703 // 群成员减少
	FriendInfoUpdateMessage      = 704 // 联系人信息更新
	ReceivePaymentMessage        = 705 // 收款消息
	FriendAuthMessage            = 706 // 好友验证消息
	CreateGroupMessage           = 707 // 创建群聊消息
	XMLImgPathMessage            = 708 // xml照片地址信息
	Auth                         = 720
	Open                         = 721
	QrChange                     = 723
	Login                        = 724
	Logout                       = 725
	Expired                      = 729
	VoiceCallMessage             = 726 // 发起语音通话
	VoiceCallRejectMessage       = 727 // 拒绝语音通话
	VoiceCallAcceptMessage       = 728 // 接受语音通话
	ExtDisconnectMessage         = 802 // 插件连接断开通知
	WechatDisconnectMessage      = 803 // 微信连接断开通知
	HttpMessagePushErrorMessage  = 804 // http推送状态连续出错超过20次被停止推送
	Tips                         = 810 // 系统提示点击确定通知
)

type IWxext interface {
	Run(pid int32)
	Show(pid int32)
	ClickLogin(pid int32)
	GotoQr(pid int32)
	List()
	GetPids()
	GetInfo(pid int32)
	GetUser(pid int32)
	GetGroup(pid int32)
	GetGroupUser(pid int32, groupWxid string)
	GetUserImg(pid int32, wxid string)
	GetUserInfo(pid int32, wxid string)
	SendText(pid int32, wxid string, atid string, msg string)
	SendFileByUrl(pid int32, wxid string, fileUrl string)
	SendImage(pid int32, wxid string, imgPath string)
	SendImageByUrl(pid int32, wxid string, imgUrl string)
	SendEmojiForward(pid int32, wxid string, xml string)
	SendAppmsgForward(pid int32, wxid string, xml string)
	SendCard(pid int32, wxid, xml string)
	SetGroupAnnouncement(pid int32, groupWxid string, msg string)
	SetRemark(pid int32, groupWxid string, alias string)
	AddGroupMember(pid int32, groupWxid string, wxid string)
	DeleteGroupMember(pid int32, groupWxid string, wxid string)
	SetGroupName(pid int32, groupWxid string, groupName string)
	QuitGroup(pid int32, groupWxid string)
	SendGroupInvite(pid int32, groupWxid string, wxid string)
	CallVoipAudio(wxid string)
	ClearMsgList(pid int32)
	ExtList()
	GetDbName(pid int32)
	RunSql(pid int32, sql string)
	LoginOut(pid int32)
	Kill(pid int32)
	AgreeUser(encryptusername, ticket string)
	AgreeCash(pid int32, wxid, transferid string)
	GetFile(path string)
	GetMAC()
	DeQR(t, data string)
	CreateRoom(pid int32, msg string)
	AddUser(pid int32, wxid, msg string)
	DeleteUser(pid int32, wxid string)
	NetUpdateUser(pid int32, wxid string)
	GetToken(pid int32, token string)
	InstallExt(extName string)
	Ext(pid int32, action, extName string)
	DelExt(extName string)
}

type Wxext struct {
	Name          string
	Key           string
	Addr          string
	Port          uint16
	WebsocketPort uint16
	Send          chan<- map[string]interface{}
	Recv          <-chan map[string]interface{}
	ErrChan       <-chan bool
}

type Option func(wxext *Wxext)

func SetAddr(addr string) Option {
	return func(wxext *Wxext) {
		wxext.Addr = addr
	}
}

func SetPort(port uint16) Option {
	return func(wxext *Wxext) {
		wxext.Port = port
	}
}

func SetWebsocketPort(websocketPort uint16) Option {
	return func(wxext *Wxext) {
		wxext.WebsocketPort = websocketPort
	}
}

func NewWxext(Name, Key string, opts ...Option) *Wxext {
	wxext := &Wxext{
		Name:          Name,
		Key:           Key,
		Addr:          "127.0.0.1",
		Port:          8203,
		WebsocketPort: 8202,
	}
	for _, opt := range opts {
		opt(wxext)
	}
	return wxext
}

func (w *Wxext) Conn() error {
	ws := newWS(w.Addr, w.WebsocketPort, w.Name, w.Key)
	send, recv, errChan, err := ws.Conn()
	if err != nil {
		return err
	}

	w.Send = send
	w.Recv = recv
	w.ErrChan = errChan
	return nil
}

func (w Wxext) requests(req map[string]interface{}) (map[string]interface{}, error) {
	jsonStr, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	r, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/api?json&key=%s", w.Addr, w.Port, w.Key), bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var respMap map[string]interface{}
	err = json.Unmarshal(rb, &respMap)
	return respMap, nil
}

// Run 启动一个微信
// 0取当前第一个微信，-1新启动一个微信
func (w Wxext) Run(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "run",
		"pid":    pid,
	}
}

// Show 抖动微信
// 将微信抖动置顶
func (w Wxext) Show(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "show",
		"pid":    pid,
	}
}

// ClickLogin 点击登录
// 点击微信启动时的登录页面
func (w Wxext) ClickLogin(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "clickLogin",
		"pid":    pid,
	}
}

// GotoQr 跳转二维码
func (w Wxext) GotoQr(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "gotoQr",
		"pid":    pid,
	}
}

// LoginOut 退出登录
func (w Wxext) LoginOut(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "loginOut",
		"pid":    pid,
	}
}

// Kill 结束进程
func (w Wxext) Kill(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "kill",
		"pid":    pid,
	}
}

// List 连接列表
func (w Wxext) List() {
	w.Send <- map[string]interface{}{
		"method": "list",
		"pid":    0,
	}
}

// GetPids 获取进程列表
func (w Wxext) GetPids() {
	w.Send <- map[string]interface{}{
		"method": "pids",
	}
}

// ExtList 插件列表
func (w Wxext) ExtList() {
	w.Send <- map[string]interface{}{
		"method": "ext",
		"action": "list",
	}
}

// GetInfo 获取登录信息
func (w Wxext) GetInfo(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "getInfo",
		"pid":    pid,
	}
}

// GetUser 获取通讯录
func (w Wxext) GetUser(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "getUser",
		"pid":    pid,
	}
}

// GetGroup 获取群列表
func (w Wxext) GetGroup(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "getGroup",
		"pid":    pid,
	}
}

// GetGroupUser 获取群成员
// 群内昵称，列表第一个为群主
func (w Wxext) GetGroupUser(pid int32, groupWxid string) {
	w.Send <- map[string]interface{}{
		"method": "getGroupUser",
		"wxid":   groupWxid,
		"pid":    pid,
	}
}

// SetGroupAnnouncement 设置群公告
func (w Wxext) SetGroupAnnouncement(pid int32, groupWxid string, msg string) {
	w.Send <- map[string]interface{}{
		"method": "setRoomAnnouncement",
		"wxid":   groupWxid,
		"msg":    msg,
		"pid":    pid,
	}
}

// SetRemark 设置备注
func (w Wxext) SetRemark(pid int32, groupWxid string, alias string) {
	w.Send <- map[string]interface{}{
		"method": "setRemark",
		"wxid":   groupWxid,
		"msg":    alias,
		"pid":    pid,
	}
}

// DeleteGroupMember 群踢人
func (w Wxext) DeleteGroupMember(pid int32, groupWxid string, wxid string) {
	w.Send <- map[string]interface{}{
		"method": "delRoomMember",
		"wxid":   groupWxid,
		"msg":    wxid,
		"pid":    pid,
	}
}

// AddGroupMember 群拉人
func (w Wxext) AddGroupMember(pid int32, groupWxid string, wxid string) {
	w.Send <- map[string]interface{}{
		"method": "addGroupMember",
		"wxid":   groupWxid,
		"msg":    wxid,
		"pid":    pid,
	}
}

// SetGroupName 设置群名称
func (w Wxext) SetGroupName(pid int32, groupWxid string, groupName string) {
	w.Send <- map[string]interface{}{
		"method": "setRoomName",
		"wxid":   groupWxid,
		"msg":    groupName,
		"pid":    pid,
	}
}

// QuitGroup 退出群聊
func (w Wxext) QuitGroup(pid int32, groupWxid string) {
	w.Send <- map[string]interface{}{
		"method": "quitChatRoom",
		"wxid":   groupWxid,
		"pid":    pid,
	}
}

// SendGroupInvite 群邀请发送
func (w Wxext) SendGroupInvite(pid int32, groupWxid string, wxid string) {
	w.Send <- map[string]interface{}{
		"method": "sendGroupInvite",
		"wxid":   groupWxid,
		"msg":    wxid,
		"pid":    pid,
	}
}

// GetUserImg 查头像
func (w Wxext) GetUserImg(pid int32, wxid string) {
	w.Send <- map[string]interface{}{
		"method": "getUserImg",
		"wxid":   wxid,
		"pid":    pid,
	}
}

// GetUserInfo 查昵称信息
func (w Wxext) GetUserInfo(pid int32, wxid string) {
	w.Send <- map[string]interface{}{
		"method": "getUserInfo",
		"wxid":   wxid,
		"pid":    pid,
	}
}

// GetDbName 获取数据库列表
func (w Wxext) GetDbName(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "getDbName",
		"pid":    pid,
	}
}

// RunSql 执行数据库语句
func (w Wxext) RunSql(pid int32, sql string) {
	w.Send <- map[string]interface{}{
		"method": "runSql",
		"db":     "MicroMsg.db",
		"sql":    sql,
		"pid":    pid,
	}
}

// SendText 发送文本消息
// 需要艾特人时，传入atid
// 多人艾特用|分割，同时msg要对应多个艾特文本
func (w Wxext) SendText(pid int32, wxid string, atid string, msg string) {
	w.Send <- map[string]interface{}{
		"method": "sendText",
		"wxid":   wxid,
		"msg":    msg,
		"atid":   atid,
		"pid":    pid,
	}
}

// SendFileByUrl 发送文件(通过链接)
func (w Wxext) SendFileByUrl(pid int32, wxid string, fileUrl string) {
	w.Send <- map[string]interface{}{
		"method":   "sendFile",
		"wxid":     wxid,
		"file":     fileUrl,
		"fileType": "url",
		"pid":      pid,
	}
}

// SendImage 转发图片
func (w Wxext) SendImage(pid int32, wxid string, imgPath string) {
	w.Send <- map[string]interface{}{
		"method":  "sendImage",
		"wxid":    wxid,
		"img":     imgPath, // 图片本地路径
		"imgType": "file",
		"pid":     pid,
	}
}

// SendImageByUrl 转发图片(图片链接)
func (w Wxext) SendImageByUrl(pid int32, wxid string, imgUrl string) {
	w.Send <- map[string]interface{}{
		"method":  "sendImage",
		"wxid":    wxid,
		"img":     imgUrl, // 图片本地路径
		"imgType": "url",
		"pid":     pid,
	}
}

// SendEmojiForward 转发动态表情
func (w Wxext) SendEmojiForward(pid int32, wxid string, xml string) {
	w.Send <- map[string]interface{}{
		"method": "sendEmojiForward",
		"wxid":   wxid,
		"xml":    xml, // type=47收到的msg中的xml
		"pid":    pid,
	}
}

// SendAppmsgForward 转发文章或小程序
func (w Wxext) SendAppmsgForward(pid int32, wxid string, xml string) {
	w.Send <- map[string]interface{}{
		"method": "sendAppmsgForward",
		"wxid":   wxid,
		"xml":    xml, // type=49收到的msg中的xml
		"pid":    pid,
	}
}

// SendCard 发名片
func (w Wxext) SendCard(pid int32, wxid, xml string) {
	w.Send <- map[string]interface{}{
		"method": "sendCard",
		"wxid":   wxid,
		"xml":    xml,
		"pid":    pid,
	}
}

// CallVoipAudio 发送语音通话
func (w Wxext) CallVoipAudio(wxid string) {
	w.Send <- map[string]interface{}{
		"method": "callVoipAudio",
		"wxid":   wxid,
	}
}

// AgreeUser 同意好友
func (w Wxext) AgreeUser(encryptusername, ticket string) {
	w.Send <- map[string]interface{}{
		"method":          "agreeUser",
		"encryptusername": encryptusername,
		"ticket":          ticket,
		"scene":           14,
	}
}

// ClearMsgList 清空聊天记录
func (w Wxext) ClearMsgList(pid int32) {
	w.Send <- map[string]interface{}{
		"method": "ClearMsgList",
		"pid":    pid,
	}
}

// AgreeCash 接收转账收款
func (w Wxext) AgreeCash(pid int32, wxid, transferid string) {
	w.Send <- map[string]interface{}{
		"method":     "agreeCash",
		"wxid":       wxid,
		"transferid": transferid,
		"pid":        pid,
	}
}

// GetFile 获取文件数据
// map[data:map[flag:03f780e1b9bf8c02e31b7a26aad3cd4a fromid:xxxx@chatroom memid:wxid_xxxxx path:C:\Users\Administrator\Documents\WeChat Files\wxid_xxxx\FileStorage\File\2021-04\073f67298166f2ad83cd2e71f9240fd4.gif source:<msgsource>
//	<silence>1</silence>
//	<membercount>4</membercount>
//</msgsource>
//] method:xmlinfo myid:wxid_xxxxxx pid:51852 type:708]
func (w Wxext) GetFile(path string) {
	w.Send <- map[string]interface{}{
		"method": "getfile",
		"path":   path,
	}
}

// GetMAC 获取机器码
func (w Wxext) GetMAC() {
	w.Send <- map[string]interface{}{
		"method": "mac",
	}
}

// DeQR 识别二维码
func (w Wxext) DeQR(t, data string) {
	w.Send <- map[string]interface{}{
		"method": "deqr",
		"type":   t,
		"data":   data,
	}
}

// CreateRoom 创建群组
func (w Wxext) CreateRoom(pid int32, msg string) {
	w.Send <- map[string]interface{}{
		"method": "createRoom",
		"msg":    msg,
		"pid":    pid,
	}
}

func (w Wxext) AddUser(pid int32, wxid, msg string) {
	w.Send <- map[string]interface{}{
		"method": "addUser",
		"wxid":   wxid,
		"msg":    msg,
		"pid":    pid,
	}
}

// DeleteUser 删除好友
func (w Wxext) DeleteUser(pid int32, wxid string) {
	w.Send <- map[string]interface{}{
		"method": "deleteUser",
		"wxid":   wxid,
		"pid":    pid,
	}
}

// NetUpdateUser 更新好友
func (w Wxext) NetUpdateUser(pid int32, wxid string) {
	w.Send <- map[string]interface{}{
		"method": "netUpdateUser",
		"wxid":   wxid,
		"pid":    pid,
	}
}

// GetToken 查询token数据
func (w Wxext) GetToken(pid int32, token string) {
	w.Send <- map[string]interface{}{
		"method": "token",
		"token":  token,
		"pid":    pid,
	}
}

// InstallExt 安装插件
func (w Wxext) InstallExt(extName string) {
	w.Send <- map[string]interface{}{
		"method": "ext",
		"action": "install",
		"name":   extName,
	}
}

// Ext 插件操作
func (w Wxext) Ext(pid int32, action, extName string) {
	w.Send <- map[string]interface{}{
		"method": "ext",
		"action": action,
		"name":   extName,
		"pid":    pid,
	}
}

// DelExt 删除插件
func (w Wxext) DelExt(extName string) {
	w.Send <- map[string]interface{}{
		"method": "ext",
		"action": "del",
		"name":   extName,
	}
}
