package serializer

import (
	"encoding/json"
	"log"
	"time"
)

const (
	HowImUsedToItBeing   = "2006-01-02 15:04:05 -0700"
	JavaScriptTimeFormat = "Mon, 02 Jan 2006 15:04:05 MST"
)

func MarshalJson(r interface{}) string {
	resp_json, err := json.Marshal(r)
	if err != nil {
		log.Println("Error converting ResponseMessage to JSON", r, err)
		return err.Error()
	}
	return string(resp_json)
}

func ParseJavaScriptTime(timeString string) *time.Time {
	loc, _ := time.LoadLocation("Local")
	t, err := time.ParseInLocation(JavaScriptTimeFormat, timeString, loc)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &t
}

func TimeToXML(t time.Time) string {
	return t.Format(HowImUsedToItBeing)
}
