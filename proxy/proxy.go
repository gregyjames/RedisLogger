package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"go.uber.org/zap"

	"redislogger/config"
	"redislogger/protocol"
)

// Proxy represents a Redis proxy server
type Proxy struct {
	config *config.Config
	logger *zap.Logger
}

// New creates a new Redis proxy
func New(cfg *config.Config, logger *zap.Logger) *Proxy {
	return &Proxy{
		config: cfg,
		logger: logger,
	}
}

// Start starts the Redis proxy server
func (p *Proxy) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", p.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	p.logger.Info("Redis proxy started", zap.String("listen_addr", p.config.ListenAddr))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			conn, err := listener.Accept()
			if err != nil {
				p.logger.Error("Failed to accept connection", zap.Error(err))
				continue
			}
			go p.handleConnection(conn)
		}
	}
}

func (p *Proxy) handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	connLogger := p.logger.With(zap.String("client_addr", clientAddr))
	connLogger.Info("New connection established")

	redisConn, err := net.Dial("tcp", p.config.RedisAddr)
	if err != nil {
		connLogger.Error("Failed to connect to Redis", zap.Error(err))
		return
	}
	defer redisConn.Close()

	parser := protocol.New(conn)
	var wg sync.WaitGroup
	wg.Add(2)

	// Forward commands from client to Redis
	go func() {
		defer wg.Done()
		for {
			cmd, err := parser.ReadCommand()
			if err != nil {
				if err != io.EOF {
					connLogger.Error("Failed to read command", zap.Error(err))
				}
				return
			}

			// Log command details with appropriate fields based on command type
			fields := []zap.Field{
				zap.String("command", cmd.Name),
			}

			// Add command-specific fields
			if len(cmd.Args) > 0 {
				switch strings.ToUpper(cmd.Name) {
				case "SET":
					if len(cmd.Args) >= 2 {
						fields = append(fields,
							zap.String("key", cmd.Args[0]),
							zap.String("value", cmd.Args[1]),
						)
						// Add SET options if present
						if len(cmd.Args) > 2 {
							options := make([]string, 0)
							for i := 2; i < len(cmd.Args); i++ {
								opt := strings.ToUpper(cmd.Args[i])
								switch opt {
								case "EX", "PX", "EXAT", "PXAT":
									if i+1 < len(cmd.Args) {
										options = append(options, fmt.Sprintf("%s=%s", opt, cmd.Args[i+1]))
										i++ // Skip the next argument as it's the value for this option
									}
								case "NX", "XX", "KEEPTTL":
									options = append(options, opt)
								}
							}
							if len(options) > 0 {
								fields = append(fields, zap.Strings("options", options))
							}
						}
					}
				case "GET", "MGET":
					fields = append(fields, zap.Strings("keys", cmd.Args))
				case "DEL", "EXISTS", "EXPIRE", "TTL", "PTTL", "PERSIST", "TYPE":
					fields = append(fields, zap.String("key", cmd.Args[0]))
				case "INCR", "DECR", "INCRBY", "DECRBY", "INCRBYFLOAT":
					if len(cmd.Args) >= 2 {
						fields = append(fields,
							zap.String("key", cmd.Args[0]),
							zap.String("amount", cmd.Args[1]),
						)
					}
				case "HSET", "HGET", "HDEL", "HEXISTS", "HINCRBY", "HINCRBYFLOAT":
					if len(cmd.Args) >= 2 {
						fields = append(fields,
							zap.String("key", cmd.Args[0]),
							zap.String("field", cmd.Args[1]),
						)
						if len(cmd.Args) > 2 {
							fields = append(fields, zap.String("value", cmd.Args[2]))
						}
					}
				case "LPUSH", "RPUSH", "LPUSHX", "RPUSHX":
					if len(cmd.Args) >= 2 {
						fields = append(fields,
							zap.String("key", cmd.Args[0]),
							zap.Strings("values", cmd.Args[1:]),
						)
					}
				case "SADD", "SREM", "SISMEMBER", "SCARD", "SPOP", "SRANDMEMBER":
					if len(cmd.Args) >= 1 {
						fields = append(fields, zap.String("key", cmd.Args[0]))
						if len(cmd.Args) > 1 {
							if cmd.Name == "SPOP" || cmd.Name == "SRANDMEMBER" {
								fields = append(fields, zap.String("count", cmd.Args[1]))
							} else {
								fields = append(fields, zap.Strings("members", cmd.Args[1:]))
							}
						}
					}
				case "ZADD":
					if len(cmd.Args) >= 3 {
						fields = append(fields, zap.String("key", cmd.Args[0]))
						pairs := make([]string, 0)
						for i := 1; i < len(cmd.Args); i += 2 {
							if i+1 < len(cmd.Args) {
								pairs = append(pairs, fmt.Sprintf("%s=%s", cmd.Args[i], cmd.Args[i+1]))
							}
						}
						fields = append(fields, zap.Strings("score_member_pairs", pairs))
					}
				default:
					fields = append(fields, zap.Strings("args", cmd.Args))
				}
			}

			connLogger.Info("Received command", fields...)

			if _, err := redisConn.Write(cmd.Message); err != nil {
				connLogger.Error("Failed to write to Redis", zap.Error(err))
				return
			}
		}
	}()

	// Forward responses from Redis to client
	go func() {
		defer wg.Done()
		io.Copy(conn, redisConn)
	}()

	wg.Wait()
	connLogger.Info("Connection closed")
} 