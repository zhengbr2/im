package main

import (
	"encoding/json"
	"im/libs/proto"

	"github.com/Shopify/sarama"
)

const (
	KafkaPushsTopic = "KafkaPushsTopic"
)

var (
	producer sarama.AsyncProducer
)

func InitKafka(kafkaAddrs []string) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = 1
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err = sarama.NewAsyncProducer(kafkaAddrs, config)
	if err != nil {
		return
	}
	go handleSuccess()
	go handleError()
	return
}

func handleSuccess() {
	var (
		pm *sarama.ProducerMessage
	)
	for {
		pm = <-producer.Successes()
		if pm != nil {
			// producer message success, partition:pm.Partition offset:pm.Offset key:pm.Key value:pm.Value
		}
	}
}

func handleError() {
	var (
		err *sarama.ProducerError
	)
	for {
		err = <-producer.Errors()
		if err != nil {
			// producer message error, partition:pm.Partition offset:pm.Offset key:pm.Key value:pm.Value
		}
	}
}

func mpushKafka(key string, roomid int32, p *proto.Proto) (err error) {
	var (
		vBytes []byte
		v      = &proto.KafkaMsg{Key: key, RoomId: roomid, Operation: p.Operation, Body: p.Body}
	)
	if vBytes, err = json.Marshal(v); err != nil {
		return
	}
	producer.Input() <- &sarama.ProducerMessage{Topic: KafkaPushsTopic, Value: sarama.ByteEncoder(vBytes)}
	return
}
