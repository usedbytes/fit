package fit

import (
	"fmt"
	"reflect"
)

// File represents a decoded FIT file.
type File struct {
	// Header is the FIT file header.
	Header Header

	// CRC is the FIT file CRC.
	CRC uint16

	// FileId is a message required for all FIT files.
	FileId FileIdMsg

	// Common messages for all FIT file types.
	FileCreator          *FileCreatorMsg
	TimestampCorrelation *TimestampCorrelationMsg
	DeviceInfo           *DeviceInfoMsg

	// UnknownMessages is a slice of unknown messages encountered during
	// decoding. It is sorted by message number.
	UnknownMessages []UnknownMessage

	// UnknownFields is a slice of unknown fields for known messages
	// encountered during decoding. It is sorted by message number.
	UnknownFields []UnknownField

	msgAdder msgAdder

	activity        *ActivityFile
	device          *DeviceFile
	settings        *SettingsFile
	sport           *SportFile
	workout         *WorkoutFile
	course          *CourseFile
	schedules       *SchedulesFile
	weight          *WeightFile
	totals          *TotalsFile
	goals           *GoalsFile
	bloodPressure   *BloodPressureFile
	monitoringA     *MonitoringAFile
	activitySummary *ActivitySummaryFile
	monitoringDaily *MonitoringDailyFile
	monitoringB     *MonitoringBFile
	segment         *SegmentFile
	segmentList     *SegmentListFile
}

type msgAdder interface {
	add(reflect.Value)
}

func (f *File) add(msg reflect.Value) {
	x := msg.Interface()
	switch x.(type) {
	case FileIdMsg:
		f.FileId = x.(FileIdMsg)
	case FileCreatorMsg:
		tmp := x.(FileCreatorMsg)
		f.FileCreator = &tmp
	case TimestampCorrelationMsg:
		tmp := x.(TimestampCorrelationMsg)
		f.TimestampCorrelation = &tmp
	case DeviceInfoMsg:
		tmp := x.(DeviceInfoMsg)
		f.DeviceInfo = &tmp
	default:
		f.msgAdder.add(msg)
	}
}

func (f *File) init() error {
	t := f.FileId.Type
	switch t {
	case FileTypeActivity:
		f.activity = new(ActivityFile)
		f.msgAdder = f.activity
	case FileTypeDevice:
		f.device = new(DeviceFile)
		f.msgAdder = f.device
	case FileTypeSettings:
		f.settings = new(SettingsFile)
		f.msgAdder = f.settings
	case FileTypeSport:
		f.sport = new(SportFile)
		f.msgAdder = f.sport
	case FileTypeWorkout:
		f.workout = new(WorkoutFile)
		f.msgAdder = f.workout
	case FileTypeCourse:
		f.course = new(CourseFile)
		f.msgAdder = f.course
	case FileTypeSchedules:
		f.schedules = new(SchedulesFile)
		f.msgAdder = f.schedules
	case FileTypeWeight:
		f.weight = new(WeightFile)
		f.msgAdder = f.weight
	case FileTypeTotals:
		f.totals = new(TotalsFile)
		f.msgAdder = f.totals
	case FileTypeGoals:
		f.goals = new(GoalsFile)
		f.msgAdder = f.goals
	case FileTypeBloodPressure:
		f.bloodPressure = new(BloodPressureFile)
		f.msgAdder = f.bloodPressure
	case FileTypeMonitoringA:
		f.monitoringA = new(MonitoringAFile)
		f.msgAdder = f.monitoringA
	case FileTypeActivitySummary:
		f.activitySummary = new(ActivitySummaryFile)
		f.msgAdder = f.activitySummary
	case FileTypeMonitoringDaily:
		f.monitoringDaily = new(MonitoringDailyFile)
		f.msgAdder = f.monitoringDaily
	case FileTypeMonitoringB:
		f.monitoringB = new(MonitoringBFile)
		f.msgAdder = f.monitoringB
	case FileTypeSegment:
		f.segment = new(SegmentFile)
		f.msgAdder = f.segment
	case FileTypeSegmentList:
		f.segmentList = new(SegmentListFile)
		f.msgAdder = f.segmentList
	case FileTypeInvalid:
		return FormatError("file type was set invalid")
	default:
		switch {
		case t > FileTypeMonitoringB && t < FileTypeMfgRangeMin:
			return FormatError(
				fmt.Sprintf("unknown file type: %v", t),
			)
		case t >= FileTypeMfgRangeMin && t <= FileTypeMfgRangeMax:
			return NotSupportedError("manufacturer specific file types")
		default:
			return FormatError(
				fmt.Sprintf("unknown file type: %v", t),
			)
		}
	}
	return nil
}

