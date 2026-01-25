package celeritas

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/ansufw/celeritas/cache"
	"github.com/ansufw/celeritas/mailer"
	"github.com/ansufw/celeritas/render"
	"github.com/ansufw/celeritas/session"
	"github.com/dgraph-io/badger/v4"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const version = "1.0.0"

var myRedisCache *cache.RedisCache
var myBadgerCache *cache.Badger
var redisPool *redis.Pool
var badgerConn *badger.DB

type Celeritas struct {
	AppName       string
	Debug         bool
	Version       string
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	RootPath      string
	Routes        *chi.Mux
	Render        *render.Render
	Session       *scs.SessionManager
	DB            Database
	JetView       *jet.Set
	config        config
	EncryptionKey string
	Cache         cache.Cache
	Scheduler     *cron.Cron
	Mail          *mailer.Mail
	Server        *Server
}

type Server struct {
	ServerName string
	Port       string
	Secure     bool
	URL        string
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
	redis       redisConfig
}

func (c *Celeritas) New(rootPath string) error {
	c.RootPath = rootPath
	c.Version = version

	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "mail", "data", "public", "tmp", "logs", "middleware"},
	}

	if err := c.Init(pathConfig); err != nil {
		return err
	}

	err := c.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	// read .env file
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// create logger
	infoLog, errorLog := c.startLoggers()
	c.InfoLog = infoLog
	c.ErrorLog = errorLog

	c.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))

	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := c.OpenDB(os.Getenv("DATABASE_TYPE"), c.BuildDSN())
		if err != nil {
			errorLog.Printf("Failed to open database | error: %v | dsn: %v\n", err, c.BuildDSN())
			os.Exit(1)
		}
		c.DB = Database{
			DataType: os.Getenv("DATABASE_TYPE"),
			Pool:     db,
		}
	}

	c.Scheduler = cron.New()

	// connect to redis
	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		myRedisCache = c.createClientRedisCache()
		c.Cache = myRedisCache
		redisPool = myRedisCache.Conn
	}

	if os.Getenv("CACHE") == "badger" {
		myBadgerCache = c.createClientBadgerCache()
		c.Cache = myBadgerCache
		badgerConn = myBadgerCache.Conn

		_, err := c.Scheduler.AddFunc("@daily", func() {
			_ = myBadgerCache.Conn.RunValueLogGC(0.7)
		})
		if err != nil {
			return err
		}
	}

	c.Routes = c.routes()
	c.Mail = c.createMailer()

	c.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSIST"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      c.BuildDSN(),
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	secure := true
	if strings.ToLower(os.Getenv("SECURE")) == "false" {
		secure = false
	}

	c.Server = &Server{
		ServerName: os.Getenv("SERVER_NAME"),
		Port:       os.Getenv("PORT"),
		Secure:     secure,
		URL:        os.Getenv("APP_URL"),
	}

	// create session
	sess := &session.Session{
		CookieLifetime: c.config.cookie.lifetime,
		CookiePersist:  c.config.cookie.persist,
		CookieName:     c.config.cookie.name,
		CookieDomain:   c.config.cookie.domain,
		SessionType:    c.config.sessionType,
		CookieSecure:   c.config.cookie.secure,
		// DBPool:         c.DB.Pool,
	}

	switch c.config.sessionType {
	case "redis":
		sess.RedisPool = myRedisCache.Conn
	case "mysql", "postgres", "mariadb", "postgresql":
		sess.DBPool = c.DB.Pool
	}

	c.Session = sess.Init()
	c.EncryptionKey = os.Getenv("KEY")

	if c.Debug {
		c.JetView = jet.NewSet(
			jet.NewOSFileSystemLoader(c.RootPath+"/views"),
			jet.InDevelopmentMode(),
		)
	} else {
		c.JetView = jet.NewSet(
			jet.NewOSFileSystemLoader(c.RootPath + "/views"),
		)
	}

	c.JetView = jet.NewSet(
		jet.NewOSFileSystemLoader(c.RootPath+"/views"),
		jet.InDevelopmentMode(),
	)

	c.createRenderer()
	go c.Mail.ListenForMail()

	return nil
}

func (c *Celeritas) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		// create folder if not exists
		err := c.createDirIfNotExists(root + "/" + path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Celeritas) ListenAndServe() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", c.config.port),
		ErrorLog:     c.ErrorLog,
		Handler:      c.Routes,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	if c.DB.Pool != nil {
		defer c.DB.Pool.Close()
	}

	if redisPool != nil {
		defer redisPool.Close()
	}

	if badgerConn != nil {
		defer badgerConn.Close()
	}

	c.InfoLog.Println("Starting server on port", c.config.port)
	return srv.ListenAndServe()
}

func (c *Celeritas) checkDotEnv(rootPath string) error {
	err := c.createFileIfNotExists(fmt.Sprintf("%s/.env", rootPath))
	if err != nil {
		return err
	}

	return nil
}

func (c *Celeritas) startLoggers() (*log.Logger, *log.Logger) {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return infoLog, errorLog
}

func (c *Celeritas) createRenderer() {
	c.Render = &render.Render{
		Renderer: c.config.renderer,
		RootPath: c.RootPath,
		Port:     c.config.port,
		JetViews: c.JetView,
		Session:  c.Session,
	}
}

func (c *Celeritas) createMailer() *mailer.Mail {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	return &mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Templates:   c.RootPath + "/mail",
		Port:        port,
		Host:        os.Getenv("SMTP_HOST"),
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromName:    os.Getenv("SMTP_FROM_NAME"),
		FromAddress: os.Getenv("SMTP_FROM_ADDRESS"),
		Jobs:        make(chan mailer.Message, 20),
		Results:     make(chan mailer.Result, 20),
		API:         os.Getenv("MAILER_API"),
		APIKey:      os.Getenv("MAILER_KEY"),
		APIUrl:      os.Getenv("MAILER_URL"),
	}

}

func (c *Celeritas) createClientRedisCache() *cache.RedisCache {
	return &cache.RedisCache{
		Conn:   c.createRedisPool(),
		Prefix: c.config.redis.prefix,
	}
}

func (c *Celeritas) createClientBadgerCache() *cache.Badger {
	return &cache.Badger{
		Conn:   c.createBadgerConn(),
		Prefix: c.config.redis.prefix,
	}
}

func (c *Celeritas) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10_000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				c.config.redis.host,
				redis.DialPassword(c.config.redis.password))
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}
}

func (c *Celeritas) createBadgerConn() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(c.RootPath + "/tmp/badger"))
	if err != nil {
		c.ErrorLog.Printf("Failed to open badger database | error: %v | dsn: %v\n", err, c.RootPath+"/tmp/badger")
		os.Exit(1)
	}
	return db
}

func (c *Celeritas) BuildDSN() string {

	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_PASS"), os.Getenv("DATABASE_NAME"), os.Getenv("DATABASE_SSL_MODE"))
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
			os.Getenv("DATABASE_USER"), os.Getenv("DATABASE_PASS"), os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"), os.Getenv("DATABASE_NAME"))
	}

	return dsn
}
