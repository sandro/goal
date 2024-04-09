package goal

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sandro/go-sqlite-lite/sqx"
	"github.com/sandro/imigrate"
	"zgo.at/isbot"

	_ "embed"
)

//go:embed goal.js
var GoalJS string

//go:embed migrations/*.sql
var migrationsDir embed.FS

type Configuration struct {
	HitBufferSize   int
	PersistInterval time.Duration
	DBPath          string
	GoalIDKey       string
}

var DefaultConfig = Configuration{
	HitBufferSize:   10000,
	PersistInterval: time.Second * 2,
	DBPath:          "goal-01-09-2024.sqlite3",
	GoalIDKey:       "hitID",
}
var Config Configuration
var hitQueue chan Hit
var updateHitQueue chan Hit
var started = false
var pollStop = make(chan bool)
var pool *sqx.DBPool
var ErrNotStarted = fmt.Errorf("goal not started. Use Start function")
var ctx context.Context
var ctxCancel context.CancelFunc
var hourDuration = time.Second * 60 * 60
var dayDuration = hourDuration * 24

func Start(config ...Configuration) {
	if started {
		return
	}
	ctx, ctxCancel = context.WithCancel(context.Background())
	Config = DefaultConfig
	if len(config) > 0 {
		Config = config[0]
	}
	hitQueue = make(chan Hit, Config.HitBufferSize)
	updateHitQueue = make(chan Hit, Config.HitBufferSize)
	pool = sqx.NewDBPool("file:"+Config.DBPath, 5)
	migrate()
	go poll()
	go summarizeJob()
	started = true
}

func Stop() {
	pollStop <- true
	ctxCancel()
	started = false
}

func migrate() {
	conn := pool.CheckoutWriter()
	defer pool.CheckinWriter(conn)
	migrator := imigrate.NewIMigrator(conn, http.FS(migrationsDir))
	imigrate.Logger = imigrate.DiscardLogger
	migrator.Up(-1, 0)
}

func poll() {
	ticker := time.NewTicker(Config.PersistInterval)
	defer ticker.Stop()
	for {
		select {
		case <-pollStop:
			persistHits()
			return
		case <-ticker.C:
			persistHits()
		}
	}
}

func persistHits() {
	insertHits()
	updateHits()
}

func insertHits() {
	size := len(hitQueue)
	if size < 1 {
		return
	}

	conn := pool.CheckoutWriter()
	defer pool.CheckinWriter(conn)

	bulk := sqx.NewBulkInserter(
		`INSERT INTO hits (id, path, query, visitor_id, ip, referer, user_agent, browser, os, response_time, is_bot, created_at)`,
		`ON CONFLICT(id) DO UPDATE set
			path=excluded.path,
			query=excluded.query,
			visitor_id=excluded.visitor_id,
			ip=excluded.ip,
			referer=excluded.referer,
			user_agent=excluded.user_agent,
			browser=excluded.browser,
			os=excluded.os,
			response_time=excluded.response_time,
			is_bot=excluded.is_bot
			`,
		conn)
	for i := 0; i < size; i++ {
		hit := <-hitQueue
		fmt.Println("inserting path", hit.Path)
		bulk.Add(
			hit.ID,
			hit.Path,
			hit.Query,
			hit.VisitorID,
			hit.IP,
			hit.Referer,
			hit.UserAgent,
			hit.Browser,
			hit.OS,
			hit.ResponseTime,
			hit.IsBot,
			time.Now().UnixMilli(),
		)
	}

	err := bulk.Done()
	if err != nil {
		fmt.Println("INSERT ERR", err)
	}
}

func updateHits() {
	size := len(updateHitQueue)
	if size < 1 {
		return
	}

	// pool.Tx(func (conn *sqx.Conn) error {

	// for i := 0; i < size; i++ {
	// 	hit := <-updateHitQueue
	// 	conn.UpdateValues("update hits set", map[string]any{
	// 		"title": hit.Title,
	// 		"time_on_page": hit.TimeOnPage,
	// 		"width": hit.Width,
	// 		"height": hit.Height,
	// 		"device_pixel_ratio": hit.DevicePixelRatio,
	// 	}, "where id=?", hit.ID)
	// }
	// 	return nil
	// })
	conn := pool.CheckoutWriter()
	defer pool.CheckinWriter(conn)

	compressedHits := make(map[string]Hit)
	for i := 0; i < size; i++ {
		hit := <-updateHitQueue
		compressedHits[hit.ID] = hit
	}
	bulk := sqx.NewBulkInserter(
		`INSERT INTO hits (id, path, query, title, time_on_page, width, height, device_pixel_ratio)`,
		`ON CONFLICT(id) DO UPDATE set
			path=coalesce(excluded.path, path),
			query=coalesce(excluded.query, query),
			title=coalesce(excluded.title, title),
			time_on_page=coalesce(excluded.time_on_page, time_on_page),
			width=coalesce(excluded.width, width),
			height=coalesce(excluded.height, height),
			device_pixel_ratio=coalesce(excluded.device_pixel_ratio, device_pixel_ratio)
			`,
		conn)
	for _, hit := range compressedHits {
		bulk.Add(
			hit.ID,
			hit.Path,
			hit.Query,
			hit.Title,
			hit.TimeOnPage,
			hit.Width,
			hit.Height,
			hit.DevicePixelRatio,
		)
	}
	err := bulk.Done()
	if err != nil {
		fmt.Println("UPDATE  ERR", err)
	}
}