// Type returns the FIT file type.
func (f *File) Type() FileType {
	return f.FileId.Type
}

type wrongFileTypeError struct {
	actual, requested FileType
}

func (e wrongFileTypeError) Error() string {
	return fmt.Sprintf("fit file type is %v, not %v", e.actual, e.requested)
}

// Activity returns f's Activity file. An error is returned if the FIT file is
// not of type activity.
func (f *File) Activity() (*ActivityFile, error) {
	if !(f.FileId.Type == FileTypeActivity) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeActivity}
	}
	return f.activity, nil
}

// Device returns f's Device file. An error is returned if the FIT file is
// not of type device.
func (f *File) Device() (*DeviceFile, error) {
	if !(f.FileId.Type == FileTypeDevice) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeDevice}
	}
	return f.device, nil
}

// Settings returns f's Settings file. An error is returned if the FIT file is
// not of type settings.
func (f *File) Settings() (*SettingsFile, error) {
	if !(f.FileId.Type == FileTypeSettings) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeSettings}
	}
	return f.settings, nil
}

// Sport returns f's Sport file. An error is returned if the FIT file is
// not of type sport.
func (f *File) Sport() (*SportFile, error) {
	if !(f.FileId.Type == FileTypeSport) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeSport}
	}
	return f.sport, nil
}

// Workout returns f's Workout file. An error is returned if the FIT file is
// not of type workout.
func (f *File) Workout() (*WorkoutFile, error) {
	if !(f.FileId.Type == FileTypeWorkout) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeWorkout}
	}
	return f.workout, nil
}

// Course returns f's Course file. An error is returned if the FIT file is
// not of type course.
func (f *File) Course() (*CourseFile, error) {
	if !(f.FileId.Type == FileTypeCourse) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeCourse}
	}
	return f.course, nil
}

// Schedules returns f's Schedules file. An error is returned if the FIT file is
// not of type schedules.
func (f *File) Schedules() (*SchedulesFile, error) {
	if !(f.FileId.Type == FileTypeSchedules) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeSchedules}
	}
	return f.schedules, nil
}

// Weight returns f's Weight file. An error is returned if the FIT file is
// not of type weight.
func (f *File) Weight() (*WeightFile, error) {
	if !(f.FileId.Type == FileTypeWeight) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeWeight}
	}
	return f.weight, nil
}

// Totals returns f's Totals file. An error is returned if the FIT file is
// not of type totals.
func (f *File) Totals() (*TotalsFile, error) {
	if !(f.FileId.Type == FileTypeTotals) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeTotals}
	}
	return f.totals, nil
}

// Goals returns f's Goals file. An error is returned if the FIT file is
// not of type goals.
func (f *File) Goals() (*GoalsFile, error) {
	if !(f.FileId.Type == FileTypeGoals) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeGoals}
	}
	return f.goals, nil
}

// BloodPressure returns f's BloodPressure file. An error is returned if the FIT file is
// not of type blood pressure.
func (f *File) BloodPressure() (*BloodPressureFile, error) {
	if !(f.FileId.Type == FileTypeBloodPressure) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeBloodPressure}
	}
	return f.bloodPressure, nil
}

// MonitoringA returns f's MonitoringA file. An error is returned if the FIT file is
// not of type monitoring A.
func (f *File) MonitoringA() (*MonitoringAFile, error) {
	if !(f.FileId.Type == FileTypeMonitoringA) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeMonitoringA}
	}
	return f.monitoringA, nil
}

