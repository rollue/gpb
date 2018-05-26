# Chapter3. Three Ways to Implement Profile picture
* 소셜서비스(OAuth2)에서 제공하는 사진 사용
* [gravatar](https://ko.gravatar.com/)  웹 서비스를 사용하여 사용자의 전자 메일 주소로 사진 사용
* 유저가 자신의 사진을 업로드하고 사용
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
  - StatusInternalServerError: 500

```go
if _, err := os.Stat("avatars"); os.IsNotExist(err) {
      os.Mkdir("avatars", os.ModePerm)
}
```
## Logging out
* 쿠키값 제거 후 `/chat` 경로로 redirect
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

## Test

## 기타
