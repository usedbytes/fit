package fit

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"
	"unicode/utf8"

	"github.com/tormoder/fit/dyncrc16"
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

func (e *encoder) encodeDefAndDataMesg(mesg reflect.Value) error {
	// We'll always just use local ID 0, for simplicity
	// We know the full file contents up-front, so no need to interleave
	def := getEncodeMesgDef(mesg, 0)

	err := e.writeDefMesg(def)
	if err != nil {
		return err
	}

	err = e.writeMesg(mesg, def)
	if err != nil {
		return err
	}

	return err
}

func (e *encoder) encodeFile(file reflect.Value) error {
	for i := 0; i < file.NumField(); i++ {
		v := file.Field(i)
		switch v.Kind() {
		case reflect.Struct, reflect.Ptr:
			err := e.encodeDefAndDataMesg(reflect.Indirect(v))
			if err != nil {
				return err
			}
		case reflect.Slice:
			var def *encodeMesgDef
			for j := 0; j < v.Len(); j++ {
				v2 := reflect.Indirect(v.Index(j))

				if j == 0 {
					def = getEncodeMesgDef(v2, 0)
					err := e.writeDefMesg(def)
					if err != nil {
						return err
					}
				}

				err := e.writeMesg(v2, def)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Encode writes the given FIT file into the given Writer. file.CRC and
// file.Header.CRC will be updated to the correct values.
func Encode(w io.Writer, file *File, arch binary.ByteOrder) error {
	buf := &bytes.Buffer{}
	enc := &encoder{
		w:    buf,
		arch: arch,
	}

	// XXX: Is there a better way to do this with reflection?
	var data reflect.Value
	switch file.Type() {
	case FileTypeActivity:
		activity, err := file.Activity()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*activity)
	case FileTypeDevice:
		device, err := file.Device()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*device)
	case FileTypeSettings:
		settings, err := file.Settings()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*settings)
	case FileTypeSport:
		sport, err := file.Sport()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*sport)
	case FileTypeWorkout:
		workout, err := file.Workout()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*workout)
	case FileTypeCourse:
		course, err := file.Course()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*course)
	case FileTypeSchedules:
		schedules, err := file.Schedules()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*schedules)
	case FileTypeWeight:
		weight, err := file.Weight()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*weight)
	case FileTypeTotals:
		totals, err := file.Totals()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*totals)
	case FileTypeGoals:
		goals, err := file.Goals()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*goals)
	case FileTypeBloodPressure:
		bloodPressure, err := file.BloodPressure()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*bloodPressure)
	case FileTypeMonitoringA:
		monitoringA, err := file.MonitoringA()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*monitoringA)
	case FileTypeActivitySummary:
		activitySummary, err := file.ActivitySummary()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*activitySummary)
	case FileTypeMonitoringDaily:
		monitoringDaily, err := file.MonitoringDaily()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*monitoringDaily)
	case FileTypeMonitoringB:
		monitoringB, err := file.MonitoringB()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*monitoringB)
	case FileTypeSegment:
		segment, err := file.Segment()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*segment)
	case FileTypeSegmentList:
		segmentList, err := file.SegmentList()
		if err != nil {
			return err
		}
		data = reflect.ValueOf(*segmentList)
	}

	// Encode the data
	err := enc.encodeDefAndDataMesg(reflect.ValueOf(file.FileId))
	if err != nil {
		return err
	}

	err = enc.encodeFile(data)
	if err != nil {
		return err
	}

	file.Header.DataSize = uint32(buf.Len())
	hdr, err := file.Header.MarshalBinary()
	if err != nil {
		return err
	}

	// Calculate file CRC
	crc := dyncrc16.New()

	_, err = crc.Write(hdr)
	if err != nil {
		return err
	}

	_, err = crc.Write(buf.Bytes())
	if err != nil {
		return err
	}

	file.CRC = crc.Sum16()

	// Write out the data
	_, err = w.Write(hdr)
	if err != nil {
		return err
	}

	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, file.CRC)
	if err != nil {
		return err
	}

	return nil
}
