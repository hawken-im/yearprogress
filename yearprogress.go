package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"time"

	readconfig "github.com/hawken-im/yearprogress/readconfig"

	cron "github.com/robfig/cron/v3"

	log "github.com/sirupsen/logrus"
)

func timePerc(nextPost time.Time) (perc float64) { //calculate percentage
	initialTime := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	duration := nextPost.Sub(initialTime)
	log.Info("duration is:", duration)
	perc = duration.Hours() / (365.0 * 24.0)
	log.Info("perc is:", perc)
	return
}

func printBar(perc float64) (bar string) { //print progress bar by percentage
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

	gB := perc*ttlBs - math.Floor(perc*ttlBs) //to decide which gab block to chose.
	log.Info("the gap block indicator is:", gB)
	if gB < 0.0001 && perc < 0.9999 {
		bar += emptyB
	} else if gB >= 0.0001 && gB < 0.35 {
		bar += quarterB
	} else if gB >= 0.35 && gB < 0.6 {
		bar += halfB
	} else if gB >= 0.6 && gB < 0.85 {
		bar += threeQuartersB
	} else if perc >= 0.9999 {
		log.Info("quit earlier to prevent an extra empty block ", perc*ttlBs)
		return
	} else {
		bar += fullB
	}
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

func postToRum(content string, group string, url string) { //to generate quorum http post
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
			Name:    "2022 进度条 / Year Progress 2022",
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
	configs := readconfig.ReadConfig("config.json") //default config file

	flagConfig := flag.String("config", "config.json", "config file")
	flagGroupID := flag.String("gid", "test", "group ID, default ID is for testing")
	flagTest := flag.Bool("test", false, "test mode")
	flag.Parse()
	configs = readconfig.ReadConfig(*flagConfig)
	if *flagTest { //test quorum network by adding flag "-test".
		postToRum(printBar(timePerc(time.Now().UTC())), *flagGroupID, configs.URL)
	}

	c := cron.New(cron.WithLocation(time.UTC)) //new cron with location
	for {                                      // the infinate loop, designed to run whole 2022
		startTime := time.Date(2022, time.Now().UTC().Month(), time.Now().UTC().Day(), time.Now().UTC().Hour(), time.Now().Minute(), 0, 0, time.UTC)
		log.Info("---\nstartTime:", startTime)
		for x := 0; x <= 14; x++ {
			addMinutes, _ := time.ParseDuration(fmt.Sprintf("%dm", x))
			log.Info("addMinutes:", addMinutes)
			realTimePerc := timePerc(startTime.Add(addMinutes))
			log.Info("realTimePerc:", realTimePerc)
			roundPerc := math.Ceil(realTimePerc*100) / 100
			log.Info("roundPerc:", roundPerc)
			differVal := roundPerc - realTimePerc
			log.Info("differVal:", differVal)
			if differVal < 0.00001 { // calculating every one minute, so the difference between rounded percentage(1%) and realtime percentage is less than 0.00001.
				realTime := startTime.Add(addMinutes)
				log.Info("differVal less than 0:", differVal)
				nextPostTime := fmt.Sprintf("%d %d %d %d *", realTime.Minute(), realTime.Hour(), realTime.Day(), realTime.Month())
				log.Info("nextPostTime:", nextPostTime)
				for _, groupID := range configs.Groups {
					if !groupID.TestGroup {
						c.AddFunc(nextPostTime, func() { postToRum(printBar(roundPerc), groupID.ID, configs.URL) })
						log.Info("posting to group ID:", groupID.ID)
					}
				}
				c.Start()
				log.Info("######## went to sleep for 85 hours ########")
				fmt.Println("######## went to sleep for 85 hours ########")
				time.Sleep(85 * time.Hour) // sleep for 85 hours
				break
			}
		}
		log.Info("######## went to sleep ########")
		fmt.Println("######## went to sleep ########")
		time.Sleep(15 * time.Minute)
		c.Stop()
		log.Info("############ awaken ###########")
		fmt.Println("############ awaken ###########")
	}
}
