package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	sauceUrl := fmt.Sprintf("https://api.%s.saucelabs.com/v1/storage/upload", os.Getenv("saucelabs_data_center"))
	appPath, appName, appDescription := os.Getenv("app_path"), os.Getenv("app_name"), os.Getenv("app_description")
	sauceLabsUsername, sauceLabsAccessKey := os.Getenv("saucelabs_username"), os.Getenv("saucelabs_access_key")

	// Prepare form-data body
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)

	fw, err := writer.CreateFormFile("payload", filepath.Base(appPath))
	if err != nil {
		log.Fatal(err)
	}
	fd, err := os.Open(appPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()
	_, err = io.Copy(fw, fd)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Loaded artifact from '%s'", appPath)
	fmt.Println()

	formField, err := writer.CreateFormField("name")
	if err != nil {
		log.Fatal(err)
	}
	_, err = formField.Write([]byte(appName))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Set name parameter with value '%s'", appName)
	fmt.Println()

	formField, err = writer.CreateFormField("description")
	if err != nil {
		log.Fatal(err)
	}
	_, err = formField.Write([]byte(appDescription))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Set description parameter with value '%s'", appDescription)
	fmt.Println()

	writer.Close()

	// Prepare HTTP Client with data-form body
	client := &http.Client{}
	req, err := http.NewRequest("POST", sauceUrl, form)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(sauceLabsUsername, sauceLabsAccessKey)
	fmt.Printf("POST request to be sent to SauceLabs with user '%s'", sauceLabsUsername)
	fmt.Println()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to post file to SauceLabs, error: %#v", err)
		os.Exit(2)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyText)
		cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", "SAUCELABS_RESPONSE", "--value", bodyString).CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
			os.Exit(3)
		}
	} else {
		fmt.Println("Failed to post file to SauceLabs, Status Code: ", resp.StatusCode)
		os.Exit(2)
	}

	os.Exit(0)
}
