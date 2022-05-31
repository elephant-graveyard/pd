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
	"os"
	"sort"
	"strings"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/gonvenience/bunt"
	"github.com/gonvenience/neat"
	"github.com/gonvenience/wrap"
	"github.com/homeport/pd/internal/pd"
	"github.com/spf13/cobra"
)

// onCallCmd represents the onCall command
var onCallCmd = &cobra.Command{
	Use:   "on-call",
	Short: "List on-calls for user",
	Long:  `Check PagerDuty for all on-calls of the current user`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := pd.CreatePagerDutyClient()
		if err != nil {
			return err
		}

		user, err := client.GetCurrentUserWithContext(cmd.Context(), pagerduty.GetCurrentUserOptions{})
		if err != nil {
			return wrap.Error(err, "it seems like the authtoken is not set correctly or outdated. Please update the authtoken in the .pd.yml file. If you don't know how to create your authtoken, this might help:\n https://support.pagerduty.com/docs/generating-api-keys#generating-a-personal-rest-api-key\n")
		}

		oncalls, err := pd.GetPagerDutyOnCalls(cmd.Context(), client, user)
		if err != nil {
			neat.Box(
				os.Stderr,
				"Unexpected error occurred",
				strings.NewReader(err.Error()),
				neat.HeadlineColor(bunt.FireBrick),
				neat.ContentColor(bunt.Coral),
			)
		}

		switch len(oncalls) {
		case 0:
			bunt.Printf("\nYou are fine, there seem to be *no* on-call listed for your user.\nHave a nice day.\n\n")

		default:
			bunt.Printf("\nIt turns out, you *are* on-call.\n\n")
			for timeRange, escalationPolicies := range oncalls {
				var table = [][]string{{bunt.Sprint("*EscalationPolicy*"), bunt.Sprint("*Link*")}}
				for _, policy := range escalationPolicies {
					table = append(table, []string{
						policy.Summary,
						policy.HTMLURL,
					})
				}

				sort.Slice(table, func(i, j int) bool {
					return strings.Compare(table[i][0], table[j][0]) < 0
				})

				out, err := neat.Table(table, neat.VertialBarSeparator())
				if err != nil {
					return err
				}

				neat.Box(
					os.Stdout,
					bunt.Sprintf("*on-call* from *%v* to *%v*", timeRange.Start.Local(), timeRange.End.Local()),
					strings.NewReader(out),
					neat.HeadlineColor(bunt.LightSteelBlue),
					neat.NoLineWrap(),
				)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(onCallCmd)
}
