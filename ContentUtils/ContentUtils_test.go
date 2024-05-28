package ContentUtils_test

import (
	"jaxon/embedbot/ContentUtils"
	"testing"
)

func TestTikTokDownload(t *testing.T) {
	type args struct {
		url               string
		should_be_spoiled bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TikTok Download test 1",
			args: args{url: "https://www.tiktok.com/t/ZPRKkGh19/", should_be_spoiled: false},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ContentUtils.DownloadTikTokVideo(tt.args.url, tt.args.should_be_spoiled)
			if err != nil {
				t.Errorf("DownloadTikTokVideo() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestShouldBeSpoilered(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should be spoilered test 1",
			args: args{content: "||https://google.com||"},
			want: true,
		},
		{
			name: "Should be spoilered test 2",
			args: args{content: "https://google.com"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContentUtils.ShouldBeSpoilered(tt.args.content); got != tt.want {
				t.Errorf("ShouldBeSpoilered() = %v, want %v", got, tt.want)
			}
		})
	}
}
