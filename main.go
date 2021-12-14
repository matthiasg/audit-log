package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var currentDirectory, err = os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Hello, geeksforgeeks")

	iterate(currentDirectory)
}

type Reference struct {
	Id      string `json:"id"`
	Version string `json:"v"`
	Text    string `json:"text"`
}

type DateTimeStamp struct {
	At string `json:"at"`
}

type Document struct {
	Id       string                 `json:"id"`
	Version  string                 `json:"v"`
	Form     Reference              `json:"f"`
	Created  DateTimeStamp          `json:"created"`
	Modified DateTimeStamp          `json:"modified"`
	Data     map[string]interface{} `json:"data"`
	Previous map[string]interface{} `json:"previous"`
}

type ChangeEvent struct {
	Path     string
	Type     string `json:"type"`
	Document *Document
}

type CurrentUserData struct {
	CurrentUser Reference `json:"currentUser"`
}

const NUMBER_OF_MONTHS_TO_REVIEW = 1
const VITEGRA_INSTANCE_LOGIN = "f0380b58-0987-43d9-89f7-bcffa6fff82c"
const USER_SETTINGS = "4a152c0c-df7c-4ad5-8c62-495a01216308"
const LOGIN_FORM_ID = "f0533e05-617b-47ac-a5ca-32c3a36643c2"
const USER_FORM_ID = "0f3894e8-b5ad-40b4-89a1-df30e8476a15"
const REVIEW_FORM_ID = "4ce26984-aad7-4847-9de1-9ad8d4f7fe9d"
const PATIENT_FORM_ID = "0b2d5b3a-abed-44f2-959c-591f6af5161c"
const IMAGE_FORM_ID = "032d1643-a3cb-4a8e-8cd5-237b5f32e211"
const VIDEO_FORM_ID = "ee450438-2918-47d4-8892-b6f5df3d508a"

const PROCEDURE_FORM_ID = "e90b9ab4-085a-4973-8ef0-ae683599c92c"
const CURRENT_RECORDING_INFO = "vitegra-recording-info"

