package qbit

import (
	"context"
	"crypto/tls"
	"cws/internal/storage"
	"cws/logger"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent"
)

type TransferInfo struct {
	DownloadSpeed int64
	UploadSpeed   int64
	DownloadLimit int64
	UploadLimit   int64
}

type Service interface {
	AddTorrentFile(ctx context.Context, torrentFile []byte, savePath string, category string, skipHashCheck bool) error
	DeleteTorrent(ctx context.Context, hash string, deleteFiles bool) error
	GetTorrents(ctx context.Context) ([]qbittorrent.Torrent, error)
	GetTorrentPropertiesCached(ctx context.Context, hash string) (*qbittorrent.TorrentProperties, error)
	GetDefaultSavePath(ctx context.Context) (string, error)
	GetTorrentSavePaths(ctx context.Context) ([]string, error)
	FilterTorrentsByRutrackerComment(ctx context.Context, torrents []qbittorrent.Torrent) ([]qbittorrent.Torrent, error)
	GetTorrentsCtx(ctx context.Context, options qbittorrent.TorrentFilterOptions) ([]qbittorrent.Torrent, error)
	GetTorrentPropertiesCtx(ctx context.Context, hash string) (qbittorrent.TorrentProperties, error)
	PauseAllTorrents(ctx context.Context) error
	ResumeAllTorrents(ctx context.Context) error
	PauseTorrent(ctx context.Context, hash string) error
	ResumeTorrent(ctx context.Context, hash string) error
	SetGlobalSpeedLimits(ctx context.Context, downloadLimit, uploadLimit int64) error
	GetTransferInfo(ctx context.Context) (*TransferInfo, error)
}

type service struct {
	client     *qbittorrent.Client
	baseURL    string
	httpClient *http.Client
}

func New(ctx context.Context, client *storage.Client) (Service, error) {
	var baseURL string
	if client.SSL {
		if client.Port == 443 {
			baseURL = fmt.Sprintf("https://%s", client.Host)
		} else {
			baseURL = fmt.Sprintf("https://%s:%d", client.Host, client.Port)
		}
	} else {
		if client.Port == 80 {
			baseURL = fmt.Sprintf("http://%s", client.Host)
		} else {
			baseURL = fmt.Sprintf("http://%s:%d", client.Host, client.Port)
		}
	}

	logger.Debug("Подключение к qBit клиенту: %s (пользователь: %s)", baseURL, client.Username)

	loginURL := fmt.Sprintf("%s/api/v2/auth/login", baseURL)

	jar, err := cookiejar.New(nil)
	if err != nil {
		logger.Error("Ошибка при создании cookie jar: %v", err)

		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	data := url.Values{}
	data.Set("username", client.Username)
	data.Set("password", client.Password)

	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error("Ошибка при создании запроса на логин: %v", err)

		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("Referer", baseURL+"/")
	req.Header.Set("Origin", baseURL)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:146.0) Gecko/20100101 Firefox/146.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса на логин: %v", err)

		return nil, fmt.Errorf("failed to execute login request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Ошибка при чтении ответа: %v", err)

		return nil, fmt.Errorf("failed to read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Ошибка аутентификации: статус %d, ответ: %s", resp.StatusCode, string(body))

		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	responseText := strings.TrimSpace(string(body))
	if responseText == "Fails." || strings.Contains(responseText, "Fails") {
		logger.Error("Неверные учетные данные: %s", responseText)

		return nil, fmt.Errorf("invalid credentials")
	}

	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "SID" {
			sessionCookie = cookie

			break
		}
	}

	if sessionCookie == nil {
		logger.Warn("Cookie сессии не найден, но логин успешен. Продолжаем...")
	}

	logger.Debug("Успешная аутентификация в qBit клиенту %s", baseURL)

	cfg := qbittorrent.Config{
		Host:          baseURL,
		Username:      client.Username,
		Password:      client.Password,
		TLSSkipVerify: true,
	}

	qbClient := qbittorrent.NewClient(cfg)

	logger.Debug("Успешное подключение к qBit клиенту: %s (HTTP логин выполнен)", baseURL)

	return &service{
		client:     qbClient,
		baseURL:    baseURL,
		httpClient: httpClient,
	}, nil
}
