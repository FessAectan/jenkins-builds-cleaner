package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/buger/jsonparser"
)

/* global variables (get them from ENV)
username - username in Jenkins
token - his/her token
jenkinsHostname - like jenkins.example.com
*/
var username = os.Getenv("JENKINS_USERNAME")
var token = os.Getenv("JENKINS_TOKEN")
var jenkinsHostname = os.Getenv("JENKINS_HOSTNAME")

// ParseJSON returns us a slice with job's names
func ParseJSON(responseBody []uint8, jsonHead, jsonTarget string) []string {
	var jobs []string
	jsonparser.ArrayEach(responseBody, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		v, _, _, err := jsonparser.Get(value, jsonTarget)
		if err != nil {
			log.Fatalln(err)
			return
		}
		jobs = append(jobs, string(v[:]))
	}, jsonHead)

	return jobs
}

// MakeRequest returns us a resp.Body from http.Get
// or just send POST to Jenkins for deleting old builds
func MakeRequest(url, reqType string) []uint8 {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	URLstart := "https://" + username + ":" + token + "@" + jenkinsHostname
	URLend := "/api/json?pretty=true"
	var result []uint8

	switch reqType {
	case "get":
		resultURL := URLstart + url + URLend
		resp, err := client.Get(resultURL)
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		result = body

	case "post":
		resultURL := URLstart + url
		fmt.Println("I am deleting the build: " + url + "'")
		resp, err := client.Post(resultURL, "", nil)
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()
	}

	return result
}

func main() {
	/*
	  msNameAndBranches - an empty map for jobs (microservices) names and their branches
	  msDoesntHaveBranches - a slice for jobs, that do not have branches
	  msHasBranches - an empty map for jobs, that have branches
	*/
	msNameAndBranches := make(map[string][]string)
	var msDoesntHaveBranches []string
	msHasBranches := make(map[string][]string)

	// get jobs (microservice's) names,
	// e.g. curl  https://USER:TOKEN@jenkins.example.com/api/json?pretty=true|jq
	microservices := ParseJSON(MakeRequest("", "get"), "jobs", "name")

	// fill the map msNameAndBranches
	for _, msName := range microservices {
		msNameAndBranches[string(msName)] = ParseJSON(MakeRequest("/job/"+string(msName), "get"), "jobs", "name")
	}

	// fill the slice msDoesntHaveBranches and map msHasBranches
	for msName, msBranches := range msNameAndBranches {
		if len(msBranches) == 0 {
			msDoesntHaveBranches = append(msDoesntHaveBranches, msName)
			continue
		}
		msHasBranches[string(msName)] = msBranches
	}

	// delete old build in jobs, which have branches
	for msName, msBranches := range msHasBranches {
		for _, branch := range msBranches {
			msNameS := string(msName)
			branchS := string(branch)
			msBuilds := ParseJSON(MakeRequest("/job/"+msNameS+"/job/"+branchS, "get"), "builds", "number")

			if len(msBuilds) >= 10 {
				for _, buildNumber := range msBuilds[10:] {
					MakeRequest("/job/"+msNameS+"/job/"+branchS+"/"+buildNumber+"/doDelete", "post")
				}
			}
		}
	}

	// delete builds in jobs, which don't have branches
	for _, msName := range msDoesntHaveBranches {
		msBuilds := ParseJSON(MakeRequest("/job/"+msName, "get"), "builds", "number")
		if len(msBuilds) >= 10 {
			for _, buildNumber := range msBuilds[10:] {
				MakeRequest("/job/"+msName+"/"+buildNumber+"/doDelete", "post")
			}
		}
	}
}
