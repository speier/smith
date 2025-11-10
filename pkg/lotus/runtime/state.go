package runtime

import (
	"encoding/gob"
	"os"
)

// Stateful is an optional interface for apps that need custom state serialization
// If not implemented, the library will use automatic gob encoding
type Stateful interface {
	// SaveState serializes the app's current state to bytes (optional - override default)
	SaveState() ([]byte, error)

	// LoadState restores the app's state from bytes (optional - override default)
	LoadState([]byte) error
}

// SaveAppState automatically saves an app's state using gob encoding
// Falls back to custom SaveState() if the app implements Stateful
func SaveAppState(app App, path string) error {
	var data []byte
	var err error

	// Check if app provides custom serialization
	if stateful, ok := app.(Stateful); ok {
		data, err = stateful.SaveState()
	} else {
		// Automatic serialization using gob
		var buf []byte
		enc := gob.NewEncoder(&gobBuffer{buf: &buf})
		err = enc.Encode(app)
		data = buf
	}

	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// LoadAppState automatically loads an app's state using gob decoding
// Falls back to custom LoadState() if the app implements Stateful
func LoadAppState(app App, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Check if app provides custom deserialization
	if stateful, ok := app.(Stateful); ok {
		return stateful.LoadState(data)
	}

	// Automatic deserialization using gob
	dec := gob.NewDecoder(&gobBuffer{buf: &data})
	return dec.Decode(app)
}

// gobBuffer is a helper for gob encoding/decoding to/from byte slices
type gobBuffer struct {
	buf *[]byte
	pos int
}

func (b *gobBuffer) Write(p []byte) (n int, err error) {
	*b.buf = append(*b.buf, p...)
	return len(p), nil
}

func (b *gobBuffer) Read(p []byte) (n int, err error) {
	if b.pos >= len(*b.buf) {
		return 0, os.ErrClosed
	}
	n = copy(p, (*b.buf)[b.pos:])
	b.pos += n
	return n, nil
}
