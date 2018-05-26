# Chapter3. Three Ways to Implement Profile picture
* 인증서버(OAuth2)에서 제공하는 사진 사용
* [gravatar](https://ko.gravatar.com/)  웹 서비스를 사용하여 사용자의 메일 주소로 사진 사용
* 유저가 자신의 사진을 업로드하고 사용
## Avatars from the OAuth2 server
- 인증 서비스마다 사진을 담은 url 필드명이 다르다. Gomniauth에서 공용 필드를 가져 오기 위한 Interface 지원 
  - 깃헙: avatar_url
  - 구글: picture
  - 페이스북: picture -> url
- auth.go, message.go, client.go url 필드 추가
```go
func loginHandler(w http.ResponseWriter, r *http.Request) {        
        ...
        creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
        user, err := provider.GetUser(creds)
        authCookieValue := objx.New(map[string]interface{}{
                "name":       user.Name(),
                "avatar_url":  user.AvatarURL(),
        }).MustBase64()
}

type message struct {
	...
	AvatarURL string
}

func (c *client) read() {
        ...
	if avatarURL, ok := c.userData["avatar_url"]; ok {
		msg.AvatarURL = avatarURL.(string)
	}
}
```

## The Gravatar implementation
- Gravatar의  메일 주소에서 MD5 해시를 생성하고 fmt.Sprintf를 사용하여 기본 URL과 함께 리턴
  - `//www.gravatar.com/avatar/1aedb8d9dc4751e229a335e371db8058`
- main.go, room.go, avatar.go, auth.go 변경
```go
func main() {
        ...
        r := newRoom(UseGravatar)
}

func newRoom(avatar GravatarAvatar) *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
                avatar:  avatar,
	}
}

func(GravatarAvatar) GetAvatarURL(c *client) (string, error) {
        if email, ok := c.userData["email"]; ok {
                if emailStr, ok := email.(string); ok {
                m := md5.New()
                io.WriteString(m, strings.ToLower(emailStr))
                return fmt.Sprintf("//www.gravatar.com/avatar/%x", m.Sum(nil)), nil
        } }
        return "", ErrNoAvatarURL
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	...
        authCookieValue := objx.New(map[string]interface{}{
              "name":       user.Name(),
              "avatar_url": user.AvatarURL(),
              "email":       user.Email(),
       }).MustBase64()
}
```

## Uploading an avatar picture
- upload.html 통해 이미지 전송
- uploaderHandler: avatarFile 받은 후 바이트를 읽을 수 있는 내장모듈 사용
  - [FormValue](https://golang.org/pkg/net/http/#Request.FormFile)
  - [ioutil](https://golang.org/pkg/io/ioutil/)
  - [path](https://golang.org/pkg/path/#Join)
```go
func uploaderHandler(w http.ResponseWriter, req *http.Request) {
	userID := req.FormValue("userid")
	file, header, err := req.FormFile("avatarFile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filename := path.Join("avatars", userID+path.Ext(header.Filename))
	err = ioutil.WriteFile(filename, data, 0777)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, "Successful")
}
``` 
- avatars dir 존재하지 않는경우 
  - `StatusInternalServerError`: 500

```go
if _, err := os.Stat("avatars"); os.IsNotExist(err) {
      os.Mkdir("avatars", os.ModePerm)
}
```
## Logging out
* 쿠키값 제거 후 `/chat` redirect
```go
http.HandleFunc("/logout", func(w http.ResponseWriter, r  *http.Request) {
     http.SetCookie(w, &http.Cookie{
       Name:   "auth",
       Value:  "",
       Path:   "/",
       MaxAge: -1,
     })
     w.Header().Set("Location", "/chat")
     w.WriteHeader(http.StatusTemporaryRedirect)
})
```
## 기타
- [dep](https://github.com/golang/dep)
- Go Live Reload Server
  - [gin](https://github.com/codegangsta/gin)
  - ??
