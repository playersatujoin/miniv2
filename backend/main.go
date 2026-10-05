package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"miniv2/backend/internal/api"
	"miniv2/backend/internal/sim"
	"miniv2/backend/internal/store"
	"miniv2/backend/internal/world"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dataDir := flag.String("data", "data", "directory where maps are stored")
	staticDir := flag.String("static", "", "optional built frontend to serve, e.g. ../frontend/dist")
	flag.Parse()

	st, err := store.Open(*dataDir)
	if err != nil {
		slog.Error("open store", "err", err)
		os.Exit(1)
	}
	if len(st.List()) == 0 {
		if err := seedStarterMap(st); err != nil {
			slog.Error("seed starter map", "err", err)
			os.Exit(1)
		}
	}

	// Every map has a living world that keeps evolving while the server runs.
	sims, err := sim.NewManager(filepath.Join(*dataDir, "sims"))
	if err != nil {
		slog.Error("open worlds", "err", err)
		os.Exit(1)
	}
	for _, sum := range st.List() {
		if m, err := st.Get(sum.ID); err == nil {
			sims.Start(m)
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", api.New(st, sims))
	if *staticDir != "" {
		mux.Handle("/", spaHandler(*staticDir))
	}

	// Cancelled when shutdown begins so long-lived event streams end promptly.
	baseCtx, cancelBase := context.WithCancel(context.Background())
	srv := &http.Server{
		Addr:              *addr,
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return baseCtx },
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		cancelBase()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	slog.Info("listening", "addr", *addr, "data", *dataDir)
	err = srv.ListenAndServe()
	if closeErr := sims.Close(); closeErr != nil {
		slog.Error("save worlds", "err", closeErr)
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}

func seedStarterMap(st *store.FileStore) error {
	m := world.Generate(world.GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
	m.ID = store.NewID()
	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = m.CreatedAt
	slog.Info("created starter map", "id", m.ID)
	return st.Save(m)
}

// spaHandler serves files from dir, falling back to index.html so client-side
// routes like /maps/abc work on reload.
func spaHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(p); err != nil || info.IsDir() {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		files.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Unwrap lets http.ResponseController reach the real writer (for Flush).
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("http", "method", r.Method, "path", r.URL.Path, "status", rec.status, "dur", time.Since(start).Round(time.Microsecond))
	})
}
