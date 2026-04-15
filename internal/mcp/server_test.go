package mcp_test

import (
	"context"
	"testing"
	"time"

	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/mcp"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/stretchr/testify/require"
)

func TestServer_Run_StartsAndStops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MCP server integration test in short mode")
	}

	tmpFile := t.TempDir() + "/test.db"
	database, err := db.New(tmpFile)
	require.NoError(t, err)
	defer func() { _ = database.Close() }()

	memRepo := memory.NewRepository(database.DB())
	server := mcp.NewServer(memRepo, nil, nil, nil)
	server.SetVersion("test-1.0.0")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx)
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("server.Run did not exit within expected time")
	}
}

func TestServer_SetVersion(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	database, err := db.New(tmpFile)
	require.NoError(t, err)
	defer func() { _ = database.Close() }()

	memRepo := memory.NewRepository(database.DB())
	server := mcp.NewServer(memRepo, nil, nil, nil)

	server.SetVersion("2.0.0")
}
