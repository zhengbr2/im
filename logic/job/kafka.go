package main

import (
	"sync"

	"github.com/Shopify/sarama"
)

const (
	KafkaPushsTopic = "KafkaPushsTopic"
)

var (
	wg sync.WaitGroup
)

func InitKafka(kafkaAddrs []string) error {
	consumer, err := sarama.NewConsumer(kafkaAddrs, nil)
	if err != nil {
		return err
	}

	var partitionList []int32
	partitionList, err = consumer.Partitions(KafkaPushsTopic)
	if err != nil {
		return err
	}

	for partions := range partitionList {
		pc, err := consumer.ConsumePartition(KafkaPushsTopic, int32(partions), sarama.OffsetNewest)
		if err != nil {
			continue
		}
		defer pc.AsyncClose()

		wg.Add(1)
		go func(sarama.PartitionConsumer) {
			defer wg.Done()
			for msg := range pc.Messages() {
				push(msg.Value)
			}
		}(pc)
	}

	wg.Wait()
	consumer.Close()
	return nil
}
