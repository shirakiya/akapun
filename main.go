package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rollbar/rollbar-go"
)

type ButtonClickedEvent struct {
	ClickType string `json:"clickType"`
}

type DeviceEvent struct {
	ButtonClicked ButtonClickedEvent `json:"buttonClicked"`
}

type IoTClickEvent struct {
	DeviceEvent DeviceEvent `json:"deviceEvent"`
}

type AkashiStampParams struct {
	Token    string `json:"token"`
	Type     int    `json:"type"`
	Timezone string `json:"timezone"`
}

type AkashiResponse struct {
	Success bool `json:"success"`
}

type ClickType int

const (
	ClickTypeSingle ClickType = iota
	ClickTypeDouble
	ClickTypeLong
)

type Recorder interface {
	Do(context.Context, ClickType) error
}

type AkashiRecorder struct {
	BaseURL string
	CorpID  string
	Token   string
}

func (rec AkashiRecorder) Do(ctx context.Context, cType ClickType) error {
	const (
		// means "出勤" in Akashi API.
		PunchIn = 11

		// means "退勤" in Akashi API.
		PunchOut = 12
	)

	var t int
	switch cType {
	case ClickTypeSingle, ClickTypeLong:
		t = PunchIn
	case ClickTypeDouble:
		t = PunchOut
	}

	params := AkashiStampParams{
		Token:    rec.Token,
		Type:     t,
		Timezone: "+09:00",
	}

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := rec.buildURL()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonParams))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))

	status := resp.StatusCode
	if status != http.StatusOK {
		return fmt.Errorf("status code was not 200: %d", status)
	}

	var ar AkashiResponse
	err = json.Unmarshal(body, &ar)
	if err != nil {
		return err
	}

	if !ar.Success {
		return errors.New("something error from Akashi")
	}

	return nil
}

func (rec AkashiRecorder) buildURL() string {
	return fmt.Sprintf("%s/%s/stamps", rec.BaseURL, rec.CorpID)
}

type Akapun struct {
	Recorder Recorder
}

func (akapun Akapun) HandleRequest(
	ctx context.Context,
	event IoTClickEvent,
) (string, error) {
	var cType ClickType
	switch t := event.DeviceEvent.ButtonClicked.ClickType; t {
	case "SINGLE":
		cType = ClickTypeSingle
	case "DOUBLE":
		cType = ClickTypeDouble
	case "LONG":
		cType = ClickTypeLong
	default:
		panic(fmt.Errorf("unknown click type was given: %s", t))
	}

	if err := akapun.Recorder.Do(ctx, cType); err != nil {
		panic(err)
	}

	return "OK", nil
}

func setupRollbar(token string) {
	rollbar.SetToken(token)
	rollbar.SetEnvironment("production")
	rollbar.SetServerHost("AWS Lambda")
	rollbar.SetServerRoot("github.com/shirakiya/akapun")
}

func main() {
	const AkashiURL = "https://atnd.ak4.jp/api/cooperation"

	corpID := os.Getenv("AKASHI_CORP_ID")
	akashiToken := os.Getenv("AKASHI_TOKEN")
	rollbarToken := os.Getenv("ROLLBAR_TOKEN")

	setupRollbar(rollbarToken)

	akapun := Akapun{
		Recorder: AkashiRecorder{
			BaseURL: AkashiURL,
			CorpID:  corpID,
			Token:   akashiToken,
		},
	}

	lambda.Start(rollbar.LambdaWrapper(akapun.HandleRequest))
}
