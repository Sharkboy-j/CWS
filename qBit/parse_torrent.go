package qBit

import (
	"bytes"
	"crypto/sha1"
	"cws/logger"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/autobrr/go-qbittorrent"
	"github.com/jackpal/bencode-go"
)

// TorrentInfo содержит информацию о торренте
type TorrentInfo struct {
	InfoHash string
	Name     string
}

// ParseTorrentFile парсит торрент файл и извлекает info hash и название
func ParseTorrentFile(torrentData []byte) (*TorrentInfo, error) {
	logger.Debug("Парсинг торрент файла для извлечения info hash и названия")

	reader := bytes.NewReader(torrentData)

	// Используем Decode вместо Unmarshal для более надежного парсинга
	decoded, err := bencode.Decode(reader)
	if err != nil {
		logger.Error("Ошибка при декодировании торрент файла: %v", err)

		return nil, fmt.Errorf("failed to decode torrent file: %w", err)
	}

	// Приводим к map
	torrentDataMap, ok := decoded.(map[string]interface{})
	if !ok {

		return nil, fmt.Errorf("torrent file is not a dictionary")
	}

	// Извлекаем info словарь
	infoRaw, ok := torrentDataMap["info"]
	if !ok {

		return nil, fmt.Errorf("info dictionary not found in torrent file")
	}

	info, ok := infoRaw.(map[string]interface{})
	if !ok {

		return nil, fmt.Errorf("info is not a dictionary")
	}

	// Вычисляем info hash (SHA-1 хеш от bencoded info словаря)
	// Повторно кодируем info словарь в bencode формат
	var infoBencodedBuffer bytes.Buffer
	err = bencode.Marshal(&infoBencodedBuffer, info)
	if err != nil {
		logger.Error("Ошибка при кодировании info словаря: %v", err)

		return nil, fmt.Errorf("failed to marshal info dictionary: %w", err)
	}

	infoBencoded := infoBencodedBuffer.Bytes()
	hash := sha1.Sum(infoBencoded)
	infoHash := hex.EncodeToString(hash[:])

	// Извлекаем название
	name, ok := info["name"].(string)
	if !ok {
		name = ""
	}

	logger.Debug("Извлечено из торрент файла: hash=%s, name=%s", infoHash, name)

	return &TorrentInfo{
		InfoHash: infoHash,
		Name:     name,
	}, nil
}

// FindTorrentByHash ищет торрент по хешу в списке торрентов
func FindTorrentByName(torrents []qbittorrent.Torrent, name string) *qbittorrent.Torrent {
	for _, torrent := range torrents {
		if strings.EqualFold(torrent.Name, name) {

			return &torrent
		}
	}

	return nil
}
