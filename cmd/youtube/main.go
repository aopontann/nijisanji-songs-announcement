package youtube

import (
	"context"
	"os"

	"github.com/aopontann/nijisanji-songs-announcement/db"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Youtube struct {
	service *youtube.Service
	queries *db.Queries
}

func New(db *db.Queries) (*Youtube, error) {
	ctx := context.Background()
	s, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		log.Fatal().Err(err).Msg("youtube.NewService create failed")
		return nil, err
	}

	return &Youtube{
		service: s,
		queries: db,
	}, nil
}
