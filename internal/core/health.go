package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// CheckResult - что получилось при проверке доступности дискорда.
type CheckResult struct {
	SiteOK  bool
	CDNOK   bool
	Elapsed time.Duration
	Err     string // первая значимая ошибка, если что-то не так
}

func (r CheckResult) OK() bool { return r.SiteOK || r.CDNOK }

func (r CheckResult) String() string {
	mark := func(ok bool) string {
		if ok {
			return "ок"
		}
		return "нет"
	}
	s := fmt.Sprintf("сайт: %s, картинки: %s (%.1f с)", mark(r.SiteOK), mark(r.CDNOK), r.Elapsed.Seconds())
	if r.Err != "" && !r.OK() {
		s += " - " + r.Err
	}
	return s
}

// probe GET-ит один url, считает доступным любой ответ < 500.
func probe(client *http.Client, url string) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500, nil
}

// CheckDiscord стучится в два адреса дискорда.
// Ничего не гарантирует про голос - UDP-звонки проверяются уже в самом клиенте.
func CheckDiscord(timeout time.Duration) CheckResult {
	start := time.Now()
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}
	res := CheckResult{}

	site, siteErr := probe(client, "https://discord.com/api/v10/gateway")
	cdn, cdnErr := probe(client, "https://cdn.discordapp.com/embed/avatars/0.png")
	res.SiteOK, res.CDNOK = site, cdn
	res.Elapsed = time.Since(start)

	if !res.OK() {
		switch {
		case siteErr != nil:
			res.Err = fmt.Sprint(siteErr)
		case cdnErr != nil:
			res.Err = fmt.Sprint(cdnErr)
		default:
			res.Err = "оба адреса ответили ошибкой"
		}
	}
	return res
}

// WaitReachable ждёт доступности до maxWait, дёргая CheckDiscord каждые 700 мс.
func WaitReachable(ctx context.Context, maxWait time.Duration) (CheckResult, bool) {
	deadline := time.Now().Add(maxWait)
	for {
		r := CheckDiscord(2500 * time.Millisecond)
		if r.OK() {
			return r, true
		}
		if time.Now().After(deadline) {
			return r, false
		}
		select {
		case <-ctx.Done():
			return r, false
		case <-time.After(700 * time.Millisecond):
		}
	}
}
