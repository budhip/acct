package flag

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	featureFlag "bitbucket.org/Amartha/go-feature-flag-sdk"
	"bitbucket.org/Amartha/go-feature-flag-sdk/listener"
)

type FlaggerClient interface {
	featureFlag.IFlagger
}

type Variant[T any] struct {
	Enabled bool
	Value   T
}

func New(cfg *config.Configuration) (FlaggerClient, error) {
	flagConfig := featureFlag.Config{
		AppName:               cfg.App.Name,
		FeatureFlagServiceURL: cfg.FeatureFlag.URL,
		Token:                 cfg.FeatureFlag.Token,
		Env:                   cfg.FeatureFlag.Env,
		RefreshInterval:       cfg.FeatureFlag.RefreshInterval,
		Listener:              listener.DebugListener{},
		HttpClient:            http.DefaultClient,
	}
	c, err := featureFlag.NewFlagger(&flagConfig)
	if err != nil {
		return nil, err
	}
	c.WaitForReady()

	return c, nil
}

// GetVariant returns the variant for the given key.
// We use this method because golang doesn't support generic type parameters in method interfaces
// [link_issue](https://github.com/golang/go/issues/49085)
func GetVariant[T any](c FlaggerClient, key string) (*Variant[T], error) {
	variant := c.GetVariant(key)
	if variant == nil {
		return nil, models.GetErrMap(models.ErrKeyDataNotFound, fmt.Sprintf("variant for key %s not found", key))
	}

	var res T
	if !variant.Enabled {
		return &Variant[T]{
			Enabled: variant.Enabled,
			Value:   res,
		}, nil
	}

	if err := json.Unmarshal([]byte(variant.Payload.Value), &res); err != nil {
		return nil, models.GetErrMap(models.ErrKeyFailedUnmarshal, fmt.Sprintf("unmarshal variant for key %s failed: %v", key, err))
	}

	return &Variant[T]{
		Enabled: variant.Enabled,
		Value:   res,
	}, nil
}
