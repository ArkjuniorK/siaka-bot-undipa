package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/paginator"
	"github.com/gocolly/colly"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	c *colly.Collector
	r *redis.Client
	l *log.Logger
}

func NewHandler(c *colly.Collector, r *redis.Client, l *log.Logger) *Handler {
	return &Handler{c, r, l}
}

// Default would handle /start command to greet the users
func (h *Handler) Default(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message.Text
	if msg != "/start" {
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Selamat datang, silahkan masuk untuk memulai!",
		ChatID: update.Message.Chat.ID,
	})

	return

}

// Login handle the /login and stb-pass command. This handler would save the
// cookie to cache as JSON if user successfully logged in.
func (h *Handler) Login(ctx context.Context, b *bot.Bot, update *models.Update) {
	csStr, _ := h.r.Get(ctx, strconv.Itoa(int(update.Message.From.ID))).Result()
	if len(csStr) != 0 {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			Text:   "Anda telah masuk!",
			ChatID: update.Message.Chat.ID,
		})

		return
	}

	msg := update.Message.Text
	if msg == "/login" {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			Text:   "Silahkan masuk dengan mengetikkan Stambuk dan Kata Sandi. Adapun format yang di gunakan yaitu :\n\nstambuk-kata sandi\n\nTerimakasih.",
			ChatID: update.Message.Chat.ID,
		})

		return
	}

	data := strings.Split(msg, "-")
	payload := map[string][]byte{
		"email": []byte(data[0]),
		"pasw":  []byte(data[1]),
		"juser": []byte("MHS"),
	}

	if err := h.c.PostMultipart(SIAKA+"?page=login&aksi=masuk", payload); err != nil {
		h.l.Println(err)
		h.handlerError(ctx, b, update)
		return
	}

	h.c.OnResponse(func(rs *colly.Response) {
		cs := h.c.Cookies(rs.Request.URL.String())
		csj, err := json.Marshal(cs)
		if err != nil {
			h.l.Println(err)
			h.handlerError(ctx, b, update)
			return
		}

		h.r.Set(ctx, strconv.Itoa(int(update.Message.From.ID)), csj, 40*time.Minute)
		h.r.Save(ctx)
	})

	if err := h.c.Visit(SIAKA); err != nil {
		panic(err)
	}

	text := "Berhasil masuk! Gunakan menu yang telah disediakan untuk mengakses SIAKA."
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   text,
		ChatID: update.Message.Chat.ID,
	})

	_, _ = b.DeleteMessage(ctx, &bot.DeleteMessageParams{ChatID: update.Message.Chat.ID, MessageID: update.Message.ID})
	return
}

func (h *Handler) Status(ctx context.Context, b *bot.Bot, update *models.Update) {
	cookies, _ := h.r.Get(ctx, strconv.Itoa(int(update.Message.From.ID))).Result()
	if len(cookies) != 0 {
		_, _ =
			b.SendMessage(ctx, &bot.SendMessageParams{
				Text:   "Anda telah masuk!",
				ChatID: update.Message.Chat.ID,
			})

		return
	}
	_, _ =
		b.SendMessage(ctx, &bot.SendMessageParams{
			Text:   "Anda belum masuk kedalam sistem!",
			ChatID: update.Message.Chat.ID,
		})

	return
}

// Logout would remove the user's data from cache.
func (h *Handler) Logout(ctx context.Context, b *bot.Bot, update *models.Update) {
	var (
		key    = strconv.Itoa(int(update.Message.From.ID))
		keyBPP = key + "bpp"
		keySch = key + "schedule"
	)

	cookies := h.getCookie(ctx, b, update)
	if cookies == nil {
		return
	}

	h.r.Del(ctx, key)
	h.r.Del(ctx, keyBPP)
	h.r.Del(ctx, keySch)

	msg := "Anda telah keluar dari SIAKA Universitas Dipa Makassar\nTerima kasih."
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   msg,
		ChatID: update.Message.Chat.ID,
	})
}

func (h *Handler) BPP(ctx context.Context, b *bot.Bot, update *models.Update) {
	var (
		url = SIAKA + "?page=vbpp"
		key = strconv.Itoa(int(update.Message.From.ID)) + "bpp"
	)

	cookies := h.getCookie(ctx, b, update)
	if cookies == nil {
		return
	}

	sch, _ := h.r.Get(ctx, key).Result()
	if sch != "" {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprint(sch),
		})

		return
	}

	if err := h.c.SetCookies(url, cookies); err != nil {
		h.l.Println(err)
		h.handlerError(ctx, b, update)
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Memproses data BPP, mohon tunggu...",
	})

	bpp := make([][]string, 0)
	h.c.OnHTML("table.table-common:nth-child(1)", func(e *colly.HTMLElement) {
		if e.Index > 0 {
			return
		}

		e.ForEach("tr", func(i int, el *colly.HTMLElement) {
			tr := make([]string, 3)

			for j := 0; j <= 2; j++ {
				tr[j] = el.ChildText(fmt.Sprintf("td:nth-child(%d)", j+2))
			}

			bpp = append(bpp, tr)
		})
	})

	if err := h.c.Visit(url); err != nil {
		h.l.Println(err)
		h.handlerError(ctx, b, update)
		return
	}

	h.c.Wait()

	bppRs := "Riwayat pembayaran BPP:"
	for _, s := range bpp {
		if s[0] == "" {
			continue
		}

		txt := fmt.Sprintf("\n\nSemester: %s\nTanggal Bayar: %s\nBPP (Rp.): %s", s[0], s[1], s[2])
		bppRs += txt
	}

	h.r.Set(ctx, key, bppRs, 30*time.Minute)
	h.r.Save(ctx)

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprint(bppRs),
	})
}

