package fit

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"
	"unicode/utf8"

	"github.com/tormoder/fit/internal/types"
)

type encoder struct {
	w    io.Writer
	arch binary.ByteOrder
}

func encodeString(str string, size byte) ([]byte, error) {
	length := len(str)
	if length > int(size)-1 {
		length = int(size) - 1
	}

	bstr := make([]byte, size)
	copy(bstr, str[:length])
	if !utf8.Valid(bstr) {
		return nil, fmt.Errorf("Can't encode %+v as UTF-8 string", str)
	}
	return bstr, nil
}

func (e *encoder) writeField(value interface{}, f *field) error {
	if f.t.Array() {
		return fmt.Errorf("Can't encode Arrays")
	}

	switch f.t.Kind() {
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
		if f.t.BaseType() == types.BaseString {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("Not a string: %+v", value)
			}

			var err error
			value, err = encodeString(str, f.length)
			if err != nil {
				return fmt.Errorf("Can't encode %+v as UTF-8 string: %v", value, err)
			}
		}
		binary.Write(e.w, e.arch, value)
	default:
		return fmt.Errorf("Unknown Fit type %+v", f.t)
	}

	return nil
}

type encodeMesgDef struct {
	globalMesgNum MesgNum
	localMesgNum  byte
	fields        []*field
}

func (e *encoder) writeMesg(mesg reflect.Value, def *encodeMesgDef) error {
	hdr := def.localMesgNum & 0xF
	err := binary.Write(e.w, e.arch, hdr)
	if err != nil {
		return err
	}

	for _, f := range def.fields {
		value := mesg.Field(f.sindex).Interface()

		err := e.writeField(value, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func profileFieldDef(m MesgNum) [256]*field {
	return _fields[m]
}

func getFieldBySindex(index int, fields [256]*field) *field {
	for _, f := range fields {
		if f != nil && index == f.sindex {
			return f
		}
	}

	return fields[255]
}

// getEncodeMesgDef generates an appropriate encodeMesgDef to will encode all
// of the valid fields in mesg. Any fields which are set to their respective
// invalid value will be skipped (not present in the returned encodeMesgDef)
func getEncodeMesgDef(mesg reflect.Value, localMesgNum byte) *encodeMesgDef {
	mesgNum := getGlobalMesgNum(mesg.Type())
	allInvalid := getMesgAllInvalid(mesgNum)
	profileFields := profileFieldDef(mesgNum)

	if mesg.NumField() != allInvalid.NumField() {
		panic(fmt.Sprintf("Mismatched number of fields in type %+v", mesg.Type()))
	}

	def := &encodeMesgDef{
		globalMesgNum: mesgNum,
		localMesgNum:  localMesgNum,
		fields:        make([]*field, 0, mesg.NumField()),
	}

	for i := 0; i < mesg.NumField(); i++ {
		if mesg.Field(i).Interface() == allInvalid.Field(i).Interface() {
			// Don't encode invalid values
			continue
		}

		// FIXME: No message can exceed 255 bytes

		field := getFieldBySindex(i, profileFields)
		def.fields = append(def.fields, field)
	}

	return def
}

func (e *encoder) writeDefMesg(def *encodeMesgDef) error {
	hdr := (1 << 6) | def.localMesgNum&0xF
	err := binary.Write(e.w, e.arch, hdr)
	if err != nil {
		return err
	}

	err = binary.Write(e.w, e.arch, byte(0))
	if err != nil {
		return err
	}

	switch e.arch {
	case binary.LittleEndian:
		err = binary.Write(e.w, e.arch, byte(0))
	case binary.BigEndian:
		err = binary.Write(e.w, e.arch, byte(1))
	}
	if err != nil {
		return err
	}

	err = binary.Write(e.w, e.arch, def.globalMesgNum)
	if err != nil {
		return err
	}

	err = binary.Write(e.w, e.arch, byte(len(def.fields)))
	if err != nil {
		return err
	}

	for _, f := range def.fields {
		if f.t.Array() {
			return fmt.Errorf("TODO: Arrays not supported")
		}

		fdef := fieldDef{
			num:   f.num,
			size:  byte(f.t.BaseType().Size()),
			btype: f.t.BaseType(),
		}
		if fdef.btype == types.BaseString {
			fdef.size = f.length
		}

		err := binary.Write(e.w, e.arch, fdef)
		if err != nil {
			return err
		}
	}

	return nil
}
