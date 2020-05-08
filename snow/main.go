package main

import (
	"fmt"
	"sync"
	"time"
)


var (

)
type SnowFlake struct {
	machineID     int64 // 机器 id 占10位, 十进制范围是 [ 0, 1023 ]
	sn            int64 // 序列号占 12 位,十进制范围是 [ 0, 4095 ]
	lastTimeStamp int64 // 上次的时间戳(毫秒级), 1秒=1000毫秒, 1毫秒=1000微秒,1微秒=1000纳秒
	lock *sync.Mutex
}

var flake SnowFlake

func init() {
	flake.sn = 0
	flake.machineID = 10
	flake.lastTimeStamp = time.Now().UnixNano() / 1000000
	flake.lock = &sync.Mutex{}
}



func SetMachineId(mid int64) {
	// 把机器 id 左移 12 位,让出 12 位空间给序列号使用
	flake.machineID = mid << 12
}


func GetSnowflakeId() int64 {
	curTimeStamp := time.Now().UnixNano() / 1000000
	flake.lock.Lock()
	defer func() {
		flake.lock.Unlock()
	}()
	flake.sn++
	// 同一毫秒
	if curTimeStamp == flake.lastTimeStamp {
		// 序列号占 12 位,十进制范围是 [ 0, 4095 ]
		if flake.sn > 4095 {
			time.Sleep(time.Millisecond)
			curTimeStamp += 1
			flake.lastTimeStamp = curTimeStamp
			flake.sn= 0
		}
		// 取 64 位的二进制数 0000000000 0000000000 0000000000 0001111111111 1111111111 1111111111  1 ( 这里共 41 个 1 )和时间戳进行并操作
		// 并结果( 右数 )第 42 位必然是 0,  低 41 位也就是时间戳的低 41 位
		rightBinValue := curTimeStamp & 0x1FFFFFFFFFF
		// 机器 id 占用10位空间,序列号占用12位空间,所以左移 22 位; 经过上面的并操作,左移后的第 1 位,必然是 0
		rightBinValue <<= 42
		id := rightBinValue | flake.machineID | flake.sn
		return id
	}

	if curTimeStamp > flake.lastTimeStamp {
		if flake.sn > 4095 {
			time.Sleep(time.Millisecond)
			curTimeStamp +=1
			flake.lastTimeStamp = curTimeStamp
			flake.sn= 0
		}
		flake.lastTimeStamp = curTimeStamp
		// 取 64 位的二进制数 0000000000 0000000000 0000000000 0001111111111 1111111111 1111111111  1 ( 这里共 41 个 1 )和时间戳进行并操作
		// 并结果( 右数 )第 42 位必然是 0,  低 41 位也就是时间戳的低 41 位
		rightBinValue := curTimeStamp & 0x1FFFFFFFFFF
		// 机器 id 占用10位空间,序列号占用12位空间,所以左移 22 位; 经过上面的并操作,左移后的第 1 位,必然是 0
		rightBinValue <<= 42
		id := rightBinValue | flake.machineID | flake.sn
		return id
	}


	if curTimeStamp < flake.lastTimeStamp {
		fmt.Println(1)
		return 0
	}
	fmt.Println(2)

	return 0
}

func main() {
	totalNum := 4096
	ch := make(chan int64,totalNum)
	for i:=0;i<totalNum;i++{
		go func() {
			ch <-GetSnowflakeId()
		}()
		//6664444700541321222
		//6664444768174473223
		//6664446356674838536
	}
	dMap := make(map[int64]bool)
	repeatCount := 0
	for i:=0;i<totalNum;i++{
		x := <-ch
		//fmt.Println(x)
		if dMap[x] {
			repeatCount++
		}
		dMap[x]= true
	}
	fmt.Println("重复数",repeatCount,"长度",len(dMap))
}