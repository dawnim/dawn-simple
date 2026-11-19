package momentcontent

import (
	"reflect"
	"testing"
)

func TestParseMedia(t *testing.T) {
	tests := []struct {
		text   string
		ids    []uint
		inline bool
	}{
		{"图片 [attachment://3] [attachment://8] [attachment://3]", []uint{3, 8}, true},
		{"`[attachment://3]`\n\n```md\n[attachment://4]\n```\n\n\\[attachment://5]", nil, false},
		{"[attachment://3](https://example.com) ![alt](https://example.com/a.png)", nil, true},
		{"![alt][figure]\n\n[figure]: ./img.png", nil, true},
		{"![test](attachment://3) [upload://3]", nil, true},
		{"<audio src=\"/audio.mp3\"></audio>", nil, true},
		{"MEDIA::video::https://example.com/movie.mp4", nil, true},
	}
	for _, test := range tests {
		media := ParseMedia(test.text)
		if !reflect.DeepEqual(media.IDs, test.ids) || media.Inline != test.inline {
			t.Errorf("ParseMedia(%q) = %#v", test.text, media)
		}
	}
}
