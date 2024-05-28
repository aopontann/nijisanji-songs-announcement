package nsa

import (
	"os"
	"testing"
)

func TestPlaylists(t *testing.T) {
	youtubeApiKey := os.Getenv("YOUTUBE_API_KEY")
	yt := NewYoutube(youtubeApiKey)

	pidList := []string{"UUe_p3YEuYJb8Np0Ip9dk-FQ", "UUveZ9Ic1VtcXbsyaBgxPMvg"}
	playlists, err := yt.Playlists(pidList)
	if err != nil {
		t.Error(err)
	}

	if playlists["UUe_p3YEuYJb8Np0Ip9dk-FQ"] != 1 {
		t.Errorf("except 1, but %d", playlists["UUe_p3YEuYJb8Np0Ip9dk-FQ"])
	}

	if playlists["UUveZ9Ic1VtcXbsyaBgxPMvg"] != 18 {
		t.Errorf("except 18, but %d", playlists["UUveZ9Ic1VtcXbsyaBgxPMvg"])
	}
}

func TestPlaylistItems(t *testing.T) {
	youtubeApiKey := os.Getenv("YOUTUBE_API_KEY")
	yt := NewYoutube(youtubeApiKey)

	pidList := []string{"UUe_p3YEuYJb8Np0Ip9dk-FQ", "UUveZ9Ic1VtcXbsyaBgxPMvg"}
	vidList, err := yt.PlaylistItems(pidList)
	if err != nil {
		t.Error(err)
	}

	if len(vidList) != 4 {
		t.Errorf("except 4, but %d", len(vidList))
	}
}

func TestVideos(t *testing.T) {
	youtubeApiKey := os.Getenv("YOUTUBE_API_KEY")
	yt := NewYoutube(youtubeApiKey)

	vidList := []string{"EsP19D0ruk0", "Ff4YK5clkZ4"}
	videos, err := yt.Videos(vidList)
	if err != nil {
		t.Error(err)
	}

	if len(videos) != 2 {
		t.Errorf("except 2, but %d", len(vidList))
	}
}
