package ganalytics

import (
	"context"

	// "database/sql/driver"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type QueryType int

const (
	QueryPageViews QueryType = iota
	QueryPageViewList
	QueryUniqueVisitors
	QueryReferrerHost
	QueryReferrer
	QueryBrowsers
	QueryOs
	QueryCountry
)

type Event struct {
	ID          int64
	SiteID      string
	OccuredAt   int32
	Type        string
	UserID      string
	Event       string
	Category    string
	Referrer    string
	IsTouch     bool
	BrowserName string
	OSName      string
	DeviceType  string
	Country     string
	Region      string
	Timestamp   time.Time
}

type MetricData struct {
	What   QueryType `json:"what"`
	SiteID string `json:"siteId"`
	Start  uint32 `json:"start"`
	End    uint32 `json:"end"`
	Extra  string `json:"extra"`
}

type Metric struct {
	OccuredAt uint32 `json:"occuredAt"`
	Value     string `json:"value"`
	Count     uint64 `json:"count"`
}

type ReqQueue struct {
	trk Tracking
	ua  UserAgent
	geo *GeoInfo
}

type Events struct {
	DB   driver.Conn
	Ch   chan ReqQueue
	lock sync.RWMutex
	q    []ReqQueue
}

func (e *Events) Open() error {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{config.ClickHouseHost},
		Auth: clickhouse.Auth{
			Database: config.ClickHouseDB,
			Username: config.ClickHouseUser,
			Password: config.ClickHousePassword,
		},
	})
	if err != nil {
		fmt.Println("error", err.Error())
		return err
	}

	if err := conn.Ping(context.Background()); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return err
	}

	e.DB = conn

	return err
}

func (e *Events) Insert() error {
	query := `
		INSERT INTO events 
		(
			site_id, occured_at, type, user_id, event,
			category, referrer, referrer_domain, is_touch, 
			browser_name, os_name, device_type, 
			country, region
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14
		);
	`

	var tmp []ReqQueue
	e.lock.Lock()
	tmp = append(tmp, e.q...)
	e.q = nil
	e.lock.Unlock()

	ctx := context.Background()
	batch, err := e.DB.PrepareBatch(ctx, query)
	if err != nil {
		return err
	}

	for _, qd := range tmp {
		err := batch.Append(
			qd.trk.SiteID,
			TimeToInt(time.Now()),
			qd.trk.Action.Type,
			qd.trk.Action.Identity,
			qd.trk.Action.Event,
			qd.trk.Action.Category,
			qd.trk.Action.Referrer,
			qd.trk.Action.ReferrerHost,
			qd.trk.Action.IsTouchDevice,
			qd.ua.Name,
			qd.ua.OS,
			qd.ua.Device,
			// strings.ToLower(qd.ua.Name),
			// strings.ToLower(qd.ua.OS),
			// strings.ToLower(qd.ua.Device),
			qd.geo.Country,
			qd.geo.RegionName,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

func (e *Events) EnsureTable() error {
	create_table_qry := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s  ENGINE = Atomic", config.ClickHouseDB)

	qry := `		
		CREATE TABLE IF NOT EXISTS events (
			site_id String NOT NULL,
			occured_at UInt32 NOT NULL,
			type String NOT NULL,
			user_id String NOT NULL,
			event String NOT NULL,
			category String NOT NULL,
			referrer String NOT NULL,
			referrer_domain String NOT NULL,
			is_touch BOOLEAN NOT NULL,
			browser_name String NOT NULL,
			os_name String NOT NULL,
			device_type String NOT NULL,
			country String NOT NULL,
			region String NOT NULL,
			timestamp DateTime DEFAULT now()
		)
		ENGINE MergeTree
		ORDER BY (site_id, occured_at);
	`

	if err := e.DB.Exec(context.Background(), create_table_qry); err != nil {
		return err
	}

	return e.DB.Exec(context.Background(), qry)
}

func (e *Events) Add(trk Tracking, ua UserAgent, geo *GeoInfo) {
	e.Ch <- ReqQueue{trk: trk, ua: ua, geo: geo}
}

func (e *Events) Run() {
	timer := time.NewTimer(time.Second * 10)
	for {
		select {
		// if there is data in the channel,
		case data := <-e.Ch:
			e.lock.Lock() // prepare to write into our queue by locking it.
			e.q = append(e.q, data)
			qLen := len(e.q)
			e.lock.Unlock()

			if qLen >= 15 {
				if err := e.Insert(); err != nil {
					fmt.Println("error while inserting data: ", err)
				}
			}
		case <-timer.C:
			timer.Reset(time.Second * 10)

			e.lock.RLock()
			qLen := len(e.q)
			e.lock.RUnlock()

			if qLen > 0 {
				if err := e.Insert(); err != nil {
					fmt.Println("error while inserting data: ", err)
				}
			}
		}

	}
}

func (e *Events) GetStats(m MetricData) ([]Metric, error) {
	qry := e.GenQuery(m)
	rows, err := e.DB.Query(
		context.Background(), 
		qry, 
		m.SiteID, 
		m.Start, 
		m.End,
		m.Extra,
	)
	if err != nil {
		return nil, err
	}

	var metrics []Metric
	for rows.Next() {
		var m Metric
		if err := rows.Scan(&m.OccuredAt, &m.Value, &m.Count); err != nil {
			return nil, err
		}

		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

func TimeToInt(d time.Time) uint32 {
	now := d.Format("20060102")
	i, err := strconv.ParseInt(now, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	return uint32(i)
}


func (e *Events) GenQuery(data MetricData) string {
	field := ""
	daily := true
	where := "AND $4 = $4"
	switch data.What {
	case QueryPageViews:
		field = "event"
	case QueryPageViewList:
		field = "event"
		daily = false
	case QueryUniqueVisitors:
		field = "user_id"
	case QueryReferrerHost:
		field = "referrer_domain"
		daily = false
	case QueryReferrer:
		field = "referrer"
		where = "AND referrer_domain = $3"
		daily = false
	case QueryOs:
		field = "os_name"
		daily = false
	case QueryCountry:
		field = "country"
		daily = false
	case QueryBrowsers:
		field = "browser_name"
		daily = false
	}
	
	if daily {
		return fmt.Sprintf(`
			SELECT occured_at, %s, COUNT(*) AS count
			FROM events
			WHERE site_id = $1
			AND category = 'Page views'
			GROUP BY occured_at, %s 
			HAVING occured_at BETWEEN $2 AND $3
			ORDER BY 3 DESC
		`,field, field)
	}

	return  fmt.Sprintf(`
		SELECT toUInt32(0), %s, COUNT(*)
		FROM events
		WHERE site_id = $1
		AND occured_at BETWEEN $2 AND $3
		AND category = 'Page views'
		%s
		GROUP BY %s
		ORDER BY 3 DESC 
	`, field, where, field)
}