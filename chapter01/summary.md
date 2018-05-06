
# 1. GO-LANG INTRO

## 1.1 간단한 GO 소개
- 컴파일 언어
- 정적 타입, 강 타입
- 가비지 컬렉션 제공
- 모듈화 및 패키지(go get/go instll)
- 언어 차원에서 동시성과 병렬성 지원(Goroutine)
- [GO 기초 학습](http://golang.site/)
- [Web-based Go 편집 및 테스트](http://play.golang.org)

## 1.2 간단한 웹서버 만들기
- func ListenAndServe(addr string, handler Handler) error : HTTP 연결을 받고 요청에 응답
- func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) : 경로별로 요청을 처리할 핸들러 함수를 등록
- func NewServeMux() *ServeMux : HTTP 요청 멀티플렉서 인스턴스 생성

- 핸들러 미지정
```go
package main
import (
"log"
"net/http"
)
func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
```

- 핸들러 지정
```go
package main
import (
"log"
"net/http"
)
func main() {
    mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world2"))
	})
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
```

-[참고자료](https://golang.org/pkg/net/http/)

## 1.3 간단한 채팅 서버 만들기
- 표준 라이브러리 임포트, 코어 장고 임포트, 장고와 무관한 외부 앱 임포트, 프로젝트 앱 임포트 순으로 임포트 문들을 구성한다.

### 1.3.1 Templates
- ServeHTTP 내부에서 한 번만 템플릿을 컴파일
```go
// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}
```


### 1.3.2 Channel
- 채널별로 이벤트 매핑
```go
func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// joining
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			// leaving
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", string(msg))
			// forward message to all clients
			for client := range r.clients {
				client.send <- msg
				r.tracer.Trace(" -- sent to client")
			}
		}
	}
}
```
## 1.4 Trace
- TDD기반 유용한 디버깅 기법
