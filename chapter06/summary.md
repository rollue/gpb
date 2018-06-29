# Chapter6. Exposing Data and Functionality through a RESTful Data Web Service API

- 패키지설치
- 핸들러들간 데이터 공유
- 여론 조사 생성
  - 읽기, 지우기 제외
- Responding
- etc

### dependencies install
- [go get github.com/nsqio/go-nsq](https://godoc.org/github.com/nsqio/go-nsq)
- [go get gopkg.in/mgo.v2](https://godoc.org/labix.org/v2/mgo)
   - go get github.com/globalsign/mgo 
- [go get github.com/garyburd/go-oauth/oauth](https://github.com/garyburd/go-oauth)
  - import github.com/**matryer**/go-oauth/oauth ->  import github.com/**garyburd**/go-oauth/oauth
- mux ?

### Sharing data between handlers
- 내장되어 있는(Go 1.7 버전부터 기본 라이브러리에 탑재) **Context Package**를 사용합니다
  - `WithValue(parent Context, key, val interface{})`함수를 사용해 request에 포함되어 있는 APIKey저장 후 handlePolls 호출합니다
    - APIKey가 서버에 정의되어 있는 값과 일치하지 않을 경우 **401** 리턴합니다
  -  `Value(key interface{})` 컨텍스트에 저장한 값을 **key** 이용해 찾을 수 있습니다

```go
type contextKey struct {
	name string
}

var contextKeyAPIKey = &contextKey{"api-key"}

// Get API Key Value
func APIKey(ctx context.Context) (string, bool) {
	key := ctx.Value(contextKeyAPIKey)
	if key == nil {
		return "", false
	}
	keystr, ok := key.(string)
	return keystr, ok
}

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if !isValidAPIKey(key) {
			respondErr(w, r, http.StatusUnauthorized, "invalid API key")
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyAPIKey, key)
		fn(w, r.WithContext(ctx))
	}
}

func isValidAPIKey(key string) bool {
	return key == "abc123"
}
```

### Creating a poll
유저는 여론 조사를 만들기 위해 /polls/에 POST 요청을 합니다

1. `main.go` -> `mux.HandleFunc("/polls/", withCORS(withAPIKey(s.handlePolls)))`
2. `polls.go` -> `handlePolls` **Method**  체크 GET, POST, DELETE 외 요청시 404 리턴
3. `polls.go` -> `handlePollsPost`

```go
// main.go
func main() {
	var (
		addr  = flag.String("addr", ":8080", "endpoint address")
		mongo = flag.String("mongo", "localhost", "mongodb address")
	)
	log.Println("Dialing mongo", *mongo)
	db, err := mgo.Dial(*mongo)
	if err != nil {
		log.Fatalln("failed to connect to mongo:", err)
	}
	defer db.Close()
	s := &Server{
		db: db,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/polls/", withCORS(withAPIKey(s.handlePolls)))
	log.Println("Starting web server on", *addr)
	http.ListenAndServe(":8080", mux)
	log.Println("Stopping...")
}

// Server is the API server
type Server struct {
	db *mgo.Session
}

/*
 * polls.go
 * main.go Server struct 포인터
 **/
func (s *Server) handlePolls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handlePollsGet(w, r)
		return
	case "POST":
		s.handlePollsPost(w, r)
		return
	case "DELETE":
		s.handlePollsDelete(w, r)
		return
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Methods", "DELETE")
		respond(w, r, http.StatusOK, nil)
		return
	}
	// not found
	respondHTTPErr(w, r, http.StatusNotFound)
}

func (s *Server) handlePollsPost(w http.ResponseWriter, r *http.Request) {
	/*
	 * return db session 
	 **/
	session := s.db.Copy()
	defer session.Close()
	/*
	 * study db -> polls collection
	 **/
	c := session.DB("study").C("polls")
	var p poll
	/*
	 * respond.go에 정의된 body parser
	 * decode 에러시 400 리턴
	 **/	
	if err := decodeBody(r, &p); err != nil {
		respondErr(w, r, http.StatusBadRequest, "failed to read poll from request", err)
		return
	}
	/*
	 * main.go -> Get API Key Value
	 * APIKey를 받은 후 p.APIKey assign
	 **/	
	apikey, ok := APIKey(r.Context())
	if ok {
		p.APIKey = apikey
	}
	p.ID = bson.NewObjectId()
	/*
	 * 여론 조사 생성 후 해당 여론 조사 방으로 redirect
	 **/
	if err := c.Insert(p); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "failed to insert poll", err)
		return
	}
	w.Header().Set("Location", "polls/"+p.ID.Hex())
	respond(w, r, http.StatusCreated, nil)
}
```

### Responding
> [decode, encode, error helper](https://github.com/mhoonjeon/gpb/blob/master/chapter06/respond.go)
```go
func decodeBody(r *http.Request, v interface{}) error { ... }
func encodeBody(w http.ResponseWriter, r *http.Request, v interface{}) error { ... }
func respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) { ... }
func respondErr(w http.ResponseWriter, r *http.Request, status int, args ...interface{} ) { ... }
func respondHTTPErr(w http.ResponseWriter, r *http.Request, status int) { ... }
```

### Cross-origin resource sharing
> Cross-Origin Resource Sharing 표준은 웹 브라우저가 사용하는 정보를 읽을 수 있도록 허가된 출처 집합를 서버에게 알려주도록 허용하는 HTTP 헤더를 추가함으로써 동작합니다.
[CORS 참조](https://developer.mozilla.org/ko/docs/Web/HTTP/Access_control_CORS)

```go
func main() {
        ...
        mux.HandleFunc("/polls/", withCORS(withAPIKey(s.handlePolls)))
        ...
}

func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		fn(w, r)
	}
}
```

#### etc
BSON(gopkg.in/mgo.v2/bson)을 사용하여 별도의 처리 없이 인코딩 및 디코딩에 사용되는 필드 이름을 지정할 수 있습니다
```go 
type poll struct {
        ID      bson.ObjectId  `bson:"_id" json:"id"`
	Title   string         `json:"title" bson:"title"`
	Options []string       `json:"options"`
	Results map[string]int `json:"results,omitempty"`
	APIKey  string         `json:"apikey"`
}
```

JSONP: Cross-origin policy 에 상관없이 데이터를 주고 받을 수 있다.
