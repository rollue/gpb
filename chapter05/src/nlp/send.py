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
