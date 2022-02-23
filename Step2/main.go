package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

func timePerc(nextPost time.Time) (perc float64) { //calculate percentage
	initialTime := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	duration := nextPost.Sub(initialTime)
	perc = duration.Hours() / (365.0 * 24.0)
	return
}

func printBar(perc float64) (bar string) {
	const fullB string = "\u2588"          //0.9
	const halfB string = "\u2584"          //0.5
	const quarterB string = "\u2582"       //0.25
	const threeQuartersB string = "\u2586" //0.75
	const emptyB string = "\u2581"         //0
	const ttlBs float64 = 30               //total number of blocks
	bar = ""
	fBs := int(math.Floor(perc * ttlBs))
	for i := 0; i < fBs; i++ {
		bar += fullB
	}

	gB := perc*ttlBs - math.Floor(perc*ttlBs)
	fmt.Println("the gap block indicator is:", gB)
	if gB < 0.0001 && perc < 0.9999 {
		bar += emptyB
	} else if gB >= 0.0001 && gB < 0.35 {
		bar += quarterB
	} else if gB >= 0.35 && gB < 0.6 {
		bar += halfB
	} else if gB >= 0.6 && gB < 0.85 {
		bar += threeQuartersB
	} else if perc >= 0.9999 { //quit earlier to prevent an extra empty block
		return
	} else {
		bar += fullB
	}
	///
	eBs := int(ttlBs) - fBs - 1
	for i := 0; i < eBs; i++ {
		bar += emptyB
	}

	content := ""
	content += "2022 进度条 / Year Progress 2022\n"

	content += bar

	now := time.Now().UTC()
	displayPerC := fmt.Sprintf("%.1f", perc*100) + "%"
	bar = content + displayPerC + "\nUTC时间: " + now.Format("2006, Jan 02, 15:04:05") + "\n"
	return
}

func postToRum(title string, content string, group string, url string) { //to generate quorum http post
	type Object struct {
		Type    string `json:"type"`
		Content string `json:"content"`
		Name    string `json:"name"`
	}
	type Target struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	type Payload struct {
		Type   string `json:"type"`
		Object Object `json:"object"`
		Target Target `json:"target"`
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	data := Payload{
		Type: "Add",
		Object: Object{
			Type:    "Note",
			Content: content,
			Name:    title,
		},
		Target: Target{
			ID:   group,
			Type: "Group",
		},
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		panic(err) // handle err
	}

	fmt.Println(string(payloadBytes))

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	received, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(received))
}

func main() {
	f, err := os.OpenFile("YP.log", os.O_WRONLY|os.O_CREATE, 0755) //log file
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)

	c := cron.New(cron.WithLocation(time.UTC))
	url := "https://127.0.0.1:8002/api/v1/group/content" //Rum 定义的 api

	for {
		startTime := time.Date(2022, time.Now().UTC().Month(), time.Now().UTC().Day(), time.Now().UTC().Hour(), time.Now().Minute(), 0, 0, time.UTC) //开始时间
		log.Info("startTime:", startTime)                                                                                                            //记录一下循环开始时间
		for x := 0; x <= 14; x++ {                                                                                                                   //循环15次，下一个15分钟每分钟一次
			addMinutes, _ := time.ParseDuration(fmt.Sprintf("%dm", x)) //每次循环，在开始时间前加x分钟
			log.Info("addMinutes:", addMinutes)                        //记录一下每次加的时间对不对
			realTimePerc := timePerc(startTime.Add(addMinutes))
			log.Info("realTimePerc:", realTimePerc)        //加了时间之后的百分比，记录一下这个增长过程
			roundPerc := math.Ceil(realTimePerc*100) / 100 //计算下一个整数百分比
			log.Info("roundPerc:", roundPerc)              //虽然每次都是一样的值，但还是想看看
			differVal := roundPerc - realTimePerc          //计算差值，差值接近于零代表时间接近整数百分比了
			log.Info("differVal:", differVal)              //看看差值的变化过程，越来越接近于零
			if differVal < 0.00001 {                       //每分钟计算一次，每分钟是一年的0.000002，因此精确到小数点后5位
				realTime := startTime.Add(addMinutes)
				log.Info("differVal less than 0:", differVal) //终于到整百分点了，记录一个
				nextPostTime := fmt.Sprintf("%d %d %d %d *", realTime.Minute(), realTime.Hour(), realTime.Day(), realTime.Month())
				log.Info("nextPostTime:", nextPostTime) //报告具体的整百分点发布时间
				progressBar := printBar(roundPerc)
				c.AddFunc(nextPostTime, func() { postToRum("2022 进度条", progressBar, "fe2842cb-db6b-4e8a-b007-e83e5603131c", url) }) //设置定时任务
				c.Start()
				log.Info("######## went to sleep for 85 hours ########")    //日志里也记录一下                                                                         //开始定时任务
				fmt.Println("######## went to sleep for 85 hours ########") //休眠85个小时，因为一个百分比大概接近87个小时
				time.Sleep(85 * time.Hour)
				break
			}
		}
		log.Info("######## went to sleep ########")
		fmt.Println("######## went to sleep ########") //休眠15分钟
		time.Sleep(15 * time.Minute)
		c.Stop()
		log.Info("############ awaken ###########")
		fmt.Println("############ awaken ###########") //唤醒
	}
}
