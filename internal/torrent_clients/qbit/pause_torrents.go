package qbit

import "context"

func (s *service) PauseTorrents(ctx context.Context, hashes []string) error {
	return s.torrentsAction(ctx, "/api/v2/torrents/stop", hashes, "pause")
}
