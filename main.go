package main

import (
	"fmt"
	"github.com/bwmarrin/snowflake"
	"log"
	sf "snowflake-go/snowflake"
	"strconv"
	"time"
)

var currWoker = &sf.SnowFlake{}
var totalNum = 100000

func main() {
	testSnowflake()
	testMine()
}

func testMine()  {
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
	fmt.Println("MINE:重复数",repeatCount,len(dMap))
	fmt.Println("MINE:耗时",time.Now().UnixNano()/1000/1000-now,"毫秒")
}

func testSnowflake()  {
	// Create a new Node with a Node number of 1
	node, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	snowflake.Epoch = 1587994800000
	now :=time.Now().UnixNano()/1000/1000
	ch := make(chan int64,totalNum)
	for i:=0;i<totalNum;i++{
		go func() {
			ch <- int64(node.Generate())
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
	fmt.Println("Other:重复数",repeatCount,len(dMap))
	fmt.Println("Other:耗时",time.Now().UnixNano()/1000/1000-now,"毫秒")

	// Generate a snowflake ID.
	//id := node.Generate()
	//// Print out the ID in a few different ways.
	//fmt.Printf("Int64  ID: %d\n", id)
	//fmt.Printf("String ID: %s\n", id)
	//fmt.Printf("Base2  ID: %s\n", id.Base2())
	//fmt.Printf("Base64 ID: %s\n", id.Base64())
	//
	//// Print out the ID's timestamp
	//fmt.Printf("ID Time  : %d\n", id.Time())
	//
	//// Print out the ID's node number
	//fmt.Printf("ID Node  : %d\n", id.Node())
	//
	//// Print out the ID's sequence number
	//fmt.Printf("ID Step  : %d\n", id.Step())
	//
	//// Generate and print, all in one.
	//fmt.Printf("ID       : %d\n", node.Generate().Int64())
}