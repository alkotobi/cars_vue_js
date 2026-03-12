package frame

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/yourname/tunnel/internal/config"
)

// FrameType identifies the semantic type of a frame on the data channel.
type FrameType uint8

// Frame types for HTTP and WebSocket data and control frames.
const (
	FrameHTTPReqStart  FrameType = 0x01 // headers: method, path, req headers
	FrameHTTPReqBody   FrameType = 0x02 // raw body chunk
	FrameHTTPReqEnd    FrameType = 0x03 // end of request body (payload empty)
	FrameHTTPRespStart FrameType = 0x04 // headers: status, resp headers
	FrameHTTPRespBody  FrameType = 0x05 // raw response body chunk
	FrameHTTPRespEnd   FrameType = 0x06 // end of response (payload empty)
	FrameWSData        FrameType = 0x07 // WebSocket data frame
	FrameWSClose       FrameType = 0x08 // WebSocket close
	FramePing          FrameType = 0x09 // tunnel keepalive
	FramePong          FrameType = 0x0A // keepalive reply
)

// HeaderSize is the number of bytes in a frame header.
const HeaderSize = config.FrameHeaderSize

// Header is the decoded representation of the fixed-size frame header.
type Header struct {
	Type       FrameType
	RequestID  uint64
	PayloadLen uint32
}

// StartPayload carries the HTTP start metadata for requests and responses.
type StartPayload struct {
	Method  string
	Path    string
	Status  int
	Headers []HeaderField
}

// HeaderField is a single HTTP header key/value pair.
type HeaderField struct {
	Key   string
	Value string
}

// Sentinel errors for the frame package.
var (
	ErrInvalidFrame    = errors.New("frame: invalid frame")
	ErrPayloadTooLarge = errors.New("frame: payload too large")
	ErrTooManyHeaders  = errors.New("frame: too many headers")
	ErrHeaderTooLong   = errors.New("frame: header too long")
)

// ReadHeader reads exactly HeaderSize bytes from r and parses the header.
// It returns ErrPayloadTooLarge if the advertised payload length exceeds
// config.FrameMaxPayload.
func ReadHeader(r io.Reader) (Header, error) {
	var buf [HeaderSize]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		// Propagate io.ErrUnexpectedEOF directly for truncated headers.
		return Header{}, err
	}
	h := Header{
		Type:       FrameType(buf[0]),
		RequestID:  binary.BigEndian.Uint64(buf[1:9]),
		PayloadLen: binary.BigEndian.Uint32(buf[9:13]),
	}
	if h.PayloadLen > uint32(config.FrameMaxPayload) {
		return Header{}, ErrPayloadTooLarge
	}
	return h, nil
}

// WriteHeader encodes h and writes it to w.
func WriteHeader(w io.Writer, h Header) error {
	if h.PayloadLen > uint32(config.FrameMaxPayload) {
		return ErrPayloadTooLarge
	}
	var buf [HeaderSize]byte
	buf[0] = byte(h.Type)
	binary.BigEndian.PutUint64(buf[1:9], h.RequestID)
	binary.BigEndian.PutUint32(buf[9:13], h.PayloadLen)
	_, err := w.Write(buf[:])
	return err
}

// MarshalStart encodes a StartPayload into a byte slice using a compact
// length-prefixed binary format described in the package documentation.
func MarshalStart(s StartPayload) ([]byte, error) {
	if len(s.Headers) > config.MaxHeaderCount {
		return nil, ErrTooManyHeaders
	}
	if len(s.Method) > 0xFF {
		return nil, ErrHeaderTooLong
	}
	if len(s.Path) > 0xFFFF {
		return nil, ErrHeaderTooLong
	}
	for _, hf := range s.Headers {
		if len(hf.Key) > config.MaxHeaderKeyLen || len(hf.Key) > 0xFF {
			return nil, ErrHeaderTooLong
		}
		if len(hf.Value) > config.MaxHeaderValLen || len(hf.Value) > 0xFFFF {
			return nil, ErrHeaderTooLong
		}
	}

	// Rough size estimate: header plus all strings.
	size := 1 + 2 + 2 + 1 // method_len, path_len, status, hdr_count
	size += len(s.Method) + len(s.Path)
	for _, hf := range s.Headers {
		size += 1 + len(hf.Key) + 2 + len(hf.Value)
	}
	out := make([]byte, 0, size)

	out = append(out, uint8(len(s.Method)))
	out = append(out, []byte(s.Method)...)

	var tmp [2]byte
	binary.BigEndian.PutUint16(tmp[:], uint16(len(s.Path)))
	out = append(out, tmp[:]...)
	out = append(out, []byte(s.Path)...)

	binary.BigEndian.PutUint16(tmp[:], uint16(s.Status))
	out = append(out, tmp[:]...)

	out = append(out, uint8(len(s.Headers)))
	for _, hf := range s.Headers {
		out = append(out, uint8(len(hf.Key)))
		out = append(out, []byte(hf.Key)...)
		binary.BigEndian.PutUint16(tmp[:], uint16(len(hf.Value)))
		out = append(out, tmp[:]...)
		out = append(out, []byte(hf.Value)...)
	}

	return out, nil
}

// UnmarshalStart decodes a StartPayload from b, performing basic validation
// of header counts and key/value lengths.
func UnmarshalStart(b []byte) (StartPayload, error) {
	var s StartPayload
	if len(b) < 1+2+2+1 {
		return s, ErrInvalidFrame
	}
	i := 0

	methodLen := int(b[i])
	i++
	if len(b) < i+methodLen+2+2+1 {
		return s, ErrInvalidFrame
	}
	s.Method = string(b[i : i+methodLen])
	i += methodLen

	if len(b) < i+2 {
		return s, ErrInvalidFrame
	}
	pathLen := int(binary.BigEndian.Uint16(b[i : i+2]))
	i += 2
	if len(b) < i+pathLen+2+1 {
		return s, ErrInvalidFrame
	}
	s.Path = string(b[i : i+pathLen])
	i += pathLen

	if len(b) < i+2 {
		return s, ErrInvalidFrame
	}
	s.Status = int(binary.BigEndian.Uint16(b[i : i+2]))
	i += 2

	if len(b) < i+1 {
		return s, ErrInvalidFrame
	}
	hdrCount := int(b[i])
	i++
	if hdrCount > config.MaxHeaderCount {
		return s, ErrTooManyHeaders
	}

	s.Headers = make([]HeaderField, 0, hdrCount)
	for idx := 0; idx < hdrCount; idx++ {
		if len(b) < i+1 {
			return s, ErrInvalidFrame
		}
		keyLen := int(b[i])
		i++
		if keyLen > config.MaxHeaderKeyLen {
			return s, ErrHeaderTooLong
		}
		if len(b) < i+keyLen+2 {
			return s, ErrInvalidFrame
		}
		key := string(b[i : i+keyLen])
		i += keyLen

		valLen := int(binary.BigEndian.Uint16(b[i : i+2]))
		i += 2
		if valLen > config.MaxHeaderValLen {
			return s, ErrHeaderTooLong
		}
		if len(b) < i+valLen {
			return s, ErrInvalidFrame
		}
		val := string(b[i : i+valLen])
		i += valLen

		s.Headers = append(s.Headers, HeaderField{Key: key, Value: val})
	}

	return s, nil
}

