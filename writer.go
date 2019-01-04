package fit

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
	"unicode/utf8"

	"github.com/tormoder/fit/internal/types"
)

type encoder struct {
	w    io.Writer
	arch binary.ByteOrder
}

func encodeString(str string) ([]byte, error) {
	bstr := append([]byte(str), '\000')
	if !utf8.Valid(bstr) {
		return nil, fmt.Errorf("Can't encode %+v as UTF-8 string", str)
	}
	return bstr, nil
}

func (e *encoder) writeField(value interface{}, t types.Fit) error {
	switch t.Kind() {
	case types.TimeUTC:
		t := value.(time.Time)
		u32 := encodeTime(t)
		binary.Write(e.w, e.arch, u32)
	case types.TimeLocal:
		return fmt.Errorf("Can't encode TimeLocal")
	case types.Lat:
		lat := value.(Latitude)
		binary.Write(e.w, e.arch, lat.semicircles)
	case types.Lng:
		lng := value.(Longitude)
		binary.Write(e.w, e.arch, lng.semicircles)
	case types.NativeFit:
		if t.BaseType() == types.BaseString {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("Not a string: %+v", value)
			}

			var err error
			value, err = encodeString(str)
			if err != nil {
				return fmt.Errorf("Can't encode %+v as UTF-8 string: %v", value, err)
			}
		}
		binary.Write(e.w, e.arch, value)
	default:
		return fmt.Errorf("Unknown Fit type %+v", t)
	}

	return nil
}
