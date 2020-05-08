package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

func main() {
	totalNum := 1000000
	now :=time.Now().UnixNano()/1000/1000
	currWoker := &SnowFlake{}
	err := currWoker.InitSnowFlake(1023, 1023)
	if err != nil {
		log.Fatalln(err)
		return
	}
	ch := make(chan int64,totalNum)
	for i:=0;i<totalNum;i++{
		go func() {
			id,err :=currWoker.NextId()
			if err != nil {
				log.Fatalln(err)
				return
			}
			ch <- id
		}()
	}
	dMap := make(map[int64]bool)
	repeatCount := 0
	for i:=0;i<totalNum;i++{
		x := <-ch
		if len(strconv.FormatInt(x, 10)) != 19 {
			fmt.Println("长度变成了18位",x)
		}
		if dMap[x] {
			fmt.Println(x,"重复了")
			repeatCount++
		}
		dMap[x]= true
	}
	fmt.Println("重复数",repeatCount,len(dMap))
	fmt.Println("耗时",now-time.Now().UnixNano()/1000/1000)
}

type SnowFlake struct {
	startTime             int64//开始时间的偏移量。
	workerIdBits          uint//worker_id的长度
	dataCenterIdBits      uint
	maxWorkerId           int64
	maxDataCenterId       int64
	sequenceBits          uint
	workerIdLeftShift     uint
	dataCenterIdLeftShift uint
	timestampLeftShift    uint
	sequenceMask          int64
	workerId              int64
	dataCenterId          int64
	sequence              int64
	lastTimestamp         int64
	idLock                *sync.Mutex
}

func (sf *SnowFlake) InitSnowFlake(workerId, dataCenterId int64) error {
	if workerId >=1024||dataCenterId>=1024  {
		return errors.New("workerId or dataCenterId over max")
	}
	var baseValue int64 = -1
	sf.startTime = 1688927006000
	sf.workerIdBits = 5
	sf.dataCenterIdBits = 5
	// 这个是二进制运算，就是5 bit最多只能有31个数字，也就是说机器id最多只能是32以内
	sf.maxWorkerId = baseValue ^ (baseValue << sf.workerIdBits)
	// 这个是一个意思，就是5 bit最多只能有31个数字，机房id最多只能是32以内
	sf.maxDataCenterId = baseValue ^ (baseValue << sf.dataCenterIdBits)
	sf.sequenceBits = 12
	sf.workerIdLeftShift = sf.sequenceBits //12
	sf.dataCenterIdLeftShift = sf.workerIdBits + sf.workerIdLeftShift//17
	sf.timestampLeftShift = sf.dataCenterIdBits + sf.dataCenterIdLeftShift//22
	sf.sequenceMask = baseValue ^ (baseValue << sf.sequenceBits)
	sf.sequence = 0
	sf.lastTimestamp = -1
	sf.idLock = &sync.Mutex{}

	if sf.workerId < 0 || sf.workerId > sf.maxWorkerId {
		return errors.New(fmt.Sprintf("workerId[%v] is less than 0 or greater than maxWorkerId[%v].", workerId, dataCenterId))
	}
	if sf.dataCenterId < 0 || sf.dataCenterId > sf.maxDataCenterId {
		return errors.New(fmt.Sprintf("dataCenterId[%d] is less than 0 or greater than maxDataCenterId[%d].", workerId, dataCenterId))
	}
	sf.workerId = workerId
	sf.dataCenterId = dataCenterId
	return nil
}

func (sf *SnowFlake) NextId() (int64, error) {
	sf.idLock.Lock()
	timestamp := time.Now().UnixNano()
	//时间回拨直接return
	if timestamp < sf.lastTimestamp {
		return -1, errors.New(fmt.Sprintf("Clock moved backwards.  Refusing to generate id for %d milliseconds", sf.lastTimestamp-timestamp))
	}
	//同一毫秒，生成对应的sn
	if timestamp == sf.lastTimestamp {
		sf.sequence +=1
		if sf.sequence > 4096 {
			//超出单毫秒限制数量，等待下一毫秒
			timestamp = sf.waitNextMillis()
			sf.sequence = 0
		}
	} else {
		sf.sequence = 0
	}
	//更新最后时间
	sf.lastTimestamp = timestamp

	sf.idLock.Unlock()

	id := ((timestamp - sf.startTime) << sf.timestampLeftShift) |
		(sf.dataCenterId << sf.dataCenterIdLeftShift) |
		(sf.workerId << sf.workerIdLeftShift) |
		sf.sequence

	if id < 0 {
		id = -id
	}
	return id, nil
}

func (sf *SnowFlake) waitNextMillis() int64 {
	time.Sleep(time.Millisecond)//等待下一秒刷新
	timestamp := time.Now().UnixNano()//当前时间
	if timestamp <= sf.lastTimestamp {//如果内存时间大于当前时间，将当前时间刷新为正确时间
		timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	}
	return timestamp
}