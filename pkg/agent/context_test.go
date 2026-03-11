package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amit-vikramaditya/v1claw/pkg/config"
	"github.com/amit-vikramaditya/v1claw/pkg/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMessages_SkipsEmptyCurrentMessage(t *testing.T) {
	workspace := t.TempDir()
	cb := NewContextBuilder(workspace)

	history := []providers.Message{
		{Role: "assistant", Content: "hello"},
	}

	messages := cb.BuildMessages(history, "", "", nil, "", "")
	require.Len(t, messages, 2)
	assert.Equal(t, "system", messages[0].Role)
	assert.Equal(t, "assistant", messages[1].Role)
}

func TestResolveBuiltinSkillsDir_PrefersWorkspaceSkills(t *testing.T) {
	workspace := t.TempDir()
	skillsDir := filepath.Join(workspace, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))

	resolved := resolveBuiltinSkillsDir(workspace)
	assert.Equal(t, skillsDir, resolved)
}

func TestResolveBuiltinSkillsDir_FallsBackToGlobalSkills(t *testing.T) {
	workspace := t.TempDir()
	home := t.TempDir()
	t.Setenv(config.HomeEnvVar, home)

	globalSkillsDir := filepath.Join(home, "skills")
	require.NoError(t, os.MkdirAll(globalSkillsDir, 0755))

	resolved := resolveBuiltinSkillsDir(workspace)
	assert.Equal(t, globalSkillsDir, resolved)
}
