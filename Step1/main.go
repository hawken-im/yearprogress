package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func timePerc(nextPost time.Time) (perc float64) { //calculate percentage
	initialTime := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	duration := nextPost.Sub(initialTime)
	perc = duration.Hours() / (365.0 * 24.0)
	log.Info("perc is:", perc)
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

func postToRum(content string, group string) {
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

	req, err := http.NewRequest("POST", "https://127.0.0.1:8002/api/v1/group/content", body)
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
	postToRum(printBar(timePerc(time.Now().UTC())), "[种子网络ID]")
}
