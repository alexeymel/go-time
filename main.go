package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const ServerTimeOffsetFilename string = "sever_time_offset.txt"

type DateTime struct {
	serverTimeOffset int64
	parts            []int
	time             time.Time
	minus            bool
}

func (dateTime *DateTime) InitOffset() {
	dateTime.serverTimeOffset = 0
	dateTime.DumpOffset()
}

func (dateTime DateTime) DumpOffset() bool {
	dumpServerTimeOffsetError := ioutil.WriteFile(ServerTimeOffsetFilename, []byte(fmt.Sprintf("%d", dateTime.serverTimeOffset)), 0664)

	if dumpServerTimeOffsetError != nil {
		fmt.Printf("Can't dump server time offset to file \"%s\"\n", ServerTimeOffsetFilename)

		return false
	}

	return true
}

func (dateTime DateTime) RestoreOffset() DateTime {
	content, readServerTimeOffsetError := ioutil.ReadFile(ServerTimeOffsetFilename)

	if readServerTimeOffsetError == nil {
		serverTimeOffset, restoreServerTimeOffsetError := strconv.ParseInt(string(content), 10, 64)

		if restoreServerTimeOffsetError == nil {
			dateTime.serverTimeOffset = serverTimeOffset
		} else {
			fmt.Println("Can't parse server time offset from dump")
			dateTime.InitOffset()
		}
	} else {
		fmt.Printf("Can't restore server time offset from file \"%s\"\n", ServerTimeOffsetFilename)
		dateTime.InitOffset()
	}

	return dateTime
}

func (dateTime DateTime) Float64() (float64DateTime float64, convertError error) {
	return strconv.ParseFloat(dateTime.time.Format("060102.150405"), 64)
}

func (dateTime DateTime) String() (stringDateTime string) {
	return dateTime.time.Format(time.RFC1123Z)
}

func (dateTime DateTime) SetTimeOffset() DateTime {
	dateTime.time = time.Unix(0, time.Now().UTC().UnixNano()+dateTime.serverTimeOffset).UTC()

	return dateTime
}

func (dateTime *DateTime) TimeCorrect() bool {
	dateTime.serverTimeOffset = dateTime.time.UnixNano() - time.Now().UTC().UnixNano()

	return dateTime.DumpOffset()
}

func (dateTime *DateTime) ParseToTime(float64DateTimeString string) bool {
	float64DateTimeString = dateTime.NormalizeFloat64DateTimeString(float64DateTimeString)

	return dateTime.SetTime(float64DateTimeString)
}

func (dateTime *DateTime) ParseToParts(float64DateTimeString string) bool {
	float64DateTimeString = dateTime.NormalizeFloat64DateTimeString(float64DateTimeString)

	return dateTime.SetDateTimeParts(float64DateTimeString)
}

func (dateTime *DateTime) NormalizeFloat64DateTimeString(float64DateTimeString string) string {
	float64DateTime, convertError := strconv.ParseFloat(float64DateTimeString, 64)

	if convertError == nil {
		dateTime.minus = float64DateTime < 0
		intPart, fracPart := math.Modf(math.Abs(float64DateTime))
		float64DateTimeParts := []string{fmt.Sprintf("%06d", int(intPart))}
		fracParts := strings.Split(fmt.Sprintf("%.9f", fracPart), ".")
		filledFracPart := fracParts[1]
		float64DateTimeParts = append(float64DateTimeParts,
			filledFracPart[:len(filledFracPart)-3], filledFracPart[len(filledFracPart)-3:])
		float64DateTimeString = strings.Join(float64DateTimeParts, ".")
	}

	return float64DateTimeString
}

func (dateTime *DateTime) SetDateTimeParts(float64DateTimeString string) bool {
	dateTimeParts := []int{0, 0, 0, 0, 0, 0, 0}
	float64DateTimeRegExp, compileRegexpError := regexp.Compile(`(\d{2})(\d{2})(\d{2})\.(\d{2})(\d{2})(\d{2})\.(\d{3})`)

	if compileRegexpError == nil {
		parsedDateTime := float64DateTimeRegExp.FindStringSubmatch(float64DateTimeString)
		if len(parsedDateTime) == len(dateTimeParts)+1 {
			parsedDateTime = parsedDateTime[1:]

			for index, value := range parsedDateTime {
				int64Value, valueError := strconv.ParseInt(value, 10, 64)

				if valueError != nil {
					int64Value = 0
				}

				dateTimeParts[index] = int(int64Value)
			}

			dateTime.parts = dateTimeParts

			return true
		}
	}

	return false
}

