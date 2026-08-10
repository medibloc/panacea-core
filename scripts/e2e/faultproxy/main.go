// Command faultproxy is a test-only TCP proxy used to inject deterministic,
// container-scoped P2P delay, jitter, and effective packet loss. It never
// changes the host firewall or host network namespace.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const proxyBufferSize = 32 << 10

var errInjectedLoss = errors.New("faultproxy injected effective packet loss")

type faultConfig struct {
	ListenAddress string
	TargetAddress string
	BaseDelay     time.Duration
	Jitter        time.Duration
	DropEvery     uint64
	DialTimeout   time.Duration
	Seed          int64
}

func (c faultConfig) validate() error {
	var validationErrors []error
	if strings.TrimSpace(c.ListenAddress) == "" {
		validationErrors = append(validationErrors, errors.New("listen address is required"))
	}
	if strings.TrimSpace(c.TargetAddress) == "" {
		validationErrors = append(validationErrors, errors.New("target address is required"))
	}
	if c.BaseDelay < 0 {
		validationErrors = append(validationErrors, errors.New("base delay cannot be negative"))
	}
	if c.Jitter < 0 {
		validationErrors = append(validationErrors, errors.New("jitter cannot be negative"))
	}
	if c.DialTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("dial timeout must be positive"))
	}
	if c.Seed == 0 {
		validationErrors = append(validationErrors, errors.New("seed must be non-zero"))
	}
	return errors.Join(validationErrors...)
}

type jsonEventLogger struct {
	mu      sync.Mutex
	encoder *json.Encoder
}

func newJSONEventLogger(writer io.Writer) *jsonEventLogger {
	return &jsonEventLogger{encoder: json.NewEncoder(writer)}
}

func (l *jsonEventLogger) event(kind string, fields map[string]any) {
	if l == nil {
		return
	}
	record := make(map[string]any, len(fields)+2)
	record["recorded_at"] = time.Now().UTC()
	record["event"] = kind
	for key, value := range fields {
		record[key] = value
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.encoder.Encode(record)
}

func main() {
	config := faultConfig{}
	flag.StringVar(&config.ListenAddress, "listen", ":26656", "TCP listen address")
	flag.StringVar(&config.TargetAddress, "target", "", "TCP upstream address")
	flag.DurationVar(&config.BaseDelay, "delay", 0, "base delay applied before each stream chunk")
	flag.DurationVar(&config.Jitter, "jitter", 0, "deterministic delay variation in either direction")
	flag.Uint64Var(&config.DropEvery, "drop-every", 0, "close the proxied stream on every Nth chunk; zero disables loss")
	flag.DurationVar(&config.DialTimeout, "dial-timeout", 5*time.Second, "upstream dial timeout")
	flag.Int64Var(&config.Seed, "seed", 20260804, "non-zero deterministic jitter seed")
	flag.Parse()

	logger := newJSONEventLogger(os.Stdout)
	if err := config.validate(); err != nil {
		logger.event("configuration-error", map[string]any{"error": err.Error()})
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := runFaultProxy(ctx, config, logger); err != nil && !errors.Is(err, context.Canceled) {
		logger.event("proxy-error", map[string]any{"error": err.Error()})
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runFaultProxy(ctx context.Context, config faultConfig, logger *jsonEventLogger) error {
	if err := config.validate(); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", config.ListenAddress, err)
	}
	defer listener.Close()
	logger.event("listening", map[string]any{
		"listen":     listener.Addr().String(),
		"target":     config.TargetAddress,
		"delay":      config.BaseDelay.String(),
		"jitter":     config.Jitter.String(),
		"drop_every": config.DropEvery,
		"seed":       config.Seed,
	})

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()
	var connectionSequence atomic.Uint64
	for {
		client, acceptErr := listener.Accept()
		if acceptErr != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("accept proxied connection: %w", acceptErr)
		}
		connectionID := connectionSequence.Add(1)
		go proxyConnection(ctx, config, logger, connectionID, client)
	}
}

func proxyConnection(
	ctx context.Context,
	config faultConfig,
	logger *jsonEventLogger,
	connectionID uint64,
	client net.Conn,
) {
	defer client.Close()
	dialer := net.Dialer{Timeout: config.DialTimeout}
	upstream, err := dialer.DialContext(ctx, "tcp", config.TargetAddress)
	if err != nil {
		logger.event("upstream-dial-failed", map[string]any{
			"connection_id": connectionID,
			"target":        config.TargetAddress,
			"error":         err.Error(),
		})
		return
	}
	defer upstream.Close()
	logger.event("connection-opened", map[string]any{
		"connection_id": connectionID,
		"client":        client.RemoteAddr().String(),
		"target":        config.TargetAddress,
	})

	results := make(chan error, 2)
	go func() {
		results <- proxyCopy(
			ctx,
			upstream,
			client,
			config,
			rand.New(rand.NewSource(config.Seed+int64(connectionID)*2)),
			logger,
			connectionID,
			"client-to-target",
			time.Sleep,
		)
	}()
	go func() {
		results <- proxyCopy(
			ctx,
			client,
			upstream,
			config,
			rand.New(rand.NewSource(config.Seed+int64(connectionID)*2+1)),
			logger,
			connectionID,
			"target-to-client",
			time.Sleep,
		)
	}()
	firstErr := <-results
	_ = client.Close()
	_ = upstream.Close()
	logger.event("connection-closed", map[string]any{
		"connection_id": connectionID,
		"error":         errorText(firstErr),
	})
}

func proxyCopy(
	ctx context.Context,
	destination io.Writer,
	source io.Reader,
	config faultConfig,
	random *rand.Rand,
	logger *jsonEventLogger,
	connectionID uint64,
	direction string,
	sleep func(time.Duration),
) error {
	buffer := make([]byte, proxyBufferSize)
	var chunk uint64
	for {
		read, readErr := source.Read(buffer)
		if read > 0 {
			chunk++
			delay := jitteredDelay(config.BaseDelay, config.Jitter, random)
			if delay > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					sleep(delay)
				}
			}
			if config.DropEvery > 0 && chunk%config.DropEvery == 0 {
				logger.event("chunk-dropped", map[string]any{
					"connection_id": connectionID,
					"direction":     direction,
					"chunk":         chunk,
					"bytes":         read,
				})
				return errInjectedLoss
			}
			if _, writeErr := destination.Write(buffer[:read]); writeErr != nil {
				return writeErr
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return readErr
		}
	}
}

func jitteredDelay(base, jitter time.Duration, random *rand.Rand) time.Duration {
	if jitter <= 0 || random == nil {
		return base
	}
	width := int64(jitter)*2 + 1
	delta := time.Duration(random.Int63n(width) - int64(jitter))
	delay := base + delta
	if delay < 0 {
		return 0
	}
	return delay
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
