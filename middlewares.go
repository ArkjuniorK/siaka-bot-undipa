package main

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"regexp"
)

func Logger(logger *log.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			re := regexp.MustCompile("[0-9]-[A-Za-z0-9]")
			if update.Message != nil && !re.MatchString(update.Message.Text) {
				msg := update.Message
				logger.Printf("%s [%d] | msg : %s", msg.From.Username, msg.From.ID, msg.Text)
			}

			next(ctx, b, update)
		}
	}
}
