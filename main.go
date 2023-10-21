package main

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"time"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "siaka-undipa-bot | ", log.Lshortfile|log.Ldate|log.Ltime)
	logger.Println("Initializing dependencies, please wait...")

	// Load env file
	err := godotenv.Load(".env")
	if err != nil {
		logger.Fatalln(err)
	}

	// Setup context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Initialize redis cache and check its status, if it does not active then fatal
	r := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
	})

	pong, err := r.Ping(ctx).Result()
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Println("Redis cache initialized, msg:", pong)

	// Initialize colly package and set its configs
	c := colly.NewCollector(colly.AllowedDomains(
		"siaka."+DOMAIN,
		"apidiv."+DOMAIN,
	), colly.AllowURLRevisit())
	c.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Minute,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:           100,
		ReadBufferSize:         1024 * 100,
		IdleConnTimeout:        10 * time.Minute,
		TLSHandshakeTimeout:    10 * time.Minute,
		ExpectContinueTimeout:  10 * time.Minute,
		ResponseHeaderTimeout:  60 * time.Minute,
		MaxResponseHeaderBytes: 1024 * 100,
	})
	c.SetRequestTimeout(10 * time.Minute)

	logger.Println("Collector initialized")

	// Initialize handler
	h := NewHandler(c, r, logger)
	logger.Println("Handler initialized")

	// Initialize telegram bot
	opts := []bot.Option{
		bot.WithDefaultHandler(h.Default),
		bot.WithMiddlewares(Logger(logger)),
		bot.WithCheckInitTimeout(1 * time.Hour),
		bot.WithSkipGetMe(),
	}

	token := os.Getenv("TG_BOT_TOKEN")
	b, err := bot.New(token, opts...)
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Println("Telegram bot initialized")

	// Register handlers to telegram bot
	b.RegisterHandler(bot.HandlerTypeMessageText, "/login", bot.MatchTypeExact, h.Login)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/logout", bot.MatchTypeExact, h.Logout)

	b.RegisterHandler(bot.HandlerTypeMessageText, "/bpp", bot.MatchTypeExact, h.BPP)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/divlearn", bot.MatchTypeExact, h.APIDiv)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/schedule", bot.MatchTypeExact, h.Schedule)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/lecturer", bot.MatchTypeExact, h.Lecturer)

	stbRe := regexp.MustCompile("[0-9]-[A-Za-z0-9]")
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, stbRe, h.Login)

	logger.Println("Handler registered")

	logger.Println("Bot server start")
	b.Start(ctx)
}
