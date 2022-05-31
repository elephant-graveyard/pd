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
	"html/template"
	"os"
	"strings"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/gonvenience/bunt"
	"github.com/homeport/pd/internal/pd"
	"github.com/spf13/cobra"
)

type relevantIncidentsReturn struct {
	Title             string
	RelevantIncidents []pagerduty.Incident
	OtherIncidents    []pagerduty.Incident
}

var shiftReportCmdSettings struct {
	id           string
	templateName string
	date         string
}

// onCallCmd represents the onCall command
var shiftReportCmd = &cobra.Command{
	Use:   "shift-report",
	Args:  cobra.MaximumNArgs(1),
	Short: "Creates shift report",
	Long:  `Creates a shift report based on the provided template`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if shiftReportCmdSettings.id == "" || shiftReportCmdSettings.templateName == "" || shiftReportCmdSettings.date == "" {
			bunt.Println("\nPlease use all flags!\n")
			return nil
		}

		incidents, username, err := getRelevantIncidents(
			cmd.Context(),
			shiftReportCmdSettings.id,
			shiftReportCmdSettings.date+"T00:00:01Z",
			shiftReportCmdSettings.date+"T23:59:59Z",
		)

		if err != nil {
			return err
		}

		data, err := pd.GetTemplate(shiftReportCmdSettings.templateName)
		if err != nil {
			return err
		}

		shifts, _, shiftPos, err := pd.GetCurrentAndOwnShift()
		if err != nil {
			return err
		}

		funcMap := map[string]interface{}{
			"makeSlice":                    makeSlice,
			"getCategoryMatchingIncidents": getCategoryMatchingIncidents,
		}
		temp := template.New("template").Funcs(template.FuncMap(funcMap))

		temp, _ = temp.Parse(string(data))

		input := struct {
			Username        string
			Date            string
			StartOfOwnShift string
			EndOfOwnShift   string
			Incidents       []pagerduty.Incident
		}{
			Username:        username,
			Date:            shiftReportCmdSettings.date,
			StartOfOwnShift: convertTimeIntToString(int(shifts[shiftPos].Start)),
			EndOfOwnShift:   convertTimeIntToString(int(shifts[shiftPos].End)),
			Incidents:       incidents,
		}

		bunt.Println()
		temp.Execute(os.Stdout, input)

		return nil
	},
}

func convertTimeIntToString(i int) string {
	return bunt.Sprintf("%02d:%02d", i/60, i%60)
}

func makeSlice(args ...interface{}) []interface{} {
	return args
}

func getCategoryMatchingIncidents(name string, incidents []pagerduty.Incident) relevantIncidentsReturn {
	posSeparator := strings.Index(name, "--")
	if posSeparator == -1 {
		return relevantIncidentsReturn{Title: name, RelevantIncidents: incidents, OtherIncidents: []pagerduty.Incident{}}
	}
	searchFor := name[posSeparator+2:]
	name = name[:posSeparator]
	relevantIncidents := []pagerduty.Incident{}
	otherIncidents := []pagerduty.Incident{}
	for _, incident := range incidents {
		if strings.Contains(incident.Title, searchFor) {
			relevantIncidents = append(relevantIncidents, incident)
		} else {
			otherIncidents = append(otherIncidents, incident)
		}
	}

	return relevantIncidentsReturn{Title: name, RelevantIncidents: relevantIncidents, OtherIncidents: otherIncidents}
}

func init() {
	rootCmd.AddCommand(shiftReportCmd)

	shiftReportCmd.Flags().StringVar(&shiftReportCmdSettings.id, "id", "", "use custom ID")
	shiftReportCmd.Flags().StringVar(&shiftReportCmdSettings.templateName, "template", "", "set path of shift-report.template file")
	shiftReportCmd.Flags().StringVar(&shiftReportCmdSettings.date, "date", "", "set date of shift report")
}
