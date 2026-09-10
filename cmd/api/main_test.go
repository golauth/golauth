package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	dbmock "github.com/golauth/golauth/pkg/infra/database/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// freeAddr returns a loopback address nothing is listening on. There is a small
// window between the close here and serve() binding it, which is acceptable for
// a test.
func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())
	return addr
}

// TestServeDrainsInFlightRequestAndClosesDB is the shutdown contract: a request
// already being handled when the signal arrives runs to completion, and the
// database is closed exactly once afterwards.
func TestServeDrainsInFlightRequestAndClosesDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := dbmock.NewMockDatabase(ctrl)
	db.EXPECT().Close().Times(1)

	app := fiber.New()
	app.Get("/ping", func(c fiber.Ctx) error { return c.SendString("ok") })
	app.Get("/slow", func(c fiber.Ctx) error {
		time.Sleep(300 * time.Millisecond)
		return c.SendString("done")
	})

	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())

	var (
		wg       sync.WaitGroup
		serveErr error
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		serveErr = serve(ctx, app, addr, 5*time.Second, db)
	}()

	base := "http://" + addr
	require.Eventually(t, func() bool {
		resp, err := http.Get(base + "/ping")
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 2*time.Second, 20*time.Millisecond, "server never came up")

	type result struct {
		status int
		body   string
		err    error
	}
	done := make(chan result, 1)
	go func() {
		resp, err := http.Get(base + "/slow")
		if err != nil {
			done <- result{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		b, _ := io.ReadAll(resp.Body)
		done <- result{status: resp.StatusCode, body: string(b)}
	}()

	time.Sleep(50 * time.Millisecond) // let the slow handler start
	cancel()                          // stand in for SIGTERM

	select {
	case r := <-done:
		require.NoError(t, r.err, "in-flight request was cut instead of drained")
		require.Equal(t, http.StatusOK, r.status)
		require.Equal(t, "done", r.body)
	case <-time.After(3 * time.Second):
		t.Fatal("in-flight request never completed")
	}

	wg.Wait()
	require.NoError(t, serveErr, "graceful shutdown should return no error")
}

// TestServeReturnsListenerError surfaces a bind failure instead of hanging, and
// still closes the database.
func TestServeReturnsListenerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := dbmock.NewMockDatabase(ctrl)
	db.EXPECT().Close().Times(1)

	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = blocker.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = serve(ctx, fiber.New(), blocker.Addr().String(), time.Second, db)
	require.Error(t, err, "binding an in-use port must fail")
}
