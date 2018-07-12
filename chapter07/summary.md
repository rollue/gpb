# Chapter7. Random Recommendations Web Service
## Summary
1. Google Places API를 활용한 API endpoint 설계
2. Agile 개발:
	1. 먼저 user story를 도출해 낸 후, 이를 위한 개발 시작 
    2. agile의 핵심: module간의 dependency를 줄여서 `회귀오류`를 줄이는 것
3. Frontend, Backend가 만나는 API endpoint를 도출해 병렬적 작업 가능케 한다.
4. 초기에는 코드에 데이터를 하드 코딩했다
	1. 개발시작 단계에서 DB 조사/학습에 대한 부담을 덜기 위해
	2. 데이터가 어떻게 접근될지만 잘 정의한다면(=API endpoint), DB 저장 여부 등은 향후에 수정해도 코드작동에 문제가 없다.
5. `Facade` 인터페이스를 통해, 내부에서 작동하는 struct 형태의 데이터를 다 노출시키지 않아도 된다. -> 필요한 정보만 노출시키는 public representation을 구현한다.
6. Go에서 공식적으로 지원하지는 않지만, `Enum`을 구현한다
	1. 순서대로  increment하는 경우 `iota`를 사용할 수 있다. [GO const와 iota 참고 블로그](http://brownbears.tistory.com/295)를 참고하자.
	2. Enum을 사용할 때 의미없는 숫자를 `String` 메쏘드를 통해 변환해 로그 등을 남기자.
7. TDD를 통해 코드를 개발하자 (red/green programming)
	1. 코드 구현없이 원하는 결과를 미리 예상해서 테스트 코드를 먼저 짠 다음, (=red)
	2. 실제 코드를 구현해서 테스트를 통과하도록 한다. (=green)

## Agile 개발
* 사용자의  Story
	* what
	* why

## JSON endpoints 설계
`GET/journeys`에서 가져오는 list를 바탕으로 Google API에 추천을 GET

```
# GET /recommendations?lat=1&lng=2&journey=bar|cafe&radius=10&cost=$...$$$$$
```

## Representing data in code
* 내부에 struct 형태로 구현체를 바로 저장
	* 장점: 외부 DB dependency없이 빠르게 개발 가능
```
// Journey의 struct type
type j struct {
	Name       string
	PlaceTypes []string
}
...

// data(journey)를 그대로 리턴
// struct 형태가 그대로 외부 API에 노출된다! -> Interface abstraction(추상화) 필요
func respond(w http.ResponseWriter, r *http.Request, data []interface{})
error {
	return json.NewEncoder(w).Encode(data)
}	
```


```
[{Name: "Romantic",	PlaceTypes: [		"park",		"bar",		"movie_theater",		"restaurant",		"florist",		"taxi_stand"	]}, ...]
```


## Public Views of Go structs
* interface가 어떻게 외부로 노출되야하는 지 정의한다.
```
// public.go
package meander

// Public 함수 유무를 체크한다.
type Facade interface {
	Public() interface{}
}
func Public(o interface{}) interface{} {
	if p, ok := o.(Facade); ok {
	return p.Public()
	}
	return o
}

...
...

// journeys.go
// 외부로 노출될 데이터 형태를 변환
// 이후 main.go에서 respond 함수에서 Public() 사용을 반영해 변경
func (j j) Public() interface{} {
	return map[string]interface{}{
		"name": j.Name,
		"journey": strings.Join(j.PlaceTypes, "|"),
	}
}
```
* json 표현을 위해 tag를 사용해도 되지만, Public() method를 사용하는 방법이 더 expressive하고 clear할 수 있다.


## Generating random recommendations
```
// query.go
package meander
type Place struct {
	*googleGeometry `json:"geometry"`
	Name string `json:"name"`
	Icon string `json:"icon"`
	Photos []*googlePhoto `json:"photos"`
	Vicinity string `json:"vicinity"`
}
...
type googleGeometry struct {
	*googleLocation `json:"location"`
}
...
type googleLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
```
* 위와 같은 nested data 형태(Place -> googleGeometry -> googleLocation)라도 API를 통해 flat하게 접근할 수 있다.
```
func (p *Place) Public() interface{} {
	return map[string]interface{}{
		"name": p.Name,
		"icon": p.Icon,
		"photos": p.Photos,
		"vicinity": p.Vicinity,
		"lat": p.Lat,
		"lng": p.Lng,
	}
}
```


## Enumerators in Go
* 공식적으로는 지원되지 않는다
* Workaround  [GO const와 iota 참고 블로그](http://brownbears.tistory.com/295)
```
// cost_level.go
package meander
type Cost int8
const (
	_ Cost = iota  // 0부터 시작하기에 dummy로 사용
	Cost1
	Cost2
	Cost3
	Cost4
	Cost5
)
```
* Cons type에 String() 메쏘드를 활용하여 단순한 숫자가 아닌, 의미를 담는 string representation으로 표현해준다.


## Test-driven enumerator
```
// cost_level_test.go
package meander_test
import (
	"testing"
	"github.com/cheekybits/is"
	"path/to/meander"
)
func TestCostValues(t *testing.T) {
	is := is.New(t)
	is.Equal(int(meander.Cost1), 1)
	is.Equal(int(meander.Cost2), 2)
	is.Equal(int(meander.Cost3), 3)
	is.Equal(int(meander.Cost4), 4)
	is.Equal(int(meander.Cost5), 5)
}
```
* 참고
`package meander_test`: meander 패키지가 아닌 이유는 유저의 입장에서 test 하기 위해서이다. meander 패키지와 분리되어, meander 패키지 내부를 접속하지 못하고 오직 Public하게 노출된 데이터 타입 및 함수만 접근할 수 있다. 실제 유저의 케이스와 동일한 상태로 `true user test`라고 할 수 있다. 물론 내부에 접근이 필요한 테스트의 경우에는,  `package meander`로 하면 된다. 


## Querying the Google Places API
* `func (q *Query) find(....)` 구현
* 특이사항 없음

## Building recommendations
* 구현된 `func (q *Query) find(....)`를 concurrent하게 처리한 후 결과를 반환한다.
```
// Concurrent하게 query를 처리해서 결과를 반환
func (q *Query) Run() []interface{} {
	rand.Seed(time.Now().UnixNano())  // random 시드
	var w sync.WaitGroup  // Go루틴을 끝날때까지 받아주는 대기 그룹
	var l sync.Mutex // protects places -> mutext는 자료(여기서는 map자료 구조인 places)에 대한 접근을 "한번에 하나씩" 제어하기 위해서 사용한다. 
	places := make([]interface{}, len(q.Journey))
	
	// Journey slice(bar, cafe 등)을 loop 돌다가
	for i, r := range q.Journey {
		w.Add(1)  // waitgroup에 1을 추가하고 고루틴 실행
		
		go func(types string, i int) {
			defer w.Done()  // WaitGroup 객체에 request가 완료되었음을 알리고 find()요청을 보냄
			response, err := q.find(types)
			if err != nil {
				log.Println("Failed to find places:", err)
				return
			}
			if len(response.Results) == 0 {
				log.Println("No places found for", types)
				return
			}
			for _, result := range response.Results {
				for _, photo := range result.Photos {
					photo.URL = "https://maps.googleapis.com/maps/api/place/photo?" +
						"maxwidth=1000&photoreference=" + photo.PhotoRef + "&key=" + APIKey
				}
			}
			randI := rand.Intn(len(response.Results))
			l.Lock()  // Mutext.lock과 unlock을 이용해서, 한번에 하나의 고루틴만 접근할 수 있도록 한다. 중복연산 방지용
			places[i] = response.Results[randI]
			l.Unlock()
		}(r, i)
	}
	w.Wait() // wait for everything to finish
	return places
}
...
```


## Handlers that use query parameters
*  url에서 query params을 사용하도록 http.HandleFunc 수정
* query objects 의 key값에 대입해준다.


## CORS
* Access-Control-Allow_Origin response 헤더를 *로 설정해주면 된다. API endpoint를 노출하는 `cmd/meander`내에 설정하면 된다.  
```
// main.go
func cors(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		f(w, r)
	}
}

...
func main() {
	...
	...
	http.HandleFunc("/recommendations", cors(func(w http.ResponseWriter, r *http.Request) {
		q := &meander.Query{
			Journey: strings.Split(r.URL.Query().Get("journey"), "|"),
		}
	...
	...
```
