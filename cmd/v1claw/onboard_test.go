package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupProviderInfo_KnownProvider(t *testing.T) {
	info, ok := lookupProviderInfo("gemini")
	assert.True(t, ok)
	assert.Equal(t, "gemini", info.id)
}

func TestLookupProviderInfo_UnknownProvider(t *testing.T) {
	_, ok := lookupProviderInfo("not-a-provider")
	assert.False(t, ok)
}

func TestDefaultProviderModel(t *testing.T) {
	assert.NotEmpty(t, defaultProviderModel("gemini"))
	assert.Empty(t, defaultProviderModel("nvidia"))
}

func TestSupportedProviderList_HasNoDuplicates(t *testing.T) {
	assert.Equal(t, "gemini, vertex, openai, anthropic, bedrock, azure_openai, groq, deepseek, openrouter, nvidia", supportedProviderList())
}

func TestHasFlag(t *testing.T) {
	assert.True(t, hasFlag([]string{"--auto", "--skip-test"}, "--skip-test"))
	assert.True(t, hasFlag([]string{"--skip-test=true"}, "--skip-test"))
	assert.False(t, hasFlag([]string{"--skip-test=false"}, "--skip-test"))
	assert.False(t, hasFlag([]string{"--auto"}, "--skip-test"))
}