func iterate(path string) {
	now := time.Now().UTC()
	startDateTime := now.AddDate(0, -NUMBER_OF_MONTHS_TO_REVIEW, 0)

	minimumDateInIso := startDateTime.Format(time.RFC3339) // "2006-01-02T15:04:05-0700")

	var importantEvents []ChangeEvent

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}

		if !info.IsDir() && filepath.Ext(path) == ".json" {
			content, err := ioutil.ReadFile(path)

			if err != nil {
				log.Fatal("Error reading file: ", err)
			}

			var payload map[string]interface{}

			err = json.Unmarshal(content, &payload)

			if err != nil {
				log.Fatal("Error during JSON decode: ", err)
			}

			// fmt.Println(payload)

			changeEvent := ChangeEvent{Path: path}

			if payload["type"] != nil {
				err = json.Unmarshal(content, &changeEvent)

				if err != nil {
					log.Fatal("Error decoding typed change event: ", path, err)
				}

				// fmt.Println("Typed ChangeEvent", changeEvent)
			} else {
				var document Document
				err = json.Unmarshal(content, &document)

				if err != nil {
					log.Fatal("Error decoding document event: ", path, err)
				}

				changeEvent.Type = "updated"
				changeEvent.Document = &document
			}

			isNewerThanMinumum := changeEvent.Document.Modified.At >= minimumDateInIso
			// fmt.Println(isNewerThanMinumum, changeEvent.Document.Modified.At, minimumDateInIso)

			var isImportantEvent = false

			// Could be optimized of course, but we want to make it very clear for novices :)
			if changeEvent.Document.Id == VITEGRA_INSTANCE_LOGIN {
				isImportantEvent = true
			}

			if changeEvent.Document.Id == USER_SETTINGS {
				isImportantEvent = true
			}

			if changeEvent.Document.Form.Id == LOGIN_FORM_ID {
				isImportantEvent = true
			}

			if changeEvent.Document.Form.Id == USER_FORM_ID {
				isImportantEvent = true
			}

			if changeEvent.Document.Form.Id == REVIEW_FORM_ID {
				isImportantEvent = true
			}

			if changeEvent.Document.Form.Id == PROCEDURE_FORM_ID {
				isImportantEvent = true
			}

			if changeEvent.Document.Form.Id == VIDEO_FORM_ID {
				isImportantEvent = true
			}

			if changeEvent.Document.Form.Id == IMAGE_FORM_ID {
				isImportantEvent = true
			}

			if changeEvent.Document.Form.Id == PATIENT_FORM_ID {
				isImportantEvent = true
			}

			if isNewerThanMinumum && isImportantEvent {
				importantEvents = append(importantEvents, changeEvent)
			}
		}

		// fmt.Printf("File Name: %s\n", info.Name())
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, changeEvent := range importantEvents {
		document := changeEvent.Document

		if document != nil && changeEvent.Type == "updated" {
			// fmt.Println(document.Id, document.Form.Id, document.Form.Text)
			if document.Id == VITEGRA_INSTANCE_LOGIN {
				var loggedInUserRef = document.Data["currentUser"].(map[string]interface{})
				var userId = loggedInUserRef["id"]

				if userId != "" {
					fmt.Printf("%s;LOGIN;%s;%s;\n", document.Modified.At, loggedInUserRef["id"], loggedInUserRef["text"])
				} else {
					fmt.Printf("%s;LOGOUT;\n", document.Modified.At)
				}
			} else if document.Form.Id == LOGIN_FORM_ID {
				var loginAttemptByUser = document.Data["user"].(map[string]interface{})
				var isLoginSuccessful = document.Data["success"].(bool)

				var isLockoutOverride = false

				if document.Version != "initial" {
					var patch = document.Previous["patch"].(map[string]interface{})

					if patch != nil && patch["lockoutOverride"] != nil {
						isLockoutOverride = true
					}
				}

				if isLockoutOverride {
					fmt.Printf("%s;LOGIN_UNLOCK;%s;%s;\n", document.Modified.At, loginAttemptByUser["id"], loginAttemptByUser["text"])
				} else {
					if !isLoginSuccessful {
						var loginSuccessful = boolToString(isLoginSuccessful, "SUCCESS", "FAILED")
						fmt.Printf("%s;LOGIN_ATTEMPT;%s;%s;%s\n", document.Modified.At, loginAttemptByUser["id"], loginAttemptByUser["text"], loginSuccessful)
					}
				}
			} else if document.Form.Id == USER_FORM_ID {
				if document.Previous["patch"] == nil {
					fmt.Printf("%s;NEW_USER;%s;%s;%s\n", document.Modified.At, document.Data["name"], document.Id, changeEvent.Path)
				} else {
					var patch = document.Previous["patch"].(map[string]interface{})
					var isPasswordChange = false
					var isLockedChange = false

					if patch["password"] != nil {
						isPasswordChange = true
					}

					if patch["locked"] != nil {
						isLockedChange = true
					}

					var user = document
					var userName = document.Data["name"]

					if isPasswordChange {
						fmt.Printf("%s;PASSWORD_CHANGE;%s;%s;\n", document.Modified.At, user.Id, userName)
					}

					if isLockedChange {
						var isNowLocked = (patch["locked"].([]interface{})[1].(bool))
						fmt.Printf("%s;USER_LOCK_CHANGE;%s;%s;%s\n", document.Modified.At, user.Id, userName, boolToString(isNowLocked, "LOCKED", "UNLOCKED"))
					}
				}
			} else if document.Form.Id == PATIENT_FORM_ID {
				fmt.Printf("%s;PATIENT;%s;%s;%s\n", document.Modified.At, document.Data["name"], document.Id, changeEvent.Path)
			} else if document.Form.Id == REVIEW_FORM_ID {
				var user = document.Data["user"].(map[string]interface{})
				fmt.Printf("%s;PROCEDURE_REVIEW;%s;%s;\n", document.Modified.At, user["id"], user["text"])
			} else if document.Form.Id == PROCEDURE_FORM_ID {
				fmt.Printf("%s;PROCEDURE;%s;%s;%s\n", document.Modified.At, document.Data["name"], document.Id, changeEvent.Path)
			} else if document.Form.Id == VIDEO_FORM_ID {
				fmt.Printf("%s;VIDEO;%s;%s;%s\n", document.Modified.At, document.Data["name"], document.Id, changeEvent.Path)
			} else if document.Form.Id == IMAGE_FORM_ID {
				fmt.Printf("%s;IMAGE;%s;%s;%s\n", document.Modified.At, document.Data["name"], document.Id, changeEvent.Path)
			} else if document.Id == USER_SETTINGS {
				fmt.Printf("%s;USER_SETTINGS_CHANGE;%s;%s;%s\n", document.Modified.At, document.Data["name"], document.Id, changeEvent.Path)
				// fmt.Println(changeEvent.Path, changeEvent.Type, changeEvent.Document.Form.Text)
			} else {
				fmt.Println(changeEvent.Path, changeEvent.Type, changeEvent.Document.Form.Text)
			}
		}
	}

	// fmt.Printf("Number of important events #%d\n", len(importantEvents))
}

func boolToString(booleanValue bool, trueText string, falseText string) string {
	if booleanValue {
		return trueText
	} else {
		return falseText
	}
}
