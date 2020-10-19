// Copyright © 2020 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package pd

import (
	"strconv"
	"strings"
	"time"
)

// Requirement: Shifts must be sorted in .pd.yml file

// Shift specifies time range and name of a shift, start and end times are saved in minutes since midnight
type Shift struct {
	Start int
	End   int
	Name  string
}

// GetProbablyOwnShift returns the shift the user probably belongs to because of their time zone
func GetProbablyOwnShift() (Shift, error) {

	timeOffsetString := time.Now().String()
	if offsetLocation := strings.Index(timeOffsetString, " +"); offsetLocation != -1 {
		timeOffsetString = timeOffsetString[offsetLocation+1 : offsetLocation+6]
	} else {
		offsetLocation := strings.Index(timeOffsetString, " -")
		timeOffsetString = timeOffsetString[offsetLocation+1 : offsetLocation+6]
	}

	timeOffset, err := strconv.Atoi(timeOffsetString[1:3])
	if err != nil {
		return Shift{}, err
	}
	timeOffset *= 60

	minutes, err := strconv.Atoi(timeOffsetString[3:])
	if err != nil {
		return Shift{}, err
	}
	timeOffset += minutes

	var middayInUTC int
	if timeOffsetString[0] == '+' {
		middayInUTC = (36*60 - timeOffset) % (24 * 60)
	} else {
		middayInUTC = (12*60 + timeOffset) % (24 * 60)
	}

	return GetShiftByTime(middayInUTC)
}

// GetShiftByTime returns the shift which is active at a specific time
func GetShiftByTime(time int) (Shift, error) {

	shifts, _, err := LoadShifts()
	if err != nil {
		return Shift{}, err
	}

	rightShift := Shift{}
	for i, shift := range shifts {
		if shift.Start < shift.End { // shift starts and ends during the same day
			if time >= shift.Start && time < shift.End {
				rightShift = shifts[i]
			}
		} else {
			if time >= shift.Start || time < shift.End {
				rightShift = shifts[i]
			}
		}
	}

	return rightShift, nil
}

// GetTimeUntilShift returns information about the next shift and time until it starts
func GetTimeUntilShift(shifts []Shift, shiftPos int) (int, error) {

	timeInUTC := time.Now().UTC()
	currentTime := timeInUTC.Hour()*60 + timeInUTC.Minute()

	nextShift := shifts[shiftPos%len(shifts)]

	timeUntilNextShift := nextShift.Start - currentTime
	if currentTime > nextShift.Start { // prevents timeUntilNextShift from being negative if the next shift starts during the next day
		timeUntilNextShift += 24 * 60
	}

	return timeUntilNextShift, nil
}

// GetCurrentAndOwnShift returns all shifts in a slice and the position of the current shift
func GetCurrentAndOwnShift() ([]Shift, int, int, error) {

	timeInUTC := time.Now().UTC()
	currentTime := timeInUTC.Hour()*60 + timeInUTC.Minute()

	shifts, ownShiftName, err := LoadShifts()
	if err != nil {
		return []Shift{}, 0, 0, err
	}

	currentShiftPos := -1
	ownShiftPos := -1
	for i, shift := range shifts {
		if shift.Start < shift.End { // shift starts and ends during the same day
			if currentTime >= shift.Start && currentTime < shift.End {
				currentShiftPos = i
			}
		} else {
			if currentTime >= shift.Start || currentTime < shift.End {
				currentShiftPos = i
			}
		}
		if shift.Name == ownShiftName {
			ownShiftPos = i
		}
	}

	return shifts, currentShiftPos, ownShiftPos, nil
}

// LoadShifts loads shifts out of the .pd.yml file
func LoadShifts() ([]Shift, string, error) {

	var err error
	config, err := loadConfig()
	if err != nil {
		return nil, "", err
	}

	finalShifts := make([]Shift, len(config.ShiftTimes))
	var min int
	for i, shift := range config.ShiftTimes {
		finalShifts[i] = Shift{}
		finalShifts[i].Start, err = strconv.Atoi(shift.Start[:2])
		if err != nil {
			return nil, "", err
		}
		finalShifts[i].Start *= 60
		min, err = strconv.Atoi(shift.Start[3:])
		if err != nil {
			return nil, "", err
		}

		finalShifts[i].End, err = strconv.Atoi(shift.End[:2])
		if err != nil {
			return nil, "", err
		}
		finalShifts[i].End *= 60
		min, err = strconv.Atoi(shift.End[3:])
		if err != nil {
			return nil, "", err
		}
		finalShifts[i].End += min
		finalShifts[i].Name = shift.Name
	}

	return finalShifts, config.OwnShift, nil
}
