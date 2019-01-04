package fit

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
	"time"

	"github.com/tormoder/fit/internal/types"
)

func TestEncodeWriteField(t *testing.T) {
	type writeFieldTest struct {
		btype types.Fit
		value interface{}
		le    []byte
		be    []byte
	}

	tests := []writeFieldTest{
		{
			btype: types.MakeNative(types.BaseEnum, false),
			value: byte(0x42),
			le:    []byte{0x42},
			be:    []byte{0x42},
		},
		{
			btype: types.MakeNative(types.BaseSint8, false),
			value: int8(-0x80),
			le:    []byte{0x80},
			be:    []byte{0x80},
		},
		{
			btype: types.MakeNative(types.BaseUint8, false),
			value: uint8(0x80),
			le:    []byte{0x80},
			be:    []byte{0x80},
		},
		{
			btype: types.MakeNative(types.BaseSint16, false),
			value: int16(-0x1234),
			le:    []byte{0xCC, 0xED},
			be:    []byte{0xED, 0xCC},
		},
		{
			btype: types.MakeNative(types.BaseUint16, false),
			value: uint16(0x1234),
			le:    []byte{0x34, 0x12},
			be:    []byte{0x12, 0x34},
		},
		{
			btype: types.MakeNative(types.BaseSint32, false),
			value: int32(-0x12345678),
			le:    []byte{0x88, 0xA9, 0xCB, 0xED},
			be:    []byte{0xED, 0xCB, 0xA9, 0x88},
		},
		{
			btype: types.MakeNative(types.BaseUint32, false),
			value: uint32(0x12345678),
			le:    []byte{0x78, 0x56, 0x34, 0x12},
			be:    []byte{0x12, 0x34, 0x56, 0x78},
		},
		{
			btype: types.MakeNative(types.BaseString, false),
			value: string("Hello"),
			le:    []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x00},
			be:    []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x00},
		},
		{
			btype: types.MakeNative(types.BaseFloat32, false),
			value: float32(3.142),
			le:    []byte{0x87, 0x16, 0x49, 0x40},
			be:    []byte{0x40, 0x49, 0x16, 0x87},
		},
		{
			btype: types.MakeNative(types.BaseFloat64, false),
			value: float64(3.142),
			le:    []byte{0x89, 0x41, 0x60, 0xE5, 0xD0, 0x22, 0x09, 0x40},
			be:    []byte{0x40, 0x09, 0x22, 0xD0, 0xE5, 0x60, 0x41, 0x89},
		},
		{
			btype: types.MakeNative(types.BaseUint8z, false),
			value: uint8(0x80),
			le:    []byte{0x80},
			be:    []byte{0x80},
		},
		{
			btype: types.MakeNative(types.BaseUint16z, false),
			value: uint16(0x1234),
			le:    []byte{0x34, 0x12},
			be:    []byte{0x12, 0x34},
		},
		{
			btype: types.MakeNative(types.BaseUint32z, false),
			value: uint32(0x12345678),
			le:    []byte{0x78, 0x56, 0x34, 0x12},
			be:    []byte{0x12, 0x34, 0x56, 0x78},
		},
		{
			btype: types.MakeNative(types.BaseByte, false),
			value: byte(0x42),
			le:    []byte{0x42},
			be:    []byte{0x42},
		},
		{
			btype: types.MakeNative(types.BaseSint64, false),
			value: int64(-0x12345678ABCDEF00),
			le:    []byte{0x00, 0x11, 0x32, 0x54, 0x87, 0xA9, 0xCB, 0xED},
			be:    []byte{0xED, 0xCB, 0xA9, 0x87, 0x54, 0x32, 0x11, 0x00},
		},
		{
			btype: types.MakeNative(types.BaseUint64, false),
			value: uint64(0x12345678ABCDEF00),
			le:    []byte{0x00, 0xEF, 0xCD, 0xAB, 0x78, 0x56, 0x34, 0x12},
			be:    []byte{0x12, 0x34, 0x56, 0x78, 0xAB, 0xCD, 0xEF, 0x00},
		},
		{
			btype: types.MakeNative(types.BaseUint64z, false),
			value: uint64(0x12345678ABCDEF00),
			le:    []byte{0x00, 0xEF, 0xCD, 0xAB, 0x78, 0x56, 0x34, 0x12},
			be:    []byte{0x12, 0x34, 0x56, 0x78, 0xAB, 0xCD, 0xEF, 0x00},
		},
		{
			btype: types.Make(types.TimeUTC, false),
			value: timeBase,
			le:    []byte{0x00, 0x00, 0x00, 0x00},
			be:    []byte{0x00, 0x00, 0x00, 0x00},
		},
		/* TODO
		{
			btype: types.Make(types.TimeLocal, false),
			value: timeBase,
			le: []byte{0x00, 0x00, 0x00, 0x00},
			be: []byte{0x00, 0x00, 0x00, 0x00},
		},
		*/
		{
			btype: types.Make(types.Lat, false),
			value: NewLatitudeDegrees(52.2053),
			le:    []byte{0x51, 0xAF, 0x1F, 0x25},
			be:    []byte{0x25, 0x1F, 0xAF, 0x51},
		},
		{
			btype: types.Make(types.Lng, false),
			value: NewLongitudeDegrees(0.1218),
			le:    []byte{0x4A, 0x2C, 0x16, 0x00},
			be:    []byte{0x00, 0x16, 0x2C, 0x4A},
		},
	}

	buf := &bytes.Buffer{}

	e := &encoder{
		w: buf,
	}

	e.arch = binary.LittleEndian
	for i, test := range tests {
		buf.Reset()

		err := e.writeField(test.value, test.btype)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(buf.Bytes(), test.le) {
			t.Errorf("LE %d (%s): Expected '%v' got '%v'", i, test.btype, test.le, buf.Bytes())
		}
	}

	e.arch = binary.BigEndian
	for i, test := range tests {
		buf.Reset()

		err := e.writeField(test.value, test.btype)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(buf.Bytes(), test.be) {
			t.Errorf("BE %d (%s): Expected '%v' got '%v'", i, test.btype, test.be, buf.Bytes())
		}
	}
}

func TestEncodeWriteMesg(t *testing.T) {
	type TestMesg struct {
		Type         byte
		Timestamp    time.Time
		PositionLat  Latitude
		PositionLong Longitude
	}

	mesg := TestMesg{
		Type:         0x10,
		Timestamp:    timeBase.Add(32 * time.Second),
		PositionLat:  NewLatitudeDegrees(50.2053),
		PositionLong: NewLongitudeDegrees(0.1218),
	}

	def := &encodeMesgDef{
		localMesgNum: 3,
		fields: []encodeFieldDef{
			{
				sindex: 0,
				num:    0,
				ftype:  types.MakeNative(types.BaseEnum, false),
			},
			{
				sindex: 1,
				num:    253,
				ftype:  types.Make(types.TimeUTC, false),
			},
			// PositionLat intentionally omitted
			{
				sindex: 3,
				num:    3,
				ftype:  types.Make(types.Lng, false),
			},
		},
	}

	expect := []byte{
		0x03,
		0x10,
		0x20, 0x00, 0x00, 0x00,
		0x4A, 0x2C, 0x16, 0x00,
	}

	buf := &bytes.Buffer{}

	e := &encoder{
		w:    buf,
		arch: binary.LittleEndian,
	}

	err := e.writeMesg(reflect.ValueOf(mesg), def)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(buf.Bytes(), expect) {
		t.Errorf("Expected '%v', got '%v'", expect, buf.Bytes())
	}
}