func (h *Handler) Schedule(ctx context.Context, b *bot.Bot, update *models.Update) {
	var (
		url = SIAKA + "?page=kelas"
		key = strconv.Itoa(int(update.Message.From.ID)) + "schedule"
	)

	cookies := h.getCookie(ctx, b, update)
	if cookies == nil {
		return
	}

	sch, _ := h.r.Get(ctx, key).Result()
	if sch != "" {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprint(sch),
		})
		return
	}

	if err := h.c.SetCookies(url, cookies); err != nil {
		h.l.Println(err)
		h.handlerError(ctx, b, update)
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Memproses data mata kuliah, mohon tunggu...",
	})

	schedule := make([][]string, 0)
	h.c.OnHTML("table.table-common", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(i int, el *colly.HTMLElement) {
			tr := make([]string, 6)

			for j := 0; j <= 5; j++ {
				tr[j] = el.ChildText(fmt.Sprintf("td:nth-child(%d)", j+2))
			}

			schedule = append(schedule, tr)
		})
	})

	if err := h.c.Visit(url); err != nil {
		h.l.Println(err)
		h.handlerError(ctx, b, update)
		return
	}

	h.c.Wait()

	schRs := "Jadwal mata kuliah:"
	for _, s := range schedule {
		if s[0] == "" {
			continue
		}

		txt := fmt.Sprintf("\n\nMata Kuliah: %s\nKelas: %s\nHari: %s\nJam: %s\nRuang: %s\nDosen: %s", s[0], s[1], s[2], s[3], s[4], s[5])
		schRs += txt
	}

	h.r.Set(ctx, key, schRs, 30*time.Minute)
	h.r.Save(ctx)

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprint(schRs),
	})

	return

}

func (h *Handler) Lecturer(ctx context.Context, b *bot.Bot, update *models.Update) {
	var (
		url  = SIAKA + "?page=vdosen"
		key  = "lecturer"
		opts = []paginator.Option{paginator.PerPage(5), paginator.WithCloseButton("Tutup")}
	)

	cookies := h.getCookie(ctx, b, update)
	if cookies == nil {
		return
	}

	lecs, _ := h.r.Get(ctx, key).Result()
	if lecs != "" {
		lecturers := strings.Split(lecs, "~")
		p := paginator.New(lecturers, opts...)
		_, _ = p.Show(ctx, b, update.Message.Chat.ID)
		return
	}

	if err := h.c.SetCookies(url, cookies); err != nil {
		h.l.Println(err)
		h.handlerError(ctx, b, update)
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Memproses data dosen, mohon tunggu...",
	})

	data := make([][]string, 0)
	h.c.OnHTML("tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(i int, el *colly.HTMLElement) {
			tr := make([]string, 4)

			for j := 0; j <= 3; j++ {
				tr[j] = el.ChildText(fmt.Sprintf("td:nth-child(%d)", j+1))
			}

			data = append(data, tr)
		})
	})

	if err := h.c.Visit(url); err != nil {
		h.l.Println(err)
		h.handlerError(ctx, b, update)
		return
	}

	h.c.Wait()

	lecturers := make([]string, 0)
	for _, l := range data {
		phone := strings.Join(strings.Split(l[3], "-"), "")

		if len(phone) == 0 {
			phone = "Tidak ada"
		}

		d := fmt.Sprintf("NIDN: %s\nNama: %v\nTelepon: %s", l[0], l[1], phone)
		lecturers = append(lecturers, d)
	}

	h.r.Set(ctx, key, strings.Join(lecturers, "~"), 12*time.Hour)
	h.r.Save(ctx)

	p := paginator.New(lecturers, opts...)
	_, _ = p.Show(ctx, b, update.Message.Chat.ID)
}

func (h *Handler) APIDiv(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Coming soon...",
		ChatID: update.Message.Chat.ID,
	})
}

func (h *Handler) handlerError(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Terjadi kesalahan! Coba beberapa saat lagi.",
		ChatID: update.Message.Chat.ID,
	})
}

func (h *Handler) getCookie(ctx context.Context, b *bot.Bot, update *models.Update) []*http.Cookie {
	csStr, err := h.r.Get(ctx, strconv.Itoa(int(update.Message.From.ID))).Result()
	if err == redis.Nil && update.Message.Text != "/login" {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			Text:   "Sesi anda telah berakhir, silahkan masuk kembali untuk mengakses SIAKA!",
			ChatID: update.Message.Chat.ID,
		})

		return nil
	}

	if csStr != "" {
		var cookies []*http.Cookie
		if err = json.Unmarshal([]byte(csStr), &cookies); err != nil {
			h.l.Println(err)
			h.handlerError(ctx, b, update)
			return nil
		}

		return cookies
	}

	return nil
}
