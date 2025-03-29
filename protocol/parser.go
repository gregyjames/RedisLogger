package protocol

import (
	"fmt"
	"io"
	"strings"
)

// Parser parses Redis protocol messages
type Parser struct {
	reader io.Reader
}

// New creates a new Redis protocol parser
func New(reader io.Reader) *Parser {
	return &Parser{reader: reader}
}

// Command represents a parsed Redis command
type Command struct {
	Name    string
	Message []byte
	Args    []string
}

// ReadCommand reads and parses the next Redis command
func (p *Parser) ReadCommand() (*Command, error) {
	// Read the first byte which indicates the message type
	header := make([]byte, 1)
	if _, err := p.reader.Read(header); err != nil {
		return nil, err
	}

	// Create a buffer to store the complete message
	var buf strings.Builder
	buf.Write(header)

	switch header[0] {
	case '*': // Array
		return p.parseArray(&buf)
	case '$': // Bulk string
		return p.parseBulkString(&buf)
	case '+': // Simple string
		return p.parseSimpleString(&buf)
	case '-': // Error
		return p.parseError(&buf)
	case ':': // Integer
		return p.parseInteger(&buf)
	default:
		return nil, fmt.Errorf("unknown protocol type: %c", header[0])
	}
}

func (p *Parser) parseArray(buf *strings.Builder) (*Command, error) {
	// Read the number of arguments
	var argCount int
	if _, err := fmt.Fscanf(p.reader, "%d\r\n", &argCount); err != nil {
		return nil, err
	}
	buf.WriteString(fmt.Sprintf("%d\r\n", argCount))

	if argCount < 1 {
		return nil, fmt.Errorf("invalid argument count: %d", argCount)
	}

	// Read the command name
	var cmdLen int
	if _, err := fmt.Fscanf(p.reader, "$%d\r\n", &cmdLen); err != nil {
		return nil, err
	}
	buf.WriteString(fmt.Sprintf("$%d\r\n", cmdLen))

	cmd := make([]byte, cmdLen)
	if _, err := io.ReadFull(p.reader, cmd); err != nil {
		return nil, err
	}
	buf.Write(cmd)

	if _, err := p.reader.Read(make([]byte, 2)); err != nil {
		return nil, err
	}
	buf.WriteString("\r\n")

	// Read remaining arguments
	args := make([]string, 0, argCount-1)
	for i := 1; i < argCount; i++ {
		var argLen int
		if _, err := fmt.Fscanf(p.reader, "$%d\r\n", &argLen); err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprintf("$%d\r\n", argLen))

		arg := make([]byte, argLen)
		if _, err := io.ReadFull(p.reader, arg); err != nil {
			return nil, err
		}
		buf.Write(arg)
		args = append(args, string(arg))

		if _, err := p.reader.Read(make([]byte, 2)); err != nil {
			return nil, err
		}
		buf.WriteString("\r\n")
	}

	return &Command{
		Name:    string(cmd),
		Message: []byte(buf.String()),
		Args:    args,
	}, nil
}

func (p *Parser) parseBulkString(buf *strings.Builder) (*Command, error) {
	var length int
	if _, err := fmt.Fscanf(p.reader, "%d\r\n", &length); err != nil {
		return nil, err
	}
	buf.WriteString(fmt.Sprintf("%d\r\n", length))

	if length == -1 {
		return &Command{
			Name:    "nil",
			Message: []byte(buf.String()),
		}, nil
	}

	str := make([]byte, length)
	if _, err := io.ReadFull(p.reader, str); err != nil {
		return nil, err
	}
	buf.Write(str)

	if _, err := p.reader.Read(make([]byte, 2)); err != nil {
		return nil, err
	}
	buf.WriteString("\r\n")

	return &Command{
		Name:    string(str),
		Message: []byte(buf.String()),
	}, nil
}

func (p *Parser) parseSimpleString(buf *strings.Builder) (*Command, error) {
	line, err := p.readUntilCRLF()
	if err != nil {
		return nil, err
	}
	buf.Write(line)
	buf.WriteString("\r\n")
	return &Command{
		Name:    string(line),
		Message: []byte(buf.String()),
	}, nil
}

func (p *Parser) parseError(buf *strings.Builder) (*Command, error) {
	line, err := p.readUntilCRLF()
	if err != nil {
		return nil, err
	}
	buf.Write(line)
	buf.WriteString("\r\n")
	return &Command{
		Name:    "ERROR: " + string(line),
		Message: []byte(buf.String()),
	}, nil
}

func (p *Parser) parseInteger(buf *strings.Builder) (*Command, error) {
	line, err := p.readUntilCRLF()
	if err != nil {
		return nil, err
	}
	buf.Write(line)
	buf.WriteString("\r\n")
	return &Command{
		Name:    string(line),
		Message: []byte(buf.String()),
	}, nil
}

func (p *Parser) readUntilCRLF() ([]byte, error) {
	var line []byte
	for {
		b := make([]byte, 1)
		if _, err := p.reader.Read(b); err != nil {
			return nil, err
		}
		line = append(line, b[0])
		if len(line) >= 2 && line[len(line)-2] == '\r' && line[len(line)-1] == '\n' {
			return line[:len(line)-2], nil
		}
	}
} 