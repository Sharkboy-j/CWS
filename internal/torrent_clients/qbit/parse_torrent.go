package qbit

import (
	"bytes"
	"crypto/sha1"
	"cws/logger"
	"encoding/hex"
	"fmt"

	"github.com/jackpal/bencode-go"
)

type TorrentInfo struct {
	InfoHash string
	Name     string
}

func ParseTorrentFile(torrentData []byte) (*TorrentInfo, error) {
	logger.Debug("Парсинг торрент файла для извлечения info hash и названия")

	reader := bytes.NewReader(torrentData)

	decoded, err := bencode.Decode(reader)
	if err != nil {
		logger.Error("Ошибка при декодировании торрент файла: %v", err)

		return nil, fmt.Errorf("failed to decode torrent file: %w", err)
	}

	torrentDataMap, ok := decoded.(map[string]interface{})
	if !ok {

		return nil, fmt.Errorf("torrent file is not a dictionary")
	}

	infoRaw, ok := torrentDataMap["info"]
	if !ok {

		return nil, fmt.Errorf("info dictionary not found in torrent file")
	}

	info, ok := infoRaw.(map[string]interface{})
	if !ok {

		return nil, fmt.Errorf("info is not a dictionary")
	}

	var infoBencodedBuffer bytes.Buffer
	err = bencode.Marshal(&infoBencodedBuffer, info)
	if err != nil {
		logger.Error("Ошибка при кодировании info словаря: %v", err)

		return nil, fmt.Errorf("failed to marshal info dictionary: %w", err)
	}

	infoBencoded := infoBencodedBuffer.Bytes()
	hash := sha1.Sum(infoBencoded)
	infoHash := hex.EncodeToString(hash[:])

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
