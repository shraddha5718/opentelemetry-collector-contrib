// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlecloudexporter

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/featuregate"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter/internal/resourcemapping"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
}

func TestCreateExporter(t *testing.T) {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		t.Skip("Default credentials not set, skip creating Google Cloud exporter")
	}
	ctx := context.Background()
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	eCfg := cfg.(*Config)
	eCfg.ProjectID = "test"

	te, err := factory.CreateTraces(ctx, exportertest.NewNopSettings(metadata.Type), eCfg)
	assert.NoError(t, err)
	assert.NotNil(t, te, "failed to create trace exporter")

	me, err := factory.CreateMetrics(ctx, exportertest.NewNopSettings(metadata.Type), eCfg)
	assert.NoError(t, err)
	assert.NotNil(t, me, "failed to create metrics exporter")
}

func TestCreateLegacyExporter(t *testing.T) {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		t.Skip("Default credentials not set, skip creating Google Cloud exporter")
	}
	ctx := context.Background()
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	eCfg := cfg.(*Config)
	eCfg.ProjectID = "test"

	te, err := factory.CreateTraces(ctx, exportertest.NewNopSettings(metadata.Type), eCfg)
	assert.NoError(t, err)
	assert.NotNil(t, te, "failed to create trace exporter")

	me, err := factory.CreateMetrics(ctx, exportertest.NewNopSettings(metadata.Type), eCfg)
	assert.NoError(t, err)
	assert.NotNil(t, me, "failed to create metrics exporter")
}

func TestCustomMonitoredResourceMapping(t *testing.T) {
	_ = featuregate.GlobalRegistry().Set("exporter.googlecloud.CustomMonitoredResources", true)
	ctx := context.Background()
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	eCfg := cfg.(*Config)
	eCfg.ProjectID = "test"

	te, err := factory.CreateLogs(ctx, exportertest.NewNopSettings(metadata.Type), eCfg)
	assert.NoError(t, err)
	assert.NotNil(t, te, "failed to create logs exporter")

	actualLogFuncPointer := reflect.ValueOf(eCfg.LogConfig.MapMonitoredResource).Pointer()
	expectedLogFuncPointer := reflect.ValueOf(resourcemapping.CustomLoggingMonitoredResourceMapping).Pointer()
	assert.Equal(t, expectedLogFuncPointer, actualLogFuncPointer)

	me, err := factory.CreateMetrics(ctx, exportertest.NewNopSettings(metadata.Type), eCfg)
	assert.NoError(t, err)
	assert.NotNil(t, me, "failed to create metrics exporter")

	actualMetricsFuncPointer := reflect.ValueOf(eCfg.LogConfig.MapMonitoredResource).Pointer()
	expectedMetricsFuncPointer := reflect.ValueOf(resourcemapping.CustomLoggingMonitoredResourceMapping).Pointer()
	assert.Equal(t, expectedMetricsFuncPointer, actualMetricsFuncPointer)
}

func TestCreateLogsExporterWithServiceAuthKey(t *testing.T) {
	ctx := context.Background()
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	eCfg := cfg.(*Config)
	eCfg.ProjectID = "test-project"

	// Test that our conditional logic works by testing createClientOptions directly
	// Test with valid JSON (but incomplete for actual use)
	eCfg.ServiceAuthKey = `{"type": "service_account", "project_id": "test-project"}`
	copts, err := createClientOptions(ctx, eCfg)
	assert.NoError(t, err)
	assert.NotNil(t, copts)
	assert.Len(t, copts, 2) // Should have credentials and scopes

	// Test with empty ServiceAuthKey
	eCfg.ServiceAuthKey = ""
	copts2, err := createClientOptions(ctx, eCfg)
	assert.Error(t, err)
	assert.Nil(t, copts2)
	assert.Contains(t, err.Error(), "service_auth_key must be provided")
}

func TestCreateLogsExporterWithInvalidServiceAuthKey(t *testing.T) {
	ctx := context.Background()
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	eCfg := cfg.(*Config)
	eCfg.ProjectID = "test-project"

	// Test with invalid JSON ServiceAuthKey
	eCfg.ServiceAuthKey = `{"invalid": json}`
	copts, err := createClientOptions(ctx, eCfg)
	assert.Error(t, err)
	assert.Nil(t, copts)
	assert.Contains(t, err.Error(), "invalid service_auth_key format")
}
