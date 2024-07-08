package nsa

import (
	"testing"
)

func TestSend(t *testing.T) {
	fcm := NewFCM()

	err := fcm.Notification(
		"5分後に公開",
		[]string{
			"cXPUH6Zl18wtzVCAPI7KaE:APA91bGW4wQ3-k2PgHbEeQzod54NVhK6hwIA9GdfZZyTDt5uE9NJTLSP_QonuDm808bNz8rGHdFCvXeB1_TB9CaEIXnlDAJ5Cu5OW0VyBNV8ezQLaGeDw-eqGVTPavSP2sKvWdfNFktI",
			"dzrpIshcgHp7FjA09VrFJG:APA91bHioff34_JW3LNXTEjhxNYILdGnGLSuzW54dF3QyzLJNQqD1jTbpoPFdboXLaIP0Oj4okMwyKfipEZri9puNxWGyVZbQSWPKnKi6TzHSJdHakcFOiHbBc1NLSVeT2PUzG9fqHAg",
			"fVzPqMEQgVNG4ZKCUrf6J-:APA91bHfxrOLfNi0f3mfdnq0-p-FmqOiSzj79IIRlzWsIQoYVFXJiVw-pn8ockeaLfIFRJcwOREqilwyCyx0SNaiaeawuOiLEV_1ObJHr6Dp59GIoIOIuTdf9W5CGrdvoWOw7OQMmGXn",
		},
		&NotificationVideo{
			ID:        "Sqpmvv8uulM",
			Title:     "心予報/歌わせていただきました。",
			Thumbnail: "https://i.ytimg.com/vi/OPzbUoLxYyE/default.jpg",
		},
	)
	if err != nil {
		t.Error(err)
	}
}

// OPzbUoLxYyE	心予報/歌わせていただきました。	PT3M24S	0	upcoming	false	2024-05-26 13:20:00	2024-05-25 12:00:03	2024-05-25 12:00:03	https://i.ytimg.com/vi/OPzbUoLxYyE/default.jpg
