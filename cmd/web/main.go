package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"snippetbox.net/internal/models"
)

type config struct {
	addr      string
	staticDir string
}

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static/", "Path to static assets")

	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	errorLog := log.New(os.Stderr, "ERROR\t", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.LUTC|log.Ldate|log.Ltime)

	flag.Parse()

	db, err := openDB(*dsn)

	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	num, err := strconv.Atoi(strings.SplitAfter(cfg.addr, ":")[1])
	if err != nil {
		errorLog.Fatal(err)
	}

	if num <= 1023 {
		errorLog.Print("Bind: permission denied")
		return
	}
	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db.DB)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		//tls.CurveP256 and tls.X25519 have assembly implementations
		//others are cpu intensive. Just performance imrpocment
	}
	srv := &http.Server{
		Addr:      cfg.addr,
		ErrorLog:  app.errorLog,
		Handler:   app.routes(),
		TLSConfig: tlsConfig,

		IdleTimeout: time.Minute,     // reduce keep-alive timeout, can't increase it
		ReadTimeout: 5 * time.Second, // Time which server will wait to read the message, important to avoid Slowloris
		// prevents slow response uploads
		WriteTimeout: 10 * time.Second, // to prevent the data that the handler returns from taking too long to write
	}

	app.infoLog.Printf("Starting a server on %s", cfg.addr)

	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

	app.errorLog.Fatal(err)
}

func openDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
