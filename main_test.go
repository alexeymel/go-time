package main

import (
	"testing"
)

type SetTimeTestPair struct {
	value  string
	result int64
}

var setTimeTests = []SetTimeTestPair{
	{"200414.083500.000", 1586853300000000000},
	{"200414.083500.123", 1586853300123000000},
	{"140625.235959.999", 1403740799999000000},
}

func TestDateTime_SetTime(t *testing.T) {
	testDateTime := new(DateTime)

	for _, pair := range setTimeTests {
		if testDateTime.SetTime(pair.value) == true {
			resultUnixNanoTime := testDateTime.time.UnixNano()

			if pair.result != resultUnixNanoTime {
				t.Error(
					"For", pair.value,
					"expected", pair.result,
					"got", resultUnixNanoTime,
				)
			}
		} else {
			t.Error("Can't set time", pair.value)
		}
	}
}

type SetDateTimePartsTestPair struct {
	value  string
	result []int
}

var setDateTimePartsTests = []SetDateTimePartsTestPair{
	{"200414.083500.000", []int{20, 4, 14, 8, 35, 0, 0}},
	{"200414.083500.123", []int{20, 4, 14, 8, 35, 0, 123}},
	{"140625.235959.999", []int{14, 6, 25, 23, 59, 59, 999}},
}

func TestDateTime_SetDateTimeParts(t *testing.T) {
	testDateTime := new(DateTime)

	for _, pair := range setDateTimePartsTests {
		if testDateTime.SetDateTimeParts(pair.value) == true {
			if len(pair.result) != len(testDateTime.parts) {
				t.Error(
					"For", pair.value,
					"expected len", len(pair.result),
					"got", len(testDateTime.parts),
				)
			}

			for index, value := range testDateTime.parts {
				if value != pair.result[index] {
					t.Error(
						"For", pair.value,
						"element", index,
						"expected", pair.result[index],
						"got", value,
					)
				}
			}
		} else {
			t.Error("Can't compile regExp")
		}
	}
}

type Float64TestPair struct {
	value  string
	result float64
}

var float64Tests = []Float64TestPair{
	{"200414.083500", 200414.0835},
	{"200414.083500123", 200414.0835},
	{"625.23", 625.23},
}

func TestDateTime_Float64(t *testing.T) {
	testDateTime := new(DateTime)

	for _, pair := range float64Tests {
		if testDateTime.ParseToTime(pair.value) == true {
			resultFloat64, resultError := testDateTime.Float64()

			if resultError != nil {
				t.Error(
					"For", pair.value,
					"can't convert to float64", resultError,
				)
			}

			if pair.result != resultFloat64 {
				t.Error(
					"For", pair.value,
					"expected", pair.result,
					"got", resultFloat64,
				)
			}
		}
	}
}

type StringTestPair struct {
	value  string
	result string
}

var stringTests = []StringTestPair{
	{"200414.083500", "Tue, 14 Apr 2020 08:35:00 +0000"},
	{"200414.083500123", "Tue, 14 Apr 2020 08:35:00 +0000"},
	{"625.23", "Sun, 25 Jun 2000 23:00:00 +0000"},
}

func TestDateTime_String(t *testing.T) {
	testDateTime := new(DateTime)

	for _, pair := range stringTests {
		if testDateTime.ParseToTime(pair.value) == true {
			resultString := testDateTime.String()

			if pair.result != resultString {
				t.Error(
					"For", pair.value,
					"expected", pair.result,
					"got", resultString,
				)
			}
		}
	}
}

var timeCorrectTests = []string{
	"200414.083500",
	"200414.083500123",
	"300414",
	"625.23",
}

func TestDateTime_TimeCorrect(t *testing.T) {
	testDateTime := new(DateTime)

	for _, value := range timeCorrectTests {
		if testDateTime.ParseToTime(value) == true {
			testDateTime.TimeCorrect()

			if testDateTime.serverTimeOffset == 0 {
				t.Error(
					"For", value,
					"expected", "<> 0",
					"got", testDateTime.serverTimeOffset,
				)
			}
		}
	}
}

type NormalizeFloat64DateTimeStringPair struct {
	value  string
	result string
}

var normalizeFloat64DateTimeStringTests = []NormalizeFloat64DateTimeStringPair{
	{"200414.083500", "200414.083500.000"},
	{"200414.083500123", "200414.083500.123"},
	{"625.23", "000625.230000.000"},
	{"0", "000000.000000.000"},
}

func TestDateTime_NormalizeFloat64DateTimeString(t *testing.T) {
	testDateTime := new(DateTime)

	for _, pair := range normalizeFloat64DateTimeStringTests {
		resultNormalizedString := testDateTime.NormalizeFloat64DateTimeString(pair.value)

		if pair.result != resultNormalizedString {
			t.Error(
				"For", pair.value,
				"expected", pair.result,
				"got", resultNormalizedString,
			)
		}
	}
}

func TestDateTime_InitOffset(t *testing.T) {
	testDateTime := new(DateTime)
	testDateTime.serverTimeOffset = 123456
	testDateTime.InitOffset()

	if testDateTime.serverTimeOffset != 0 {
		t.Error(
			"Expected", 0,
			"got", testDateTime.serverTimeOffset,
		)
	}
}

var dumpRestoreOffsetTests = []int64{
	0,
	1586853300000000000,
	1586853300123000000,
	1403740799999000000,
}

func TestDateTime_DumpOffset(t *testing.T) {
	testDateTime := new(DateTime)

	for _, value := range dumpRestoreOffsetTests {
		testDateTime.serverTimeOffset = value
		testDateTime.DumpOffset()
		testDateTime.serverTimeOffset = -1

		resultOffset := testDateTime.RestoreOffset().serverTimeOffset

		if resultOffset != value {
			t.Error(
				"Expected", value,
				"got", resultOffset,
			)
		}
	}
}

func TestDateTime_SetTimeOffset(t *testing.T) {
	testDateTime := new(DateTime)

	testDateTime.InitOffset()
	testUnixNow := testDateTime.SetTimeOffset().time.Unix()
	testDateTime.serverTimeOffset = 1586853300000000000
	resultUnix := testDateTime.SetTimeOffset().time.Unix()

	if resultUnix == testUnixNow {
		t.Error("Set time offset filed")
	}
}

type DeltaTestTriple struct {
	time   string
	delta  string
	result float64
}

var deltaTests = []DeltaTestTriple{
	{"200414.083500", "36", 200520.0835},
	{"200414.083500123", "-0.000000500", 200414.083459},
	{"625.23", "200000", 200625.23},
	{"201231.235959", "0.000001", 210101},
}

func TestDateTime_Delta(t *testing.T) {
	testDateTime := new(DateTime)

	for _, triple := range deltaTests {
		testDateTime.ParseToTime(triple.time)

		testDeltaDateTime := new(DateTime)
		testDeltaDateTime.ParseToParts(triple.delta)

		resultFloat64, resultError := testDateTime.Delta(testDeltaDateTime).Float64()

		if resultError == nil {
			if triple.result != resultFloat64 {
				t.Error(
					"For", triple.time,
					"add delta", triple.delta,
					"expected", triple.result,
					"got", resultFloat64,
				)
			}
		}
	}
}
