package internal

import (
	"fmt"
	"io"
	"os/exec"
	"time"
)

const (
	timeout = time.Second * 10
)

// ExecuteTests run groovy testing scripts.
func ExecuteTests(paths []string, stdout, stderr io.Writer) bool {
	status := true
	for _, path := range paths {
		cmd := exec.Command("groovy", "-cp", "src", path)
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		fmt.Println("Running tests: ", path)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error tests: %s\n%v\n", path, err)
			status = false
		}
	}
	return status
}

// UploadScripts authenticates over oauth2, then upload iflow scripts.
func UploadScripts(client IBTPClient, readFile func(string) ([]byte, error), iflows []Iflow) error {
	if err := client.RequestToken(); err != nil {
		return fmt.Errorf("RequestToken: %w", err)
	}
	if err := client.FetchCSRFToken(); err != nil {
		return fmt.Errorf("FetchCSRFToken: %w", err)
	}

	var uploadErr error
	for _, iflow := range iflows {
		for _, script := range iflow.Scripts {
			data, err := readFile(script.Path)
			if err != nil {
				fmt.Printf("FAILURE reading %s, %v\n", script.Path, err)
				uploadErr = fmt.Errorf("some reading/uploading scripts failed")
				continue
			}
			if err := client.UpdateIflowResource(data, iflow, script); err != nil {
				fmt.Printf("FAILURE uploading %s, %v\n", script.ID, err)
				uploadErr = fmt.Errorf("some reading/uploading scripts failed")
				continue
			}
			fmt.Printf("SUCCESS uploading %s\n", script.ID)
		}
	}
	return uploadErr
}
