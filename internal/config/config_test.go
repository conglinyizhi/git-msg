package config_test

import (
	"testing"

	"gitmsg/internal/config"
	"gitmsg/internal/utils"

	"github.com/go-playground/assert/v2"
	"github.com/spf13/afero"
)

const testTomlContent = "API_KEY = 'test-key'\nBASE_URL = 'https://api.deepseek.com/'\nMODEL_NAME = 'deepseek-v4-flash'"

func TestReadConfigTest(t *testing.T) {
	fs := afero.NewMemMapFs()
	tomlFileURL, err := utils.GetConfigRootDir("llm.toml")
	if err != nil {
		t.Fatal("因 utils.GetConfigRootDir 崩溃导致的测试失败，", err)
	}
	afero.WriteFile(fs, tomlFileURL, []byte(testTomlContent), 0644)
	cfg, err := config.GetConfigValue(fs)
	assert.Equal(t, "test-key", cfg.API_KEY)
	assert.Equal(t, "https://api.deepseek.com/", cfg.BASE_URL)
	assert.Equal(t, "deepseek-v4-flash", cfg.MODEL_NAME)
}
