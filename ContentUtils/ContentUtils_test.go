package ContentUtils_test

import (
	"jaxon/embedbot/ContentUtils"
	"testing"
)

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
