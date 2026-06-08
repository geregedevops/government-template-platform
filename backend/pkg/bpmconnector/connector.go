// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package bpmconnector нь BPM-ийн serviceTask-аас гадаад REST API руу хийх
// гарах HTTP дуудлагыг хариуцна. URL нь хэрэглэгчийн тохируулсан утга тул
// **SSRF (Server-Side Request Forgery) хамгаалалт** заавал: хувийн / loopback /
// link-local / cloud-metadata IP рүү холбогдохыг хориглоно. Шалгалтыг
// net.Dialer.Control дотор хийдэг тул нэр→IP хувирсны ДАРАА (DNS rebinding-ийг
// хамгаалж) шалгагдана.
package bpmconnector

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"
)

// ErrBlockedHost нь хувийн/loopback зэрэг хориотой хаяг руу холбогдохыг оролдоход.
var ErrBlockedHost = errors.New("blocked host: private, loopback or link-local addresses are not allowed")

// Client нь SSRF-хамгаалалттай HTTP гүйцэтгэгч.
type Client struct {
	http     *http.Client
	maxBytes int64
}

// New нь timeout болон хариуны дээд хэмжээтэй connector үүсгэнэ.
func New(timeout time.Duration, maxBytes int64) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if maxBytes <= 0 {
		maxBytes = 1 << 20 // 1 MiB
	}
	dialer := &net.Dialer{
		Timeout: timeout,
		// Control нь resolve хийсэн бодит сокет хаяг дээр ажилладаг тул DNS
		// rebinding-ийг хүртэл барина (redirect бүрийн шинэ dial мөн шалгагдана).
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if ip == nil || isBlocked(ip) {
				return ErrBlockedHost
			}
			return nil
		},
	}
	return &Client{
		http: &http.Client{
			Timeout: timeout,
			// Proxy: nil — env proxy-аар SSRF-ийг тойрохоос сэргийлж proxy
			// ашиглахгүй.
			Transport: &http.Transport{DialContext: dialer.DialContext, Proxy: nil},
			CheckRedirect: func(_ *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return errors.New("too many redirects")
				}
				return nil
			},
		},
		maxBytes: maxBytes,
	}
}

// isBlocked нь IP нь гадаад руу чиглээгүй (дотоод/нөөц) эсэхийг шалгана.
func isBlocked(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() || // IPv4 private + IPv6 ULA (fc00::/7)
		ip.IsLinkLocalUnicast() || // 169.254.0.0/16 (cloud metadata) + fe80::/10
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast()
}

// Do нь HTTP дуудлага хийж, статус код + хариуны бие (дээд хэмжээгээр
// тайрсан)-г буцаана. Зөвхөн http/https зөвшөөрнө.
func (c *Client) Do(ctx context.Context, method, rawURL string, headers map[string]string, body string) (int, []byte, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodGet
	}
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return 0, nil, fmt.Errorf("only http(s) urls are allowed")
	}

	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, rdr)
	if err != nil {
		return 0, nil, err
	}
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, c.maxBytes))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, data, nil
}
