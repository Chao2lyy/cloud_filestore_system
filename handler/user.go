package handler

import (
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	pwd_salt = "#890"
)

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := os.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}
	r.ParseForm()

	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("invalid parameter"))
		return
	}
	//3.用户密码加盐处理
	encPasswd := util.Sha1([]byte(pwd_salt + passwd))
	//4.存入数据库 tbl_user 表并返回结果
	isSuccess := dblayer.UserSignUp(username, encPasswd)

	if isSuccess {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}
}

// login
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.Form.Get("username")
	passwd := r.Form.Get("password")
	//3.用户密码加盐处理
	encPasswd := util.Sha1([]byte(pwd_salt + passwd))

	//校验
	pwdChecked := dblayer.UserSignIn(username, encPasswd)
	if !pwdChecked {
		w.Write([]byte("Failed"))
		return
	}
	//生成token
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}
	//重定向
	//w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			"http://" + r.Host + "/static/view/home.html",
			username,
			token,
		},
	}
	w.Write(resp.JsonToBytes())
}

// UserInfoHandler: 查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.Form.Get("username")
	//token := r.Form.Get("token")

	// isValidToken := IsTokenValid(token)
	// if !isValidToken {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	return
	// }

	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// 4. 组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JsonToBytes())
}

// GenToken: 生成用户 token
func GenToken(username string) string {
	//token(40位字符 mde5 后得到的32位字符再加上截取时间戳前8位）生成规则：md5(username+timestamp+tokenSalt)+timestamp[:8]

	ts := fmt.Sprintf("%x", time.Now().In(util.CstZone).Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// IsTokenValid: token 是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO:判断 token 的时效性，是否过期

	// TODO:从数据库表 tbl_user_token 查询 username 对应的 token 信息

	// TODO: 对比两个 token 是否一致

	return true
}
