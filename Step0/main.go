package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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
	url := "https://127.0.0.1:8002/api/v1/group/content"                             //Rum 定义的 api
	postToRum("Hello Rum", "Hello Rum", "fe2842cb-db6b-4e8a-b007-e83e5603131c", url) //发布 Hello Rum 到Go语言学习小组
}
