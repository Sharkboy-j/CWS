package qbit

import "context"

func (s *service) ResumeTorrents(ctx context.Context, hashes []string) error {
	return s.torrentsAction(ctx, "/api/v2/torrents/start", hashes, "resume")
}
