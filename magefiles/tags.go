//go:build mage

package main

import (
	"fmt"
	semver "github.com/blang/semver/v4"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"os"
	"strings"
)

type Tag mg.Namespace

var git = sh.OutCmd("git")
var gh = sh.OutCmd("gh")

// Nginx returns the ingress-nginx current version
func (Tag) Nginx() {
	tag, err := getIngressNGINXVersion()
	CheckIfError(err, "")
	fmt.Printf("%v", tag)
}

func getIngressNGINXVersion() (string, error) {
	dat, err := os.ReadFile("TAG")
	CheckIfError(err, "Could not read TAG file")
	datString := string(dat)
	//remove newline
	datString = strings.Replace(datString, "\n", "", -1)
	return datString, nil
}

func checkSemVer(currentVersion, newVersion string) bool {
	cVersion, err := semver.Make(currentVersion[1:])
	if err != nil {
		ErrorF("Error Current Tag %v Making Semver : %v", currentVersion[1:], err)
		return false
	}
	nVersion, err := semver.Make(newVersion[1:])
	if err != nil {
		ErrorF("%v Error Making Semver %v \n", newVersion, err)
		return false
	}

	err = nVersion.Validate()
	if err != nil {
		ErrorF("%v not a valid Semver %v \n", newVersion, err)
		return false
	}

	//The result will be
	//0 if newVersion == currentVersion
	//-1 if newVersion < currentVersion
	//+1 if newVersion > currentVersion.
	comp := nVersion.Compare(cVersion)
	if comp <= 0 {
		Warning("SemVer:%v is not an update\n", newVersion)
		return false
	}
	return true
}

func bump(currentTag, newTag string) {
	//check if semver is valid
	if !checkSemVer(currentTag, newTag) {
		ErrorF("ERROR: Semver is not valid %v \n", newTag)
		os.Exit(1)
	}

	Debug("Updating Tag %v to %v \n", currentTag, newTag)
	err := os.WriteFile("TAG", []byte(newTag), 0666)
	CheckIfError(err, "Error Writing New Tag File")
}

// BumpNginx will update the nginx TAG
func (Tag) BumpNginx(newTag string) {
	currentTag, err := getIngressNGINXVersion()
	CheckIfError(err, "Getting Ingress-nginx Version")
	bump(currentTag, newTag)
}

// Git Returns the latest git tag
func (Tag) Git() {
	tag, err := getGitTag()
	CheckIfError(err, "Retrieving Git Tag")
	Info("Git tag: %v", tag)
}

func getGitTag() (string, error) {
	return git("describe", "--tags", "--match", "controller-v*", "--abbrev=0")
}

// ControllerTag Creates a new Git Tag for the ingress controller
func (Tag) ControllerTag(version string) {
	tag, err := git("tag", "-a", fmt.Sprintf("controller-%s", version), fmt.Sprintf("-m \"Automated Controller release %v", version))
	CheckIfError(err, "Creating git tag")
	Debug("Git :qTag: %s", tag)
}

func (Tag) AllControllerTags() {
	tags := getAllControllerTags()
	for i, s := range tags {
		Info("#%v Version %v", i, s)
	}
}

func getAllControllerTags() []string {
	allControllerTags, err := git("tag", "-l", "--sort=-v:refname", "controller-v*")
	CheckIfError(err, "Retrieving git tags")
	if !sh.CmdRan(err) {
		Warning("Issue Running Command")
	}
	if allControllerTags == "" {
		Warning("All Controller Tags is empty")
	}
	Debug("Controller Tags: %v", allControllerTags)

	temp := strings.Split(allControllerTags, "\n")
	Debug("There are %v controller tags", len(temp))
	return temp
}
