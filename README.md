[![Go](https://github.com/gregyjames/RedisLogger/actions/workflows/go.yml/badge.svg)](https://github.com/gregyjames/RedisLogger/actions/workflows/go.yml)
# Redis Logger

A lightweight Redis proxy server written in Go that provides command logging and monitoring capabilities.

## Features

- Forwards Redis commands from clients to a Redis server
- Detailed command logging with structured fields
- Support for common Redis commands with parameter logging
- Graceful shutdown handling
- Debug logging for troubleshooting

## Project Structure

```
.
├── main.go           # Main entry point
├── config/
│   └── config.go     # Configuration handling
├── protocol/
│   └── parser.go     # Redis protocol parser
├── proxy/
│   └── proxy.go      # Proxy implementation
├── config.json       # Configuration file
└── Dockerfile        # Docker build configuration
```

## Configuration

Create a `config.json` file with the following structure:

```json
{
    "listen_addr": ":9000",    // Address to listen for Redis connections
    "redis_addr": "localhost:6379"  // Address of the Redis server to proxy to
}
```

## Command Logging

The proxy logs detailed information about Redis commands, including:

- Command name
- Command-specific parameters (keys, values, options)
- Client connection details
- Connection lifecycle events

### Supported Command Types

- String commands (SET, GET, MGET, etc.)
- Hash commands (HSET, HGET, HDEL, etc.)
- List commands (LPUSH, RPUSH, etc.)
- Set commands (SADD, SREM, etc.)
- Sorted Set commands (ZADD, etc.)

## Development

### Prerequisites

- Go 1.22 or later
- Redis server

### Building

```bash
go build
```

### Running

```bash
./redislogger
```

### Docker

Build the Docker image:
```bash
docker build -t redis-proxy .
```

Run the container:
```bash
docker run -d \
  --name redis-proxy \
  -p 6379:6379 \
  -e REDIS_ADDR=your-redis-server:6379 \
  redis-proxy
```

### Debugging

The project includes VS Code debugging configuration. To debug:

1. Open the project in VS Code
2. Set breakpoints by clicking in the left margin
3. Press F5 or use the Run and Debug sidebar
4. Select "Debug Redis Proxy"

## Logging

The proxy uses structured logging with the following levels:
- Debug: Detailed program flow information
- Info: Command execution and connection events
- Error: Connection and command processing errors

## License

MIT License

Copyright (c) 2025 Greg James

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
