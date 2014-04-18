package main

import (
  "fmt"
  "time"
  kinesis "github.com/sendgridlabs/go-kinesis"
)

func getRecords(ksis *kinesis.Kinesis, streamName, ShardId string) {
  args := kinesis.NewArgs()
  args.Add("StreamName", streamName)
  args.Add("ShardId", ShardId)
  args.Add("ShardIteratorType", "TRIM_HORIZON")
  resp10, _ := ksis.GetShardIterator(args)

  shardIterator := resp10.ShardIterator

  for {
    args = kinesis.NewArgs()
    args.Add("ShardIterator", shardIterator)
    resp11, err := ksis.GetRecords(args)

    if len(resp11.Records) > 0 {
      fmt.Printf("GetRecords Data BEGIN\n")
      for _, d := range resp11.Records {
        res, err := d.GetData()
        fmt.Printf("GetRecords Data: %v, err: %v\n", string(res), err)
      }
      fmt.Printf("GetRecords Data END\n")
    } else if resp11.NextShardIterator == "" || shardIterator == resp11.NextShardIterator || err != nil {
      fmt.Printf("GetRecords ERROR: %v\n", err)
      break
    }

    shardIterator = resp11.NextShardIterator
  }
}

func main() {
  fmt.Println("Begin")

  streamName := "test"
  // set env variables AWS_ACCESS_KEY and AWS_SECRET_KEY
  ksis := kinesis.New("", "")

  err := ksis.CreateStream(streamName, 2)
  if err != nil {
    fmt.Printf("CreateStream ERROR: %v\n", err)
  }

  args := kinesis.NewArgs()
  resp2, _ := ksis.ListStreams(args)
  fmt.Printf("ListStreams: %v\n", resp2)

  resp3 := &kinesis.DescribeStreamResp{}

  timeout := make(chan bool, 30)
  for {

    args = kinesis.NewArgs()
    args.Add("StreamName", streamName)
    resp3, _ = ksis.DescribeStream(args)
    fmt.Printf("DescribeStream: %v\n", resp3)

    if resp3.StreamDescription.StreamStatus != "ACTIVE" {
      time.Sleep(4 * time.Second)
      timeout <- true
    } else {
      break
    }

  }


  for i := 0; i < 10; i++ {
    args = kinesis.NewArgs()
    args.Add("StreamName", streamName)
    args.AddData([]byte(fmt.Sprintf("Hello AWS Kinesis %d", i)))
    args.Add("PartitionKey", fmt.Sprintf("partitionKey-%d", i))
    resp4, err := ksis.PutRecord(args)
    if err != nil {
      fmt.Printf("PutRecord err: %v\n", err)
    } else {
      fmt.Printf("PutRecord: %v\n", resp4)
    }
  }

  for _, shard := range resp3.StreamDescription.Shards {
    go getRecords(ksis, streamName, shard.ShardId)
  }

  var inputGuess string
  fmt.Scanf("%s\n", &inputGuess)

  err1 := ksis.DeleteStream("test")
  if err1 != nil {
    fmt.Printf("DeleteStream ERROR: %v\n", err1)
  }

  fmt.Println("End")
}