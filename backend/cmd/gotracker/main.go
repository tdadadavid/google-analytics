package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"

	g "github.com/tdadadavid/analytics"
)

var forceIp string

func track(w http.ResponseWriter, r *http.Request) {
	defer w.WriteHeader(http.StatusOK)

	data := r.URL.Query().Get("data")
	if data == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("[Error] tracking payload is not passed."))
	}

	payload, err := g.DecodedData(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	fmt.Println(payload)

	ua := g.ParseUserAgent(payload.Action.UserAgent)

	ip, err := g.IpFromRequest(nil, r, false, forceIp)
	if err != nil {
		fmt.Println("error getting IP: ", err)
		return
	}

	geoInfo, err := g.GetGeoInfo(ip.String())
	if err != nil {
		fmt.Println("error getting geo info: ", err)
		return
	}

	if len(payload.Action.Referrer) > 0 {
		u, err := url.Parse(payload.Action.Referrer)
		if err == nil {
			payload.Action.ReferrerHost = u.Host
		}
	}

	if len(payload.Action.Identity) == 0 {
		payload.Action.Identity = fmt.Sprintf("%s-%s", geoInfo.IP, ua.Name)
	}

	fmt.Println("Payload", payload)

	go events.Add(payload, ua, geoInfo);
}

func stats(w http.ResponseWriter, r *http.Request) {
	var data g.MetricData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var metrics []g.Metric
	var err error

	metrics, err = events.GetStats(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(metrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}


var (
	events *g.Events = &g.Events{Ch: make(chan g.ReqQueue)}
)

func main() {
	flag.StringVar(&forceIp, "ip", "", "force IP request, useful in local")
	flag.Parse()

	g.LoadConfig()

	if err := events.Open(); err != nil {
		log.Fatal(err)
	}
	if err := events.EnsureTable(); err != nil {
		log.Fatal(err)
	}
	go events.Run()

	http.HandleFunc("/track", track)
	http.HandleFunc("/stats", stats)
	fmt.Println("Server running on port 7500");
	fmt.Println(http.ListenAndServe(":7500", nil))
}