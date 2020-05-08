package main

import (
	"fmt"
	"log"
	"snowflake-go/snowflake"
	"strconv"
	"time"
)

var currWoker = &snowflake.SnowFlake{}

func main() {
	totalNum := 100000
	now :=time.Now().UnixNano()/1000/1000
	err := currWoker.InitSnowFlake(1, 1)
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
			//fmt.Println(x,"重复了",currWoker.sequence,currWoker.lastTimestamp)
			repeatCount++
		}
		dMap[x]= true
	}
	fmt.Println("重复数",repeatCount,len(dMap))
	fmt.Println("耗时",time.Now().UnixNano()/1000/1000-now,"毫秒")
}
