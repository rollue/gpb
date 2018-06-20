# Chap5. Building Distributed Systems and Working with Flexible Data

* 스터디 진행자 : 배주웅
* 스터디 진행일 : 2018/06/??



## 목표

### 원래 목표

## 이 챕터의 원래 목표

* NoSQL 데이터베이스, 특히 MongoDB 알기

* 분산된 메세지 큐 시스템 배우기 (go-nsq 사용)

* 트위터 streaming API를 이용하여 정보 긁어오기



### 문제점

* streaming api business로 넘어감
* 사실 고루틴하고 채널 어떻게 쓰는지 모르겠음

### 바꾼 목표

* slack webhook api를 이용하여 정보 긁어오기
* 중간에 다른 서버를 끼고, dockerize하여 deploy하기 (도커간 통신은 nsq)
* data integrity issue 해결



**위 목표를 주안점으로 두고 개발하여 좋은 코드가 아닐 수 있습니다 ㅜ**



### Server structure

![](./assets/struct.jpg)



## 개발 내용

### Slack RTM API Receiver

`github.com/nlopes/slack` 라이브러리를 이용하였음 ㅎㅎ

```go
package main

import (
	"fmt"

	"github.com/nlopes/slack"
)

func main() {
	api := slack.New("slack-bot-token")

	messages := make(chan string, 10)

	rtm := api.NewRTM()
	go rtm.ManageConnection()
	go messager(messages)

	for msg := range rtm.IncomingEvents {

		switch msg.Data.(type) {
		case *slack.MessageEvent:
			obj := msg.Data.(*slack.MessageEvent)
             fmt.Printf("message: %s\n", obj.Text)

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:

		}
	}
}
```

받아오는 건 그렇게 중요한 것이 아니니까 딱히 다루지 않겠습니다. 이거 복사하면 됨



### Load Balancer

사실 처리가 오래 걸리는 시스템은 아니기 때문에 단순히 1/n로 분산 처리할 수 있게 하려 했지만, 그러면 실제 코드량도 적어질 뿐더러 배우는 게 없을 것 같아서 현재 몇 개의 스택이 쌓여 있는지 보고 한 곳에 과부화 되지 않게 처리하기로 함. (과부하를 구현하기 위해 message sender에서 고의적으로 1~3초 사이 랜덤하게 지연하게 구현하였음)

굳이 message queue로 구현하진 않았고, 대신 channel을 data queue로 구현하려고 마음먹었으나 예상대로 작동하지 않음.

일단 1개의 리퀘스트마다 하나하나 분배해주는 로직을 짜고 추후 수정하기로 생각



```go
func balancer(messages chan string) {
	//Receive Message From Slack RTM API
	chans := []chan string{
		make(chan string),
		make(chan string),
		make(chan string),
		make(chan string),
	}

	for i := 0; i < 4; i++ {
		go messenger(i, chans[i])
	}

    current := 0

	for text := range messages {
		current = (current + 1) % 4

		chans[current] <- text
	}
}
```



```go
func messenger(id int, value chan string) {
	for text := range value {
		message := &Message{
			Text:    text,
			Channel: id,
		}

		message.Send()

		time.Sleep(time.Duration(rand.Int()%3000+1000) * time.Millisecond)
	}
}
```



사실 후술할 RabbitMQ에서 다 지원하는 기능이긴 하지만, 그냥 짜 보았다.

