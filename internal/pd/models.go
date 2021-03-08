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
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Tip: Check https://yaml.to-go.online/ or https://mholt.github.io/json-to-go/
// for an easy way to translate YAML or JSON files into Go struct code.

// Config describes the pd tool configuration structure
type Config struct {
	Authtoken  string `yaml:"authtoken"`
	OwnShift   string `yaml:"own-shift"`
	ShiftTimes []struct {
		Name  string `yaml:"name"`
		Start string `yaml:"start"`
		End   string `yaml:"end"`
	} `yaml:"shift-times"`

	Templates map[string]string `yaml:"templates"`
}

// ChangeYAMLFile changes a specific value in the .pd.yml file
func ChangeYAMLFile(name string, newValue string) error {

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(filepath.Join(home, ".pd.yml"))
	if err != nil {
		return err
	}

	config := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	config[name] = newValue

	d, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(home, ".pd.yml"), d, 0644)
	if err != nil {
		return err
	}

	return nil
}
