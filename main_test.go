package main

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandleRequest(t *testing.T) {
	ctx := context.Background()

	type arrange struct {
		event events.IoTButtonEvent
	}

	type want struct {
		out    string
		errMsg string
	}

	for _, tt := range [...]struct {
		name    string
		arrange arrange
		want    want
	}{
		{
			name: "test",
			arrange: arrange{
				event: events.IoTButtonEvent{},
			},
			want: want{
				out: "OK",
			},
		},
	} {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleRequest(ctx, tt.arrange.event)

			if err != nil {
				if w, g := tt.want.errMsg, err.Error(); w != g {
					t.Fatalf("Err: want %s, got %s", w, g)
				}

				return
			}

			if w, g := tt.want.out, got; w != g {
				t.Fatalf("Output: want %s, got %s", w, g)
			}
		})
	}
}
