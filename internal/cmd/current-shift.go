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

package cmd

import (
	"github.com/gonvenience/bunt"
	"github.com/homeport/pd/internal/pd"
	"github.com/spf13/cobra"
)

// currentShiftCmd represents the get command
var currentShiftCmd = &cobra.Command{
	Use:   "current-shift",
	Args:  cobra.ExactArgs(0),
	Short: "Display current shift",
	Long:  `Displays the region of the current shift`,
	RunE: func(cmd *cobra.Command, args []string) error {

		shifts, shiftPos, err := pd.GetCurrentShift()
		if err != nil || shiftPos == -1 {
			return err
		}
		bunt.Printf("\nAt the moment, SkyBlue{%s} is in charge.\n", shifts[shiftPos].Name)

		nextShift, timeUntilNextShift, err := pd.GetNextShift(shifts, shiftPos)
		if err != nil {
			return err
		}
		bunt.Printf("The next shift will be SkyBlue{%s} in %d:%02d hours\n\n", nextShift.Name, timeUntilNextShift/60, timeUntilNextShift%60)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(currentShiftCmd)
}
