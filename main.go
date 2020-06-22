package main

import (
    "os"
    "flag"
    "net/http"
    "fmt"
    "time"
    "database/sql"

    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/common/log"

    _ "github.com/godror/godror"
)

var (
    version string = "1.0-dev"
    listenAddress       = flag.String("listen-address", getEnv("LISTEN_ADDRESS", ":9162"), "Address to listen on for web interface and telemetry. (env: LISTEN_ADDRESS)")
    metricPath          = flag.String("telemetry-path", getEnv("TELEMETRY_PATH", "/metrics"), "Path under which to expose metrics. (env: TELEMETRY_PATH)")
    db *sql.DB
    pingDataBase string
    dsn string
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func connect(dsn string) *sql.DB {
    db, err := sql.Open("godror", dsn)
    if err != nil {
        fmt.Println(err)
        panic(err)
    }
    
    db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(3)
    db.SetMaxOpenConns(3)
    
    return db
}

func CheckCntUsr(db *sql.DB, userCnt prometheus.Gauge) {
    rows2, err := db.Query("select count(*) from v$session")
    if err != nil {
        fmt.Println("Error running query rows2")
        fmt.Println(err)
        return
    }
    defer rows2.Close()
    var val int
    for rows2.Next() {
        rows2.Scan(&val)
    }
    fmt.Printf("The userCnt is: %d\n", val)
    userCnt.Set(float64(val))
}

// For example
func CheckAccStatus(db *sql.DB) {
    q := `select username, account_status from dba_users order by username`
    rows3, err := db.Query(q)
    if err != nil {
        fmt.Println("Error running query rows3")
        fmt.Println(err)
        return
    }
    defer rows3.Close()
    var user, status string 
    for rows3.Next() {
        rows3.Scan(&user, &status)
        fmt.Printf("The acc is: %s and status is: %s\n", user, status)
    }
}

func main() {
    dsn := os.Getenv("DATA_SOURCE_NAME")

    globalSleep := 5000 * time.Millisecond
    fmt.Println("Hello custom_exporter!")
    flag.Parse()

    oraDbUp := prometheus.NewGauge (
        prometheus.GaugeOpts {
            Name: "oradb_up",
        })
    prometheus.MustRegister(oraDbUp)

    userCnt := prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "oradb_user_cnt",
        })
    prometheus.MustRegister(userCnt)
        
    db := connect(dsn)

    rows, err := db.Query("select sysdate from dual")
    if err != nil {
        fmt.Println("Error running query")
        fmt.Println(err)
        return
    }
    defer rows.Close()

    var thedate string
    for rows.Next() {
        rows.Scan(&thedate)
    }

    go func() {
        var err error
        for {
            if err = db.Ping(); err != nil {
                log.Infoln("Reconnecting to DB")
                log.Infoln(err)
                db = connect(dsn)
            }

            if err = db.Ping(); err != nil {
                pingDataBase = "Not Ok"
                log.Infoln(pingDataBase)
                oraDbUp.Set(0)
            } else {
                pingDataBase = "Ok"
                log.Infoln(pingDataBase)
                oraDbUp.Set(1)
            }

            CheckCntUsr(db, userCnt)
            time.Sleep(globalSleep)
        }
    } ()

    fmt.Printf("The date is: %s\n", thedate)
    log.Infoln("Starting Oracle Data Base Exporter on " + dsn + ". Version " + version)
    CheckAccStatus(db)

    http.Handle(*metricPath, promhttp.Handler())
    log.Infoln("Listening on", *listenAddress)
    log.Fatal(http.ListenAndServe(*listenAddress, nil))
}