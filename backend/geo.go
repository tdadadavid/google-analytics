package ganalytics

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type GeoInfo struct {
	IP         string  `json:"ip"`
	Country    string  `json:"country"`
	CountryISO string  `json:"country_iso"`
	RegionName string  `json:"region_name"`
	RegionCode string  `json:"region_code"`
	City       string  `json:"city"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

func GetGeoInfo(ip string) (*GeoInfo, error) {
	req, err := http.NewRequest("GET", config.EchoIPHost+"/json?ip="+ip, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err  != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info GeoInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	return &info, err
}

func IpFromRequest(headers []string, r *http.Request, customIP bool, forceIp string) (net.IP, error) {
	remoteIP := ""
	if customIP && r.URL != nil {
		if v, ok := r.URL.Query()["ip"]; ok {
			remoteIP = v[0]
		}
	}
	if remoteIP == "" {
		for _, header := range headers {
			remoteIP = r.Header.Get(header)
			if http.CanonicalHeaderKey(header) == "X-Forwarded-For" {
				remoteIP = ipFromForwardedForHeader(remoteIP)
			}
			if remoteIP != "" {
				break
			}
		}
	}
	if remoteIP == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return nil, err
		}
		remoteIP = host
	}

	if len(forceIp) > 0 {
		remoteIP = forceIp;
	}
	ip := net.ParseIP(remoteIP)
	if ip == nil {
		return nil, fmt.Errorf("could not parse IP: %s", remoteIP)
	}
	return ip, nil
}

func ipFromForwardedForHeader(v string) string {
	sep := strings.Index(v, ",")
	if sep == -1 {
		return v
	}
	return v[:sep]
}