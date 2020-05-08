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
	currWoker := &IdWorker{}
	err := currWoker.InitIdWorker(1023, 1023)
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
}

type IdWorker struct {
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

func (this *IdWorker) InitIdWorker(workerId, dataCenterId int64) error {
	if workerId >=1024||dataCenterId>=1024  {
		return errors.New("workerId or dataCenterId over max")
	}
	var baseValue int64 = -1
	this.startTime = 1688927006000
	this.workerIdBits = 5
	this.dataCenterIdBits = 5
	// 这个是二进制运算，就是5 bit最多只能有31个数字，也就是说机器id最多只能是32以内
	this.maxWorkerId = baseValue ^ (baseValue << this.workerIdBits)
	// 这个是一个意思，就是5 bit最多只能有31个数字，机房id最多只能是32以内
	this.maxDataCenterId = baseValue ^ (baseValue << this.dataCenterIdBits)
	this.sequenceBits = 12
	this.workerIdLeftShift = this.sequenceBits //12
	this.dataCenterIdLeftShift = this.workerIdBits + this.workerIdLeftShift//17
	this.timestampLeftShift = this.dataCenterIdBits + this.dataCenterIdLeftShift//22
	this.sequenceMask = baseValue ^ (baseValue << this.sequenceBits)
	this.sequence = 0
	this.lastTimestamp = -1
	this.idLock = &sync.Mutex{}

	if this.workerId < 0 || this.workerId > this.maxWorkerId {
		return errors.New(fmt.Sprintf("workerId[%v] is less than 0 or greater than maxWorkerId[%v].", workerId, dataCenterId))
	}
	if this.dataCenterId < 0 || this.dataCenterId > this.maxDataCenterId {
		return errors.New(fmt.Sprintf("dataCenterId[%d] is less than 0 or greater than maxDataCenterId[%d].", workerId, dataCenterId))
	}
	this.workerId = workerId
	this.dataCenterId = dataCenterId
	return nil
}

func (this *IdWorker) NextId() (int64, error) {
	this.idLock.Lock()
	timestamp := time.Now().UnixNano()
	//时间回拨直接return
	if timestamp < this.lastTimestamp {
		return -1, errors.New(fmt.Sprintf("Clock moved backwards.  Refusing to generate id for %d milliseconds", this.lastTimestamp-timestamp))
	}
	//同一毫秒，生成对应的sn
	if timestamp == this.lastTimestamp {
		this.sequence +=1
		if this.sequence > 4096 {
			//如果是0，代表超出了
			timestamp = this.tilNextMillis()
			this.sequence = 0
		}
	} else {
		this.sequence = 0
	}
	//更新最后时间
	this.lastTimestamp = timestamp

	this.idLock.Unlock()

	id := ((timestamp - this.startTime) << this.timestampLeftShift) |
		(this.dataCenterId << this.dataCenterIdLeftShift) |
		(this.workerId << this.workerIdLeftShift) |
		this.sequence

	if id < 0 {
		id = -id
	}
	return id, nil
}

func (this *IdWorker) tilNextMillis() int64 {
	time.Sleep(time.Millisecond)//等待下一秒刷新
	timestamp := time.Now().UnixNano()//当前时间
	if timestamp <= this.lastTimestamp {//如果内存时间大于当前时间，将当前时间刷新为正确时间
		timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	}
	return timestamp
}