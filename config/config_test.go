package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/config"
)

func TestLoadMultipleConfigFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	baseCfgPath := filepath.Join(tempDir, "base.yaml")
	overlayCfgPath := filepath.Join(tempDir, "overlay.yaml")

	baseCfg := []byte(`domain: example.com
objectStorage:
  enabled: false
  providers:
    disk:
      enabled: false
`)

	overlayCfg := []byte(`objectStorage:
  enabled: true
  providers:
    disk:
      enabled: true
      bucket: ./tmp
`)

	assert.NilError(t, os.WriteFile(baseCfgPath, baseCfg, 0o600))
	assert.NilError(t, os.WriteFile(overlayCfgPath, overlayCfg, 0o600))

	configArg := baseCfgPath + "," + overlayCfgPath

	cfg, err := config.Load(&configArg)
	assert.NilError(t, err)
	assert.Assert(t, cfg != nil)

	assert.Check(t, is.Equal("example.com", cfg.Domain))
	assert.Check(t, is.Equal(true, cfg.ObjectStorage.Enabled))
	assert.Check(t, is.Equal(true, cfg.ObjectStorage.Providers.Disk.Enabled))
	assert.Check(t, is.Equal("./tmp", cfg.ObjectStorage.Providers.Disk.Bucket))
}
