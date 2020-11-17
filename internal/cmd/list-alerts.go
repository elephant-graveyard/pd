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
	"strings"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/gonvenience/bunt"
	"github.com/homeport/pd/internal/pd"
	"github.com/spf13/cobra"
)

var listAlertsCmdSettings struct {
	id   string
	from string
	to   string
}

// currentShiftCmd represents the get command
var listAlertsCmd = &cobra.Command{
	Use:   "list-alerts",
	Args:  cobra.ExactArgs(0),
	Short: "Lists all alerts",
	Long:  `Lists all alerts in a specified time period`,
	RunE: func(cmd *cobra.Command, args []string) error {

		client, err := pd.CreatePagerDutyClient()
		if err != nil {
			return err
		}

		incidents, _, err := getRelevantIncidents(listAlertsCmdSettings.id, listAlertsCmdSettings.from, listAlertsCmdSettings.to)
		if err != nil {
			return err
		}

		for i, incident := range incidents {

			bunt.Printf("\n%d. *%s*\n", i+1, incident.Title)

			if incident.Description != incident.Title {
				bunt.Printf("   *Description:* \n")
				for _, line := range strings.Split(incident.Description, "\n") {
					bunt.Printf("      %s\n", line)
				}
			}

			bunt.Printf("   *Link:* CornflowerBlue{~%s~}\n", incident.HTMLURL)

			start := mustParsePagerDutyRFC3339ishTime(incident.CreatedAt)

			end := mustParsePagerDutyRFC3339ishTime(incident.LastStatusChangeAt)

			bunt.Printf("   *Time:* %s - %s (%s)\n",
				start.Local().Format("2006-01-02 15:04:05"),
				end.Local().Format("2006-01-02 15:04:05"),
				end.Sub(start),
			)

			notes, err := client.ListIncidentNotes(incident.Id)
			if err != nil {
				return err
			}
			if len(notes) > 0 {
				bunt.Printf("   *Notes:*\n")
				for j := len(notes) - 1; j >= 0; j-- {
					bunt.Printf("      %d. ", len(notes)-j)
					for _, line := range strings.Split(notes[j].Content, "\n") {
						bunt.Printf("%s\n         ", line)
					}
					bunt.Printf("(by _%s_ at _%s_)\n", lookUpNameByUserID(client, notes[j].User.ID), formatNoteTime(notes[j].CreatedAt))
				}
			}
		}

		bunt.Println()

		return nil
	},
}

func lookUpNameByUserID(client *pagerduty.Client, id string) string {
	user, err := client.GetUser(id, pagerduty.GetUserOptions{})
	if err != nil {
		return "unknown"
	}
	return user.Name
}

func formatNoteTime(input string) string {
	time, err := time.Parse("2006-01-02T15:04:05-07:00", input)
	if err != nil {
		return input
	}

	return time.Local().Format("2006-01-02 15:04:05")
}

func mustParsePagerDutyRFC3339ishTime(input string) time.Time {
	time, err := time.Parse(time.RFC3339, input)
	if err != nil {
		panic(err)
	}
	return time
}

func init() {
	rootCmd.AddCommand(listAlertsCmd)

	listAlertsCmd.Flags().StringVar(&listAlertsCmdSettings.id, "id", "", "use custom ID")
	listAlertsCmd.Flags().StringVar(&listAlertsCmdSettings.from, "from", "", "set startpoint of custom time period")
	listAlertsCmd.Flags().StringVar(&listAlertsCmdSettings.to, "to", "", "set endpoint of custom time period")
}
