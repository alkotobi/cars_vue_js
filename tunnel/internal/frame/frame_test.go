package frame_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/frame"
)

func TestHeaderRoundTripAllTypes(t *testing.T) {
	types := []frame.FrameType{
		frame.FrameHTTPReqStart,
		frame.FrameHTTPReqBody,
		frame.FrameHTTPReqEnd,
		frame.FrameHTTPRespStart,
		frame.FrameHTTPRespBody,
		frame.FrameHTTPRespEnd,
		frame.FrameWSData,
		frame.FrameWSClose,
		frame.FramePing,
		frame.FramePong,
	}
	for i, ft := range types {
		h := frame.Header{
			Type:       ft,
			RequestID:  uint64(42 + i),
			PayloadLen: uint32(1024 + i),
		}
		var buf bytes.Buffer
		if err := frame.WriteHeader(&buf, h); err != nil {
			t.Fatalf("WriteHeader(%v): %v", ft, err)
		}
		got, err := frame.ReadHeader(&buf)
		if err != nil {
			t.Fatalf("ReadHeader(%v): %v", ft, err)
		}
		if got.Type != h.Type || got.RequestID != h.RequestID || got.PayloadLen != h.PayloadLen {
			t.Fatalf("round-trip header mismatch: got %+v, want %+v", got, h)
		}
	}
}

func TestStartPayloadRoundTripRequestAndResponse(t *testing.T) {
	headers := make([]frame.HeaderField, 0, config.MaxHeaderCount)
	for i := 0; i < config.MaxHeaderCount; i++ {
		headers = append(headers, frame.HeaderField{
			Key:   "X-Header-" + string(rune('A'+i%26)),
			Value: "value",
		})
	}

	req := frame.StartPayload{
		Method:  "GET",
		Path:    "/cars",
		Status:  0,
		Headers: headers,
	}
	resp := frame.StartPayload{
		Method:  "",
		Path:    "",
		Status:  200,
		Headers: headers,
	}

	for _, s := range []frame.StartPayload{req, resp} {
		data, err := frame.MarshalStart(s)
		if err != nil {
			t.Fatalf("MarshalStart: %v", err)
		}
		got, err := frame.UnmarshalStart(data)
		if err != nil {
			t.Fatalf("UnmarshalStart: %v", err)
		}
		if s.Method != got.Method || s.Path != got.Path || s.Status != got.Status {
			t.Fatalf("round-trip mismatch: got %+v, want %+v", got, s)
		}
		if len(s.Headers) != len(got.Headers) {
			t.Fatalf("header count mismatch: got %d, want %d", len(got.Headers), len(s.Headers))
		}
		for i := range s.Headers {
			if s.Headers[i] != got.Headers[i] {
				t.Fatalf("header[%d] mismatch: got %+v, want %+v", i, got.Headers[i], s.Headers[i])
			}
		}
	}
}

func TestStartPayloadLargePathRoundTrip(t *testing.T) {
	path := bytes.Repeat([]byte("a"), 4090)
	s := frame.StartPayload{
		Method: "GET",
		Path:   string(path),
		Status: 0,
	}
	data, err := frame.MarshalStart(s)
	if err != nil {
		t.Fatalf("MarshalStart: %v", err)
	}
	got, err := frame.UnmarshalStart(data)
	if err != nil {
		t.Fatalf("UnmarshalStart: %v", err)
	}
	if got.Path != s.Path {
		t.Fatalf("path mismatch: got len=%d, want len=%d", len(got.Path), len(s.Path))
	}
}

func TestHeaderValueWithColonRoundTrip(t *testing.T) {
	s := frame.StartPayload{
		Method: "GET",
		Path:   "/",
		Status: 0,
		Headers: []frame.HeaderField{
			{Key: "Authorization", Value: "Bearer token:with:colons"},
		},
	}
	data, err := frame.MarshalStart(s)
	if err != nil {
		t.Fatalf("MarshalStart: %v", err)
	}
	got, err := frame.UnmarshalStart(data)
	if err != nil {
		t.Fatalf("UnmarshalStart: %v", err)
	}
	if len(got.Headers) != 1 || got.Headers[0].Value != s.Headers[0].Value {
		t.Fatalf("header value mismatch: got %+v, want %+v", got.Headers, s.Headers)
	}
}

func TestReadHeaderPayloadTooLarge(t *testing.T) {
	var buf bytes.Buffer
	// type
	buf.WriteByte(byte(frame.FrameHTTPReqBody))
	// request ID
	tmp8 := make([]byte, 8)
	// zero ID is fine
	buf.Write(tmp8)
	// payload len = FrameMaxPayload + 1
	tmp4 := make([]byte, 4)
	bigLen := uint32(config.FrameMaxPayload + 1)
	binary.BigEndian.PutUint32(tmp4, bigLen)
	buf.Write(tmp4)

	_, err := frame.ReadHeader(&buf)
	if err != frame.ErrPayloadTooLarge {
		t.Fatalf("ReadHeader payload too large: got %v, want %v", err, frame.ErrPayloadTooLarge)
	}
}

func TestReadHeaderTruncated(t *testing.T) {
	data := make([]byte, config.FrameHeaderSize-1)
	_, err := frame.ReadHeader(bytes.NewReader(data))
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("ReadHeader truncated: got %v, want %v", err, io.ErrUnexpectedEOF)
	}
}

func TestUnmarshalStartTooManyHeaders(t *testing.T) {
	// method_len=0, path_len=0, status=0, hdr_count=MaxHeaderCount+1
	b := []byte{
		0,       // method_len
		0, 0,    // path_len
		0, 0,    // status
		byte(config.MaxHeaderCount + 1), // hdr_count
	}
	_, err := frame.UnmarshalStart(b)
	if err != frame.ErrTooManyHeaders {
		t.Fatalf("UnmarshalStart too many headers: got %v, want %v", err, frame.ErrTooManyHeaders)
	}
}

func FuzzUnmarshalStart(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		// Fuzzing must never panic.
		_, _ = frame.UnmarshalStart(data)
	})
}

