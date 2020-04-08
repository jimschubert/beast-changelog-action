// Copyright 2020 Jim Schubert
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jimschubert/changelog"
	"github.com/jimschubert/changelog/model"
	act "github.com/sethvargo/go-githubactions"
	log "github.com/sirupsen/logrus"
)

func main() {
	githubToken := act.GetInput("GITHUB_TOKEN")
	if githubToken == "" {
		var ok bool
		// allow for local testing
		githubToken, ok = os.LookupEnv("GITHUB_TOKEN")
		if !ok {
			log.Fatal("Missing input 'GITHUB_TOKEN' in labeler action configuration.")
		}
	}

	fullRepo := act.GetInput("GITHUB_REPOSITORY")
	if !strings.Contains(fullRepo, "/") {
		log.WithFields(log.Fields{"GITHUB_REPOSITORY": fullRepo}).Fatal("Invalid GITHUB_REPOSITORY. Must be in the format: owner/repo")
	}

	configLocation := act.GetInput("CONFIG_LOCATION")
	if _, err := os.Stat(configLocation); os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"config_location": configLocation,
		}).Fatal("The CONFIG_LOCATION does not seem to exist. Did you checkout via actions/checkout first?")
	}

	from := act.GetInput("FROM")
	to := act.GetInput("TO")
	output := act.GetInput("OUTPUT")

	_ = os.Setenv("GITHUB_TOKEN", githubToken)
	_ = os.Setenv("LOG_LEVEL", "info")

	err := os.MkdirAll(filepath.Dir(output), os.ModeDir)
	if err != nil {
		log.WithFields(log.Fields{"output": output}).Fatal("Unable to create directory for changelog output")
	}

	outputFile, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.WithFields(log.Fields{"output": output}).Fatal("Unable to open changelog output for write")
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			log.Panic(err)
		}
	}()

	repoParts := strings.Split(fullRepo, "/")
	owner := repoParts[0]
	repo := repoParts[1]

	config := model.LoadOrNewConfig(&configLocation, owner, repo)
	changes := changelog.Changelog{
		Config: config,
		From:   from,
		To:     to,
	}

	err = changes.Generate(outputFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to generate changelog")
	}
}