![P-Exchange-Queue-C 여러개 그림](https://camo.githubusercontent.com/e0cb0f73a6d2041d2220710bb6c6551c0f0b9de2/687474703a2f2f676a63686f692e6769746875622e696f2f696d616765732f7261626269746d71312f70726566657463682d636f756e742e706e67)

다음에 시간 여유가 난다면 한번쯤 다시 다뤄보는것도 재미있겠지만, 스터디 주제랑 맞지 않는 것 같아서 일단 이거 사용은 안했슴다



## NSQ vs RabbitMQ vs Kafka ??

https://stackshare.io/stackups/kafka-vs-nsq-vs-rabbitmq

golang이란 거 빼고 장점이 없어보여서 많은 star를 가지고 많이 사용하는 [RabbitMQ](https://www.rabbitmq.com/tutorials/tutorial-one-go.html) 쓰기로 했음.



### RabbitMQ란?

참조한 [링크](https://github.com/gjchoi/gjchoi.github.io/blob/master/_posts/2016-02-27-rabbit-mq-%EC%9D%B4%ED%95%B4%ED%95%98%EA%B8%B0.md)

- ISO 표준(ISO/IEC 19464) AMQP 구현
- 비동기처리를 위한 메시지큐 브로커
- erlang과 java 언어로 만들어짐 (clustering 등을 위한 고속 데이터 전송하는 부분에 erlang이 사용된 것으로 추정)
- 분산처리를 고려한 MQ ( Cluster, Federation )
- 고가용성 보장 (High Availability)
- Publish/Subscribe 방식 지원
- 다양한 plugin 지원



`go get github.com/streadway/amqp` (amqp 구현체이기 때문에 그냥 amqp 쓰면 됩니다)

우선 그래서 받은 메시지를 저장하는 구조체를 만들었습니다.

```go
type Message struct {
	Text    string
	Channel int
}
```

위 홈페이지를 바탕으로 RabbitMQ에 전송하는 메소드를 만들어 보았다.

```go
func (m *Message) Send() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Faild to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		m.getChannelName(),
		false,
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "Failed to declare a queue")

	body := m.Text

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})

	failOnError(err, "Failed to publish a message")
}
```

보내면 받을 것도 있어야 합니답. noun 추출해서 다시 golang 단으로 전송하는 python 프로그램은 다음과 같습니다.



### Python

```python
#!/usr/bin/env python
import pika
from konlpy.tag import Mecab
from send import send_message_to_database
import threading


def callback(ch, method, properties, body, id):
    mecab = Mecab()
    print(" [%d] Received %s" % (ch, body.decode('utf-8')))

    noun_list = mecab.nouns(body.decode('utf-8'))

    send_message_to_database(id, noun_list)


def start_channel(id):
    connection = \
        pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
    channel = connection.channel()

    channel.queue_declare(queue=str.format('send_channel_{}', id))

    channel.basic_consume(lambda ch, method, properties, body: callback(ch, method, properties, body, id=id),
                          queue=str.format('send_channel_{}', id),
                          no_ack=True)

    print(' [%d] Waiting for messages. To exit press CTRL+C' % id)
    channel.start_consuming()


if __name__ == "__main__":
    thread_list = []
    for i in range(4):
        thread_list.append(threading.Thread(target=start_channel, args=(i,)))

    for i in thread_list:
        i.start()
```

 Send Message 해주는 함수. 뭔가 이거 짜면서 계속 람다를 쓰게 되는 거 같네요

```python
#!/usr/bin/env python
import pika


def send_message_to_database(id, str_list):
    connection = \
        pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
    channel = connection.channel()

    channel.queue_declare(queue=str.format('receive_channel_{}', id))

    for i in str_list:
        channel.basic_publish(exchange='',
                              routing_key=str.format('receive_channel_{}', id),
                              body=i)
        print(str.format(" [{}] Send '{}'", id, i))

    connection.close()
```



### Golang Receiver

```go
package main

import "sync"

var wg sync.WaitGroup

func main() {
	for i := 0; i < 4; i++ {
		wg.Add(1) // 1 개 만큼의 새로운 루틴이 실행됩니다.
		go receiver(i)
	}

	wg.Wait() // 다 끝날 때 까지 기다리겠습니다.

}
```



자 여기서 WaitGroup은 무엇인가?

고언어의 main은 암묵적 goroutine인데, 간혹 main routine이 먼저 종료되어 서브루틴이 제대로 실행되지 않고 종료되는 경우가 발생한다. ( python Threading 사용할 때 daemon을 True로 설정했을 때랑 비슷함. )

하여 특정 루틴의 실행이 다 완료될 때 까지 main routine이 살아있어야 하는 경우가 생긴다. 이 때, 특정 루틴이 다 사용될 때 까지 기다려 주는 역할을 하는 것이 바로 waitGroup이다.



```go
wg.Add(number) // number 개 만큼의 새로운 루틴이 실행됩니다.
wg.Done()      // 하나 끝났어요.
wg.Wait()      // 다 끝날 때 까지 기다리겠습니다.
```



이제 receiver 함수를 살펴보면 함수 초기에 defer 구문으로 `wg.Done()`을 선언한 모습을 볼 수 있다.

```go
func receiver(channel int) {
	defer wg.Done() //하나 끝났어요.
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		fmt.Sprintf("receive_channel_%d", channel), // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	for d := range msgs {
		new := Message{string(d.Body[:]), channel}
		messages <- new
    }
}
```



이제 충돌방지 작업을 해보기로 한다.

```go
package main

import "fmt"

func conflictPrevent(messages chan Message) {
	dataMap := make(map[string](chan Message))
	finMap := make(map[string](chan bool))

	for msg := range messages {
		//check mgo session is exist

		_, ok := dataMap[msg.Text]
		if ok {
			fmt.Println(fmt.Sprintf("Add data %s", msg.Text))
			dataMap[msg.Text] <- msg
		} else {
			fmt.Println(fmt.Sprintf("Create New Data %s", msg.Text))
			dataMap[msg.Text] = make(chan Message, 100)
			finMap[msg.Text] = make(chan bool)

			go save(dataMap, finMap, msg.Text)
			go finalize(dataMap, finMap, msg.Text)

			dataMap[msg.Text] <- msg
		}

	}
}

func finalize(dataMap map[string](chan Message), finMap map[string](chan bool), msg string) {
	defer delete(finMap, msg)

	for data := range finMap[msg] {
		_, ok := dataMap[msg]
		if ok && data {
			fmt.Println(fmt.Sprintf("Finalize %s", msg))
			delete(dataMap, msg)
		}

		return
	}
}
```

만일 현재 입력중인 데이터가 있으면 계속 해당 채널에 삽입하고,  아니면 새롭게 채널을 만든다. 새로운 입력이 없으면 채널을 제거한다.

```go
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type DataLogs struct {
	Text string   `bson:"text"`
	Log  []string `bson:"logs"`
}

func save(dataMap map[string](chan Message), finMap map[string](chan bool), msg string) {

	session, err := mgo.Dial("127.0.0.1")

	if err != nil {
		panic(err)
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("test").C("nouns")

	for value := range dataMap[msg] {
		var results []DataLogs

		err = c.Find(bson.M{"text": msg}).All(&results)

		if len(results) <= 0 {
			c.Insert(&DataLogs{Text: msg, Log: []string{fmt.Sprintf("Add Data from channel %d", value.Channel)}})
		} else {
			pushQuery := bson.M{"$push": bson.M{"logs": fmt.Sprintf("Add Data from channel %d", value.Channel)}}
			who := bson.M{"text": results[0].Text}

			c.Update(who, pushQuery)
		}

		time.Sleep(time.Duration(rand.Int()%3000+2000) * time.Millisecond)

		if len(dataMap[msg]) <= 0 {
			finMap[msg] <- true
			return
		}
	}
}
```



데이터베이스 저장은 이렇게.
