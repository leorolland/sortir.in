package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/leorolland/sortir.in/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionGet(t *testing.T) {
	setupTestPocketBase(t)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/version", PORT))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, version.Version, body["version"])
}