func (dateTime *DateTime) SetTime(float64DateTimeString string) bool {
	parsedTime, parseTimeError := time.Parse("060102.150405.000", float64DateTimeString)

	if parseTimeError == nil {
		dateTime.time = parsedTime

		return true
	}

	return false
}

func (dateTime DateTime) Delta(deltaTime *DateTime) DateTime {
	sign := ""
	y, m, d := deltaTime.parts[0], deltaTime.parts[1], deltaTime.parts[2]

	if deltaTime.minus == true {
		sign = "-"
		y, m, d = 0-y, 0-m, 0-d
	}

	dateTime.time = dateTime.time.AddDate(y, m, d)
	duration, durationError := time.ParseDuration(fmt.Sprintf(
		"%s%dh%dm%ds%dms", sign, deltaTime.parts[3], deltaTime.parts[4], deltaTime.parts[5], deltaTime.parts[6]))
	dateTime.time = dateTime.time.Add(duration)

	if durationError != nil {
		fmt.Println("Can't create duration delta")
	}

	return dateTime
}

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		var responseContent interface{}

		switch request.Method + request.URL.Path {
		case "GET/time/now":
			timeNow, timeNowError := new(DateTime).RestoreOffset().SetTimeOffset().Float64()

			if timeNowError == nil {
				responseContent = map[string]interface{}{
					"time": timeNow,
				}
			} else {
				http.Error(writer, "Failed to get time", http.StatusInternalServerError)
			}
			break
		case "GET/time/string":
			parameterTime := request.URL.Query().Get("time")
			dateTime := new(DateTime)

			if dateTime.ParseToTime(parameterTime) == true {
				responseContent = map[string]interface{}{
					"str": dateTime.String(),
				}
			} else {
				http.Error(writer, "Incorrect \"time\" format", http.StatusBadRequest)
			}
			break
		case "GET/time/add":
			parameterTime := request.URL.Query().Get("time")
			parameterDelta := request.URL.Query().Get("delta")

			dateTime := new(DateTime)
			deltaTime := new(DateTime)

			if dateTime.ParseToTime(parameterTime) == true {
				if deltaTime.ParseToParts(parameterDelta) == true {
					timeSum, timeSumError := dateTime.Delta(deltaTime).Float64()

					if timeSumError == nil {
						responseContent = map[string]interface{}{
							"time": timeSum,
						}
					} else {
						http.Error(writer, "Failed to get time", http.StatusInternalServerError)
					}
				} else {
					http.Error(writer, "Incorrect \"delta\" format", http.StatusBadRequest)
				}
			} else {
				http.Error(writer, "Incorrect \"time\" format", http.StatusBadRequest)
			}
			break
		case "POST/time/correct":
			parameterTime := request.FormValue("time")
			dateTime := new(DateTime)

			if dateTime.ParseToTime(parameterTime) == true {
				if dateTime.TimeCorrect() == true {
					http.Error(writer, "Success", http.StatusOK)
				} else {
					http.Error(writer, "Can't correct time", http.StatusBadRequest)
				}
			} else {
				http.Error(writer, "Incorrect \"time\" format", http.StatusBadRequest)
			}
			break
		default:
			http.Error(writer, "Not Found", http.StatusNotFound)
			break
		}

		if responseContent != nil {
			writer.Header().Set("Content-Type", "application/json;charset=utf-8")
			jsonResponseContent, jsonError := json.Marshal(responseContent)

			if jsonError == nil {
				writeCount, writeError := writer.Write(jsonResponseContent)

				if writeError != nil && writeCount > 0 {
					fmt.Println("Failed to send response")
				}
			} else {
				http.Error(writer, "Failed to compose response content", http.StatusInternalServerError)
			}
		}
	})

	startServerError := http.ListenAndServe(":8080", nil)
	if startServerError != nil {
		fmt.Println("Server startup error")
	}
}
