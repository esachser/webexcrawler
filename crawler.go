package webexcrawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Room represents a Webex room.
type Room struct {
	ID           string `json:"id,omitempty"`
	Title        string `json:"title,omitempty"`
	Type         string `json:"type,omitempty"`
	IsLocked     bool   `json:"isLocked,omitempty"`
	LastActivity string `json:"lastActivity,omitempty"`
	CreatorID    string `json:"creatorId,omitempty"`
	Created      string `json:"created,omitempty"`
	OwnerID      string `json:"ownerId,omitempty"`
	IsPublic     bool   `json:"isPublic,omitempty"`
	IsReadOnly   bool   `json:"isReadOnly,omitempty"`
}

// RoomsResponse represents the response from the Webex API call to get rooms.
type RoomsResponse struct {
	Rooms []Room `json:"items,omitempty"`
}

type ErrorRetryAfter struct {
	RetryAfter int
}

func (e *ErrorRetryAfter) Error() string {
	return fmt.Sprintf("retry after %d seconds", e.RetryAfter)
}

type Crawler struct {
	ApiKey  string
	client  http.Client
	baseUrl string
}

func NewCrawler() *Crawler {
	apikey := os.Getenv("WEBEX_APIKEY")

	return &Crawler{
		ApiKey:  apikey,
		client:  *http.DefaultClient,
		baseUrl: "https://webexapis.com",
	}
}

func (c *Crawler) GetRooms(maxRooms int) ([]Room, error) {
	// Create a new request
	apiurl := fmt.Sprintf("%s/v1/rooms?max=%d&sortBy=lastactivity", c.baseUrl, maxRooms)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", apiurl, nil)
	if err != nil {
		return nil, err
	}

	// Set the authorization header
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			// Handle rate limiting
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				retryAfterInt, err := strconv.Atoi(retryAfter)
				if err == nil {
					// return nil, &ErrorRetryAfter{RetryAfter: retryAfterInt}
					log.Printf("Rate limit exceeded, retrying after %d seconds\n", retryAfterInt)
					time.Sleep(time.Duration(retryAfterInt) * time.Second)
					return c.GetRooms(maxRooms)
				}
			}
			return nil, fmt.Errorf("rate limit exceeded")
		}
		return nil, fmt.Errorf("failed to get rooms: %s", resp.Status)
	}

	// Decode the response
	var response RoomsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Rooms, nil
}

// GetMessages retrieves messages from a specific room.
func (c *Crawler) GetMessages(roomID string, maxMessages int, beforeMessageId string, before string) ([]Message, error) {
	// Create a new request
	apiurl := fmt.Sprintf("%s/v1/messages?roomId=%s&max=%d", c.baseUrl, roomID, maxMessages)
	if beforeMessageId != "" {
		apiurl += "&beforeMessage=" + beforeMessageId
	}
	if before != "" {
		apiurl += "&before=" + before
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", apiurl, nil)
	if err != nil {
		return nil, err
	}

	// Set the authorization header
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			// Handle rate limiting
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				retryAfterInt, err := strconv.Atoi(retryAfter)
				if err == nil {
					// return nil, &ErrorRetryAfter{RetryAfter: retryAfterInt}
					log.Printf("Rate limit exceeded, retrying after %d seconds\n", retryAfterInt)
					time.Sleep(time.Duration(retryAfterInt) * time.Second)
					return c.GetMessages(roomID, maxMessages, beforeMessageId, before)
				}
			}
			return nil, fmt.Errorf("rate limit exceeded")
		}
		return nil, fmt.Errorf("failed to get messages: %s", resp.Status)
	}

	// Decode the response
	var response MessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Messages, nil
}

// Message represents a message in a Webex room.
type Message struct {
	ID              string   `json:"id,omitempty"`
	ParentID        string   `json:"parentId,omitempty"`
	RoomID          string   `json:"roomId,omitempty"`
	RoomType        string   `json:"roomType,omitempty"`
	PersonID        string   `json:"personId,omitempty"`
	PersonEmail     string   `json:"personEmail,omitempty"`
	Text            string   `json:"text,omitempty"`
	Markdown        string   `json:"markdown,omitempty"`
	HTML            string   `json:"html,omitempty"`
	Files           []string `json:"files,omitempty"`
	MentionedPeople []string `json:"mentionedPeople,omitempty"`
	MentionedGroups []string `json:"mentionedGroups,omitempty"`
	Created         string   `json:"created,omitempty"`
	Updated         string   `json:"updated,omitempty"`
	IsVoiceClip     bool     `json:"isVoiceClip,omitempty"`
}

type MessagesResponse struct {
	Messages []Message `json:"items,omitempty"`
}
type Attachment struct {
	ID               string `json:"id,omitempty"`
	Description      string `json:"description,omitempty"`
	MimeType         string `json:"mimeType,omitempty"`
	Title            string `json:"title,omitempty"`
	URL              string `json:"url,omitempty"`
	Thumbnail        string `json:"thumbnail,omitempty"`
	Size             int    `json:"size,omitempty"`
	Created          string `json:"created,omitempty"`
	Modified         string `json:"modified,omitempty"`
	OwnerID          string `json:"ownerId,omitempty"`
	OwnerEmail       string `json:"ownerEmail,omitempty"`
	OwnerName        string `json:"ownerName,omitempty"`
	DownloadURL      string `json:"downloadUrl,omitempty"`
	DownloadSize     int    `json:"downloadSize,omitempty"`
	DownloadMimeType string `json:"downloadMimeType,omitempty"`
	DownloadTitle    string `json:"downloadTitle,omitempty"`
}

// Get file given the file url
func (c *Crawler) GetFile(fileUrl string) (string, []byte, error) {
	// Create a new request
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", fileUrl, nil)
	if err != nil {
		return "", nil, err
	}

	// Set the authorization header
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)
	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			// Handle rate limiting
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				retryAfterInt, err := strconv.Atoi(retryAfter)
				if err == nil {
					// return nil, &ErrorRetryAfter{RetryAfter: retryAfterInt}
					log.Printf("Rate limit exceeded, retrying after %d seconds\n", retryAfterInt)
					time.Sleep(time.Duration(retryAfterInt) * time.Second)
					return c.GetFile(fileUrl)
				}
			}
			return "", nil, fmt.Errorf("rate limit exceeded")
		}
		return "", nil, fmt.Errorf("failed to get file: %s", resp.Status)
	}
	// Read the response body
	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}
	cd := resp.Header.Get("Content-Disposition")
	// Parse the filename from the Content-Disposition header
	var filename string
	if cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			if name, ok := params["filename"]; ok {
				filename = name
			}
		}
	}
	if filename == "" {
		// Fallback to the last part of the URL
		parts := strings.Split(fileUrl, "/")
		filename = parts[len(parts)-1]
		if strings.HasSuffix(resp.Header.Get("Content-Type"), "/gif") {
			filename += ".gif"
		}
	}
	return filename, bts, nil
}
