package triplej

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchSongsFromTriplejAPI(t *testing.T) {
	type args struct {
		playlistSize int
	}
	tests := []struct {
		name       string
		args       args
		wantLength int
		wantErr    bool
	}{
		{
			name: "test valid playlist size",
			args: args{
				playlistSize: 10,
			},
			wantLength: 10,
			wantErr:    false,
		},
		{
			name: "test another valid playlist size",
			args: args{
				playlistSize: 0,
			},
			wantLength: 0,
			wantErr:    false,
		},
		{
			name: "test 0 playlist size",
			args: args{
				playlistSize: 0,
			},
			wantLength: 0,
			wantErr:    false,
		},
		{
			name: "test invalid playlist size",
			args: args{
				playlistSize: -1,
			},
			wantLength: 0,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tctx := context.Background()
			c := NewTiplejClient()
			got, err := c.FetchSongsFromTriplejAPI(tctx, tt.args.playlistSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchSongsFromTriplejAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Len(t, got, tt.wantLength, "FetchSongsFromTriplejAPI() got = %v, wantLength %v", got, tt.wantLength)
		})
	}
}
