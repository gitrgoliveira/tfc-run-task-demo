package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type PrePlanPayload struct {
	PayloadVersion                  int    `json:"payload_version"`
	AccessToken                     string `json:"access_token"`
	Stage                           string `json:"stage"`
	IsSpeculative                   bool   `json:"is_speculative"`
	TaskResultID                    string `json:"task_result_id"`
	TaskResultEnforcementLevel      string `json:"task_result_enforcement_level"`
	TaskResultCallbackURL           string `json:"task_result_callback_url"`
	RunAppURL                       string `json:"run_app_url"`
	RunID                           string `json:"run_id"`
	RunMessage                      string `json:"run_message"`
	RunCreatedAt                    string `json:"run_created_at"`
	RunCreatedBy                    string `json:"run_created_by"`
	WorkspaceID                     string `json:"workspace_id"`
	WorkspaceName                   string `json:"workspace_name"`
	WorkspaceAppURL                 string `json:"workspace_app_url"`
	OrganizationName                string `json:"organization_name"`
	VCSRepoURL                      string `json:"vcs_repo_url"`
	VCSBranch                       string `json:"vcs_branch"`
	VCSPullRequestURL               string `json:"vcs_pull_request_url"`
	VCSCommitURL                    string `json:"vcs_commit_url"`
	ConfigurationVersionID          string `json:"configuration_version_id"`
	ConfigurationVersionDownloadURL string `json:"configuration_version_download_url"`
	WorkspaceWorkingDirectory       string `json:"workspace_working_directory"`
}

type Result struct {
	Data ResultData `json:"data"`
}
type ResultData struct {
	Type       string           `json:"type"`
	Attributes ResultAttributes `json:"attributes"`
}

type ResultAttributes struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	URL     string `json:"url,omitempty"`
}

// Queue to store the jobs (JSON payloads)
var jobQueue = make(chan PrePlanPayload, 100)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// log.Printf("Received payload: %s\n", string(body))

	// Parse the JSON payload
	var payload PrePlanPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	// Add the job to the queue
	jobQueue <- payload

	// Respond with an HTTP 200 OK status
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

func main() {
	// Start the job processor in a separate goroutine
	go processJobs()

	// Define the HTTP handler function
	http.HandleFunc("/", handleRequest)

	// Start the server on port 80
	log.Println("Server listening on port 80...")
	log.Fatal(http.ListenAndServe(":80", nil))
}

// downloadTarGz downloads a tar.gz file from the specified URL and extracts its contents.
//
// It takes a URL as a parameter and returns an error if any error occurs during the download or extraction process.
// The function creates a temporary file to store the downloaded tar.gz file, sends a GET request to download the file,
// copies the response body to the temporary file, opens the downloaded tar.gz file for reading, creates a reader for
// the gzip file, and extracts files from the tar archive. It returns nil if the process is successful.
func downloadConfigVersion(url, token, filename string) error {
	// Create a new HTTP request with the provided URL
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Add the token to the request header
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	// Create a temporary file to store the downloaded tar.gz file
	tempFile, err := os.Create(filename + ".tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	// Copy the response body to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return err
	}

	// Close the temporary file
	err = tempFile.Close()
	if err != nil {
		return err
	}
	log.Println("download OK")

	err = extractTarGz(tempFile.Name(), "./"+filename)
	if err != nil {
		return err
	} else {
		log.Println("File extracted OK")
	}

	return nil
}

