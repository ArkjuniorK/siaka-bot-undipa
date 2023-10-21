package main

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/gocolly/colly"
	"github.com/redis/go-redis/v9"
	"log"
)

type Handler struct {
	c *colly.Collector
	r *redis.Client
	l *log.Logger
}

func NewHandler(c *colly.Collector, r *redis.Client, l *log.Logger) *Handler {
	return &Handler{c, r, l}
}

func (h *Handler) Default() bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {

	}
}

func (h *Handler) Login() bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {

	}
}

func (h *Handler) Logout() bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {

	}
}

func (h *Handler) BPP() bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {

	}
}

func (h *Handler) Schedule() bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {

	}
}

func (h *Handler) Lecturer() bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {

	}
}

func (h *Handler) APIDiv() bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {

	}
}

//func (h *Handler) sendAlert(ctx context.Context, b *bot.Bot, update *models.Update) {
//	h.Logout(ctx, b, update)
//	b.SendMessage(ctx, &bot.SendMessageParams{
//		ChatID: update.Message.Chat.ID,
//		Text:   "Tidak dapat mengakses SIAKA, masuk terlebih dahulu!",
//	})
//}
