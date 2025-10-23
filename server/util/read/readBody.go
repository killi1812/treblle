package read

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// readBody safely reads the request body and replaces it so it can be read again.
// This is crucial for proxying requests.
func ReadBody(req *http.Response) ([]byte, error) {
	if req.Body == nil {
		// Handle nil body (e.g., for GET requests), return empty bytes
		return []byte{}, nil
	}

	// Read the entire body
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// Restore the body so it can be read again
	// This is crucial for the ReverseProxy to work
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// Update the content length for the proxy, otherwise the proxy
	// might send a 0-length body.
	req.ContentLength = int64(len(bodyBytes))

	return bodyBytes, nil
}
