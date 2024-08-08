package ganalytics

import (
	"encoding/base64"
	"encoding/json"

	"github.com/mileusna/useragent"
)

type Tracking struct {
	SiteID string       `json:"site_id"`
	Action TrackingData `json:"tracking"`
}

type TrackingData struct {
	Type          string `json:"type"`
	Identity      string `json:"identity"`
	UserAgent     string `json:"userAgent"`
	Event         string `json:"event"`
	Category      string `json:"category"`
	IsTouchDevice bool   `json:"isTouchDevice"`
	Referrer      string `json:"referrer"`
	ReferrerHost string `json:"referrerHost"`
}

type UserAgent struct {
	// Name of the browser
	Name   string
	// The OS underneath the browser
	OS     string
	// The Device used for browsing.
	Device string
	// Trackers wether the user is bot.
	Bot    bool
}

func DecodedData(s string) (data Tracking, err error) {
	d, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return
	}
	err = json.Unmarshal(d, &data)
	return
}

func ParseUserAgent(data string) UserAgent {
	ua := useragent.Parse(data)
	return UserAgent{
		Name:   ua.Name,
		OS:     ua.OS,
		Device: ua.Device,
		Bot:    ua.Bot,
	}
}