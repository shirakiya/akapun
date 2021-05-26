package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/go-cmp/cmp"
)

func TestIoTClickEventCanParseEventJSON(t *testing.T) {
	f, err := os.Open("event.sample.json")
	if err != nil {
		t.Fatal(err)
	}

	var event events.IoTOneClickEvent

	if err := json.NewDecoder(f).Decode(&event); err != nil {
		t.Fatal(err)
	}

	if event.DeviceEvent.ButtonClicked.ClickType != "SINGLE" {
		t.Fatal("Unmarshal error!")
	}
}

// nolint:funlen
func TestAkashiRecorder_Do(t *testing.T) {
	ctx := context.Background()

	type arrange struct {
		corpID     string
		token      string
		clickType  ClickType
		statusCode int
		respBody   []byte
	}

	type want struct {
		path   string
		req    map[string]interface{}
		errMsg string
	}

	for _, tt := range [...]struct {
		name    string
		arrange arrange
		want    want
	}{
		{
			name: "click type is single",
			arrange: arrange{
				corpID:     "corp-id",
				token:      "foo",
				clickType:  ClickTypeSingle,
				statusCode: http.StatusOK,
				respBody:   []byte("{\"success\":true}"),
			},
			want: want{
				path: "/corp-id/stamps",
				req: map[string]interface{}{
					"token":    "foo",
					"type":     float64(11),
					"timezone": "+09:00",
				},
			},
		},
		{
			name: "click type is double",
			arrange: arrange{
				corpID:     "corp-id",
				token:      "foo",
				clickType:  ClickTypeDouble,
				statusCode: http.StatusOK,
				respBody:   []byte("{\"success\":true}"),
			},
			want: want{
				path: "/corp-id/stamps",
				req: map[string]interface{}{
					"token":    "foo",
					"type":     float64(12),
					"timezone": "+09:00",
				},
			},
		},
		{
			name: "click type is long",
			arrange: arrange{
				corpID:     "corp-id",
				token:      "foo",
				clickType:  ClickTypeLong,
				statusCode: http.StatusOK,
				respBody:   []byte("{\"success\":true}"),
			},
			want: want{
				path: "/corp-id/stamps",
				req: map[string]interface{}{
					"token":    "foo",
					"type":     float64(11),
					"timezone": "+09:00",
				},
			},
		},
		{
			name: "error if Akashi responses that status code is not 200",
			arrange: arrange{
				corpID:     "corp-id",
				token:      "foo",
				clickType:  ClickTypeLong,
				statusCode: http.StatusForbidden,
				respBody:   []byte("{\"success\":true}"),
			},
			want: want{
				path: "/corp-id/stamps",
				req: map[string]interface{}{
					"token":    "foo",
					"type":     float64(11),
					"timezone": "+09:00",
				},
				errMsg: "status code was not 200: 403",
			},
		},
		{
			name: "error if Akashi responses that success is false",
			arrange: arrange{
				corpID:     "corp-id",
				token:      "foo",
				clickType:  ClickTypeSingle,
				statusCode: http.StatusOK,
				respBody:   []byte("{\"success\":false}"),
			},
			want: want{
				path: "/corp-id/stamps",
				req: map[string]interface{}{
					"token":    "foo",
					"type":     float64(11),
					"timezone": "+09:00",
				},
				errMsg: "something error from Akashi",
			},
		},
	} {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()

				if w, g := tt.want.path, r.URL.Path; w != g {
					t.Fatalf("Path: want %s, got %s", w, g)
				}

				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				var req interface{}
				if err = json.Unmarshal(body, &req); err != nil {
					t.Fatal(err)
				}

				if d := cmp.Diff(tt.want.req, req); d != "" {
					t.Fatalf("Akashi Request: %s", d)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.arrange.statusCode)

				if _, err = w.Write(tt.arrange.respBody); err != nil {
					t.Fatal(err)
				}
			}))
			defer ts.Close()

			rec := AkashiRecorder{
				BaseURL: ts.URL,
				CorpID:  tt.arrange.corpID,
				Token:   tt.arrange.token,
			}

			err := rec.Do(ctx, tt.arrange.clickType)
			if err != nil || tt.want.errMsg != "" {
				if w, g := tt.want.errMsg, err.Error(); w != g {
					t.Fatalf("Err: want %s, got %s", w, g)
				}
			}
		})
	}
}

// nolint:funlen
func TestAkapun_HandleRequest(t *testing.T) {
	ctx := context.Background()

	type arrange struct {
		event events.IoTOneClickEvent
		err   error
	}

	type want struct {
		out       string
		errMsg    string
		clickType ClickType
	}

	for _, tt := range [...]struct {
		name    string
		arrange arrange
		want    want
	}{
		{
			name: "single push",
			arrange: arrange{
				event: events.IoTOneClickEvent{
					DeviceEvent: events.IoTOneClickDeviceEvent{
						ButtonClicked: events.IoTOneClickButtonClicked{
							ClickType: "SINGLE",
						},
					},
				},
			},
			want: want{
				out:       "OK",
				clickType: ClickTypeSingle,
			},
		},
		{
			name: "double push",
			arrange: arrange{
				event: events.IoTOneClickEvent{
					DeviceEvent: events.IoTOneClickDeviceEvent{
						ButtonClicked: events.IoTOneClickButtonClicked{
							ClickType: "DOUBLE",
						},
					},
				},
			},
			want: want{
				out:       "OK",
				clickType: ClickTypeDouble,
			},
		},
		{
			name: "long push",
			arrange: arrange{
				event: events.IoTOneClickEvent{
					DeviceEvent: events.IoTOneClickDeviceEvent{
						ButtonClicked: events.IoTOneClickButtonClicked{
							ClickType: "LONG",
						},
					},
				},
			},
			want: want{
				out:       "OK",
				clickType: ClickTypeLong,
			},
		},
		{
			name: "return error if unknown click type is given",
			arrange: arrange{
				event: events.IoTOneClickEvent{
					DeviceEvent: events.IoTOneClickDeviceEvent{
						ButtonClicked: events.IoTOneClickButtonClicked{
							ClickType: "unknown-type",
						},
					},
				},
			},
			want: want{
				out:    "NG",
				errMsg: "unknown click type was given: unknown-type",
			},
		},
		{
			name: "return error if recorder returns error",
			arrange: arrange{
				event: events.IoTOneClickEvent{
					DeviceEvent: events.IoTOneClickDeviceEvent{
						ButtonClicked: events.IoTOneClickButtonClicked{
							ClickType: "SINGLE",
						},
					},
				},
				err: errors.New("test error"),
			},
			want: want{
				out:    "NG",
				errMsg: "test error",
			},
		},
	} {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					if w, g := tt.want.errMsg, fmt.Sprintf("%s", err); w != g {
						t.Fatalf("Err: want %s, got %s", w, g)
					}
				}
			}()

			rec := &mockRecorder{
				err: tt.arrange.err,
			}
			akapun := Akapun{
				Recorder: rec,
			}

			out, _ := akapun.HandleRequest(ctx, tt.arrange.event)

			if w, g := tt.want.out, out; w != g {
				t.Fatalf("Output: want %s, got %s", w, g)
			}

			if w, g := tt.want.clickType, rec.clickType; w != g {
				t.Fatalf("ClickType: want %d, got %d", w, g)
			}
		})
	}
}

type mockRecorder struct {
	clickType ClickType
	err       error
}

func (rec *mockRecorder) Do(ctx context.Context, cType ClickType) error {
	rec.clickType = cType

	return rec.err
}
