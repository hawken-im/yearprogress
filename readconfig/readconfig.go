package readconfig

import (
	"encoding/json"
	"io/ioutil"
)

type Cron struct {
	Method   string `json:"method"`
	Schedule string `json:"schedule"`
}
type Group struct {
	Name      string `json:"name"`
	ID        string `json:"ID"`
	TestGroup bool   `json:"testGroup"`
	Cron      Cron   `json:"cron"`
	TimeZone  string `json:"timeZone"`
}
type Configs struct {
	URL    string  `json:"url"`
	Groups []Group `json:"groups"`
}

func ReadConfig(jsonFile string) *Configs { // to read config file
	rawData, _ := ioutil.ReadFile(jsonFile) // filename is the JSON file to read
	//fmt.Println(string(rawData))

	var configs Configs
	json.Unmarshal(rawData, &configs)
	//	data = groups.Groups[1].ID
	json.Unmarshal(rawData, &configs)
	return &configs
}
