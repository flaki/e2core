package enginetest

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/suborbital/e2core/foundation/scheduler"
	"github.com/suborbital/e2core/sat/engine2"
	"github.com/suborbital/e2core/sat/engine2/api"
	"github.com/suborbital/systemspec/capabilities"
	"github.com/suborbital/systemspec/request"
)

func TestEngineDisabledHTTP(t *testing.T) {
	config := capabilities.DefaultCapabilityConfig()
	config.HTTP = &capabilities.HTTPConfig{Enabled: false}

	apiInstance, _ := api.NewWithConfig(zerolog.Nop(), config)

	ref, err := engine2.WasmRefFromFile("./testdata/fetch/fetch.wasm")
	if err != nil {
		t.Error(err)
		return
	}

	e := engine2.New("fetch", ref, apiInstance)

	req := &request.CoordinatedRequest{
		Method: "GET",
		URL:    "/hello/world",
		ID:     uuid.New().String(),
		Body:   []byte("https://1password.com"),
	}

	_, err = e.Do(scheduler.NewJob("fetch", req)).Then()
	if err != nil {
		if err.Error() != `{"code":1,"message":"capability is not enabled"}` {
			t.Error("received incorrect error", err.Error())
		}
	} else {
		t.Error("module should have failed")
	}
}
