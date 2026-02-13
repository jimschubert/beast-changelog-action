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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jimschubert/changelog"
	"github.com/jimschubert/changelog/model"
	act "github.com/sethvargo/go-githubactions"
	log "github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		level = "info"
		_ = os.Setenv("LOG_LEVEL", level)
	}

	logLevel, _ := log.ParseLevel(level)
	log.SetFormatter(&customFormatter{
		EnableTimestamp: logLevel == log.TraceLevel,
	})

	log.SetLevel(logLevel)

	log.Infof("beast-changelog-action %s (%s)", version, commit)
	log.Infof("https://github.com/jimschubert/beast-changelog-action")
	fmt.Println()

	githubToken := act.GetInput("GITHUB_TOKEN")
	if githubToken == "" {
		var ok bool
		// allow for local testing
		githubToken, ok = os.LookupEnv("GITHUB_TOKEN")
		if !ok {
			log.Fatal("Missing input 'GITHUB_TOKEN' in action configuration")
		}
	}

	fullRepo := act.GetInput("GITHUB_REPOSITORY")
	owner, repo, found := strings.Cut(fullRepo, "/")
	if !found {
		log.WithFields(log.Fields{"GITHUB_REPOSITORY": fullRepo}).Fatal("Invalid GITHUB_REPOSITORY format. Expected: owner/repo")
	}
	log.Infof("Target repository: %s", fullRepo)

	configLocation := act.GetInput("CONFIG_LOCATION")
	if _, err := os.Stat(configLocation); os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"config_location": configLocation,
		}).Fatal("Config file not found. Did you checkout via actions/checkout first?")
	}
	log.Infof("Using config: %s", configLocation)

	from := act.GetInput("FROM")
	to := act.GetInput("TO")
	output := act.GetInput("OUTPUT")

	log.Infof("Generating changelog: %s → %s", from, to)
	log.Infof("Output file: %s", output)

	_ = os.Setenv("GITHUB_TOKEN", githubToken)

	err := os.MkdirAll(filepath.Dir(output), 0755)
	if err != nil {
		log.WithFields(log.Fields{"output": output}).Fatal("Unable to create output directory")
	}

	outputFile, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.WithFields(log.Fields{"output": output}).Fatal("Unable to open output file for writing")
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			log.Panic(err)
		}
	}()

	log.Info("Loading configuration and generating changelog...")
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

	log.Info("✅ Changelog generated successfully!")
}