// ActivitySummary returns f's ActivitySummary file. An error is returned if the FIT file is
// not of type activity summary.
func (f *File) ActivitySummary() (*ActivitySummaryFile, error) {
	if !(f.FileId.Type == FileTypeActivitySummary) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeActivitySummary}
	}
	return f.activitySummary, nil
}

// MonitoringDaily returns f's MonitoringDaily file. An error is returned if the FIT file is
// not of type monitoring daily.
func (f *File) MonitoringDaily() (*MonitoringDailyFile, error) {
	if !(f.FileId.Type == FileTypeMonitoringDaily) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeMonitoringDaily}
	}
	return f.monitoringDaily, nil
}

// MonitoringB returns f's MonitoringB file. An error is returned if the FIT file is
// not of type monitoring B.
func (f *File) MonitoringB() (*MonitoringBFile, error) {
	if !(f.FileId.Type == FileTypeMonitoringB) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeMonitoringB}
	}
	return f.monitoringB, nil
}

// Segment returns f's Segment file. An error is returned if the FIT file is
// not of type segment.
func (f *File) Segment() (*SegmentFile, error) {
	if !(f.FileId.Type == FileTypeSegment) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeSegment}
	}
	return f.segment, nil
}

// SegmentList returns f's SegmentList file. An error is returned if the FIT file is
// not of type segment list.
func (f *File) SegmentList() (*SegmentListFile, error) {
	if !(f.FileId.Type == FileTypeSegmentList) {
		return nil, wrongFileTypeError{f.FileId.Type, FileTypeSegmentList}
	}
	return f.segmentList, nil
}

/*
func (f File) MarshalBinary() ([]byte, error) {
	hdr, err := f.Header.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return hdr, err
}

func globalMesgNum(t reflect.Type) MesgNum {
	for i, match := range msgsTypes {
		if t == match {
			return MesgNum(i)
		}
	}

	return MesgNumInvalid
}

func profileFieldDef(m MesgNum) [256]*field {
	return _fields[m]
}

func getFieldBySindex(index int, fields [256]*field) *field {
	for _, f := range fields {
		if index == f.sindex {
			return f
		}
	}

	return fields[255]
}

func definitionMessage(value reflect.Value, arch binary.ByteOrder, localMsgType uint8) *definition {
	mesgNum := globalMesgNum(value.Type())
	profileFields := profileFieldDef(mesgNum)
	allInvalid := getMesgAllInvalid(mesgNum)

	if value.NumField() != allInvalid.NumField() {
		panic(fmt.Sprintf("Mismatched number of fields in type %+v", value.Type()))
	}

	def := &definition{
		msg: defmsg{
			localMsgType: 0,
			arch: arch,
			globalMsgNum: mesgNum,
			fieldDefs: make([]fieldDef, 0),
		},
		fields: make([]*field, value.NumField()),
	}

	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Interface() == allInvalid.Field(i).Interface() {
			continue
		}

		field := getFieldBySindex(i, profileFields)
		fd := fieldDef{
			num: field.num,
			size: byte(field.t.BaseType().Size()),
			btype: field.t.BaseType(),
		}

		dm.msg.fieldDefs = append(dm.fieldDefs, fd)
		dm.msg.fields++
		dm.fields[i] = field
	}

	return dm
}

func encodeField(w io.Writer, value interface{}, fdef *fieldDef, arch binary.ByteOrder) error {
	fmt.Printf("Encode %+v with %+v\n", value, fdef)

	var err error
	switch fdef.btype.Kind() {
	case types.TimeUTC, types.TimeLocal:
		fmt.Println("Don't know how to do time...")
	default:
		err = binary.Write(w, arch, value)
	}
	return err
}

func (e *encoder) encodeMesg(w io.Writer, value reflect.Value, def *definition) error {
	for i := 0; i < value.NumField(); i++ {
		if field := def.fields[i]; field != nil {
			err := e.writeField(value.Field(i).Interface(), field)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *File) Encode(w io.Writer) error {
	var e encoder
	e.w = w

	dm := definitionMessage(reflect.ValueOf(f.FileId), binary.LittleEndian, 0)
	fmt.Printf("%+v\n", dm)
	encodeMesg(e.w, reflect.ValueOf(f.FileId), dm)


	return nil
}
*/