func extractTarGz(tarGzFile, destination string) error {
	// Open the tar.gz file for reading
	file, err := os.Open(tarGzFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a gzip reader for the file
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// Create a tar reader for the gzip reader
	tarReader := tar.NewReader(gzipReader)

	// Extract files from the tar archive
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Extract the file to the specified destination
		target := filepath.Join(destination, header.Name)

		// Check the type of entry
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory if it doesn't exist
			err = os.MkdirAll(target, os.ModePerm)
			if err != nil {
				return err
			}

		case tar.TypeReg:
			// Create the file and copy contents
			err = os.MkdirAll(filepath.Dir(target), os.ModePerm)
			if err != nil {
				return err
			}

			file, err := os.Create(target)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported file type: %v in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

func sendPatchRequest(url string, payload []byte, authToken string) error {
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		return err
	}

	defer resp.Body.Close()
	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Convert the response body to a string
	respBodyStr := string(respBody)

	// Process the response as needed
	if resp.StatusCode != http.StatusOK {
		log.Println(respBodyStr)
		return err
	}

	return nil
}

func processJobs() {
	for payload := range jobQueue {
		// Lock the job queue
		// jobQueue.Lock()

		// Check if there are any jobs in the queue
		// if len(jobQueue.PrePlanPayloads) > 0 {
		// Retrieve the first job from the queue
		// payload := jobQueue.PrePlanPayloads[0]
		// jobQueue.PrePlanPayloads = jobQueue.PrePlanPayloads[1:]

		log.Printf("Processing job: %+v\n", payload.RunID)
		// Add your job processing logic here

		err := downloadConfigVersion(payload.ConfigurationVersionDownloadURL, payload.AccessToken, payload.RunID)
		if err != nil {
			log.Println(err.Error())
		}

		patternsFile := "patternsFile.txt"
		patterns, err := readRegexPatterns(patternsFile)
		if err != nil {
			log.Fatal(err)
		}

		matchCounts := runRegexOnFolder(payload.RunID, patterns)
		err = os.RemoveAll("./" + payload.RunID)
		if err != nil {
			log.Fatal(err)
		}

		if len(matchCounts) > 0 && err == nil {
			var message strings.Builder
			for pattern, count := range matchCounts {
				if count > 0 {
					message.WriteString(fmt.Sprintf("Pattern: %s, Matches: %d\n", pattern, count))
				}
			}

			log.Println(message.String())
			result := createFailedResult(message.String())
			jsonData, err := json.Marshal(result)
			if err != nil {
				log.Println(err.Error())
			}

			err = sendPatchRequest(payload.TaskResultCallbackURL, jsonData, payload.AccessToken)
			if err != nil {
				log.Println(err.Error())
			}

		} else {
			result := createPassedResult("Configured patterns not found")
			jsonData, err := json.Marshal(result)
			if err != nil {
				log.Println(err.Error())
			}

			err = sendPatchRequest(payload.TaskResultCallbackURL, jsonData, payload.AccessToken)
			if err != nil {
				log.Println(err.Error())
			}
		}
		// }

		// Unlock the job queue
		// jobQueue.Unlock()

		// Sleep for some time before checking for the next job
		time.Sleep(1 * time.Second)
	}
}

func createPassedResult(message string) Result {
	return Result{
		Data: ResultData{
			Type: "task-results",
			Attributes: ResultAttributes{
				Status:  "passed",
				Message: message,
				URL:     "",
			},
		},
	}
}
func createFailedResult(message string) Result {
	return Result{
		Data: ResultData{
			Type: "task-results",
			Attributes: ResultAttributes{
				Status:  "failed",
				Message: message,
				URL:     "",
			},
		},
	}
}

func runRegexOnFolder(folderPath string, regexPatterns []string) map[string]int {
	matchCounts := make(map[string]int)

	// Traverse each file in the folder
	err := filepath.Walk(folderPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Read the file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Match each regular expression against the file content
		for _, pattern := range regexPatterns {
			regex, err := regexp.Compile(pattern)
			if err != nil {
				log.Printf("Error compiling regex pattern: %s\n", pattern)
				continue
			}

			matches := regex.FindAll(content, -1)
			if matches != nil {
				matchCounts[pattern] += len(matches)
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return matchCounts
}

func readRegexPatterns(filePath string) ([]string, error) {
	patterns := []string{}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := scanner.Text()
		patterns = append(patterns, pattern)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}
