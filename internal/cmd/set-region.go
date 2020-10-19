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
	"fmt"
	"strings"

	"github.com/gonvenience/bunt"
	"github.com/homeport/pd/internal/pd"
	"github.com/spf13/cobra"
)

// setRegionCmd represents the get command
var setRegionCmd = &cobra.Command{
	Use:   "set-region",
	Args:  cobra.MaximumNArgs(1),
	Short: "Sets own region",
	Long:  `Sets the region you live in`,
	RunE: func(cmd *cobra.Command, args []string) error {

		shifts, _, err := pd.LoadShifts()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			ownShift, err := pd.GetProbablyOwnShift()
			if err != nil || ownShift.Name == "" {
				return err
			}
			err = pd.ChangeYAMLFile("own-shift", ownShift.Name)
			if err != nil {
				return err
			}
			bunt.Printf("\nYou've been added to SkyBlue{%s} because of your timezone.\n", ownShift.Name)
			bunt.Printf("If this is not the right region, please run the 'LightSlateGray{set-region}' command followed by your region name.\n\n")
		} else {
			pos := -1
			for i, shift := range shifts {
				if shift.Name[5:] == args[0] {
					pos = i
				}
			}
			if pos == -1 {
				shiftNames := []string{}
				for i, shift := range shifts {
					shiftNames = append(shiftNames, shift.Name[5:])
					if i != len(shifts)-1 {
						shiftNames[i] += " /"
					}
				}
				bunt.Printf("\nYour input was invalid. Please run the 'LightSlateGray{set-region}' command followed by one of these:  %s\n\n", strings.Trim(fmt.Sprint(shiftNames), "[]"))
				return nil
			}
			err := pd.ChangeYAMLFile("own-shift", shifts[pos].Name)
			if err != nil {
				return err
			}
			bunt.Printf("\nYou've been added to SkyBlue{%s}\n\n", shifts[pos].Name)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setRegionCmd)
}