type RecordExtra struct {
	HitID            string
	VisitorID        string
	Title            string
	ResponseTime     int64 // ms
	TimeOnPage       int64 // ms
	Width            int
	Height           int
	DevicePixelRatio int
	Path             string
	Query            string
}

func copyStr(s string) string {
	ss := make([]byte, len(s))
	copy(ss, s)
	return string(ss)
}

func Record(req *http.Request, extra RecordExtra) (error, string) {
	r := req.Clone(context.Background())
	if !started {
		return ErrNotStarted, ""
	}
	if extra.DevicePixelRatio == 0 {
		extra.DevicePixelRatio = 1
	}
	if extra.HitID == "" {
		extra.HitID = GenHitID()
	}
	fmt.Println("Record", r.URL.String(), r.URL.EscapedPath())
	hit := Hit{
		ID:               extra.HitID,
		Path:             strings.Clone(r.URL.EscapedPath()),
		Query:            strings.Clone(r.URL.RawQuery),
		IP:               parseIP(r),
		Referer:          strings.Clone(r.Header.Get("Referer")),
		UserAgent:        strings.Clone(r.UserAgent()),
		Browser:          parseBrowser(r),
		OS:               parseOS(r),
		Title:            extra.Title,
		VisitorID:        extra.VisitorID,
		ResponseTime:     extra.ResponseTime,
		TimeOnPage:       extra.TimeOnPage,
		Width:            extra.Width,
		Height:           extra.Height,
		DevicePixelRatio: extra.DevicePixelRatio,
		IsBot:            isbot.Is(isbot.Bot(r)),
		CreatedAt:        time.Now().UnixMilli(),
	}
	hitQueue <- hit
	return nil, hit.ID
}

func Update(extra RecordExtra) error {
	if !started {
		return ErrNotStarted
	}
	if extra.HitID == "" {
		return nil
	}
	hit := Hit{
		ID:               extra.HitID,
		Title:            extra.Title,
		TimeOnPage:       extra.TimeOnPage,
		Width:            extra.Width,
		Height:           extra.Height,
		DevicePixelRatio: extra.DevicePixelRatio,
		Path:             extra.Path,
		Query:            extra.Query,
	}
	updateHitQueue <- hit
	return nil
}

func SqxSelect(dest interface{}, sql string, args ...interface{}) error {
	return pool.Select(dest, sql, args...)
}

func summarizeJob() {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1).Truncate(dayDuration)
	timer := time.NewTicker(tomorrow.Sub(now))
	for {
		select {
		case today := <-timer.C:
			SummarizeTables(today)
		case <-ctx.Done():
			return
		}
	}
}

func SummarizeTables(date time.Time) {
	pool.Exec(`
		INSERT INTO hits_summary (path, query, title, response_time, time_on_page, is_bot, num, date)
		SELECT path, query, title, cast(avg(response_time) as int), cast(avg(time_on_page) as int), is_bot, count(*) c, unixepoch(date(created_at/1000, 'unixepoch')) the_date
		FROM hits
		WHERE unixepoch(date(created_at/1000, 'unixepoch')) = ?
		GROUP BY path, query, title, is_bot, the_date
`, date.Unix())

	minCount := 10
	cols := []string{"referer", "browser", "os"}
	for _, col := range cols {
		sql := fmt.Sprintf(`
			INSERT INTO named_counts (type, name, num, date)
			SELECT '%s', %s, count(*) c, unixepoch(date(created_at/1000, 'unixepoch')) the_date
			FROM hits
			WHERE unixepoch(date(created_at/1000, 'unixepoch')) = ?
			GROUP BY %s
			HAVING c > %d
	`, col, col, col, minCount)
		fmt.Println("SLQ", sql)
		pool.Exec(sql, date.Unix())
	}
	sql := fmt.Sprintf(`
		INSERT INTO named_counts (type, name, num, date)
		SELECT 'WxH', wh, count(*) c, the_date
		FROM (SELECT width || 'x' || height as wh, unixepoch(date(created_at/1000, 'unixepoch')) the_date from hits
			WHERE unixepoch(date(created_at/1000, 'unixepoch')) = ? AND wh IS NOT NULL
		)
		GROUP BY wh
		HAVING c > %d
`, minCount)
	fmt.Println("sql", sql)
	pool.Exec(sql, date.Unix())
}
