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
