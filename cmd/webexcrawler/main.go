package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/esachser/webexcrawler"
)

var regexFname = regexp.MustCompile("[/\\?%*:|\"<>]")

func main() {

	maxRooms := 0
	flag.IntVar(&maxRooms, "rooms", 100, "Maximum number of rooms to fetch")

	output := ""
	flag.StringVar(&output, "output", "./webexmessages", "Output directory to save the rooms")

	onlyRooms := false
	flag.BoolVar(&onlyRooms, "onlyrooms", false, "Only fetch rooms and not messages")

	nofiles := false
	flag.BoolVar(&nofiles, "nofiles", false, "Do not download files")

	roomfile := ""
	flag.StringVar(&roomfile, "roomfile", "", "File containing rooms to fetch messages for")

	after := ""
	flag.StringVar(&after, "after", "", "Fetch messages after this date (YYYY-MM-DDTHH:MM:SSZ)")

	flag.Parse()

	// Initialize the Crawler
	crawler := webexcrawler.NewCrawler()

	rooms := []webexcrawler.Room{}
	var err error

	if roomfile != "" {
		// Read the room IDs from the file
		bts, err := os.ReadFile(roomfile)
		if err != nil {
			log.Println("Error reading room file:", err)
			return
		}
		err = json.Unmarshal(bts, &rooms)
		if err != nil {
			log.Println("Error unmarshalling room file:", err)
			return
		}
	} else {
		// Get the rooms
		log.Println("Fetching rooms...")
		rooms, err = crawler.GetRooms(maxRooms)
		if err != nil {
			log.Println("Error:", err)
			return
		}
	}

	// Print the rooms
	for _, room := range rooms {
		log.Printf("Room ID: %s, Title: %s, Last Activity: %s\n", room.ID, room.Title, room.LastActivity)
	}

	err = os.MkdirAll(output, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Println("Error creating output directory:", err)
		return
	}

	roomsPath := fmt.Sprintf("%s/rooms.json", output)
	file, err := os.Create(roomsPath)
	if err != nil {
		log.Println("Error creating rooms file:", err)
		return
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(rooms)
	if err != nil {
		log.Println("Error writing rooms to file:", err)
		file.Close()
		return
	}
	log.Printf("Rooms saved to %s\n", roomsPath)
	file.Close()

	if onlyRooms {
		log.Println("Only rooms were requested, exiting.")
		return
	}

	afterTm := time.Now().Add(-10 * 365 * 24 * time.Hour)
	if after != "" {
		afterTm, err = time.Parse(time.RFC3339, after)
		if err != nil {
			log.Println("Error parsing after date:", err)
			return
		}
	}

	for _, room := range rooms {
		lastActivity, err := time.Parse(time.RFC3339, room.LastActivity)
		if lastActivity.Before(afterTm) && err == nil {
			log.Printf("Room %s has no activity after %s, skipping all others.\n", room.Title, after)
			break
		}

		room.Title = regexFname.ReplaceAllString(room.Title, "-")
		roomDir := fmt.Sprintf("%s/%s-%s/content", output, room.Title, room.ID)
		err = os.MkdirAll(roomDir, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			log.Println("Error creating room directory:", err)
			return
		}

		log.Printf("Room directory created: %s\n", roomDir)
		log.Println("Fetching messages for room:", room.Title)
		messages, err := crawler.GetMessages(room.ID, 100, "", "")
		if err != nil {
			log.Println("Error fetching messages for room:", room.ID, err)
			continue
		}

		messagesPath := fmt.Sprintf("%s/%s-%s/messages.json", output, room.Title, room.ID)
		file, err = os.Create(messagesPath)
		if err != nil {
			log.Println("Error creating messages file:", err)
			return
		}

		fmt.Fprintf(file, "{\n")
		fmt.Fprintf(file, "  \"messages\": [\n")

		isFirst := true

		encoder = json.NewEncoder(file)
		encoder.SetIndent("    ", "  ")
	messages:
		for len(messages) > 0 {
			for _, message := range messages {
				tm2parse := message.Created
				if message.Updated != "" {
					tm2parse = message.Updated
				}
				if tm2parse == "" {
					log.Println("Message has no created or updated time:", message)
					continue
				}

				// Parse the message time
				tm, err := time.Parse(time.RFC3339, tm2parse)
				if err != nil {
					log.Println("Error parsing message time:", err)
					continue
				}
				// Check if the message is after the specified time
				if after != "" {
					if tm.Before(afterTm) {
						break messages
					}
				}

				if !isFirst {
					fmt.Fprintf(file, "    ,")
				} else {
					fmt.Fprintf(file, "    ")
				}
				isFirst = false
				if !nofiles && len(message.Files) > 0 {
					// Gets the files to the disk
					for i, file := range message.Files {
						log.Println("Downloading file:", file)
						fname, bts, err := crawler.GetFile(file)
						if err != nil {
							log.Println("Error downloading file:", file, err)
							continue
						}
						log.Printf("File: %s, Size: %d\n", fname, len(bts))
						err = os.WriteFile(fmt.Sprintf("%s/%s-%s/content/%s", output, room.Title, room.ID, fname), bts, 0644)
						if err != nil {
							log.Println("Error writing file to disk:", fname, err)
							continue
						}
						message.Files[i] = fmt.Sprintf("./content/%s", fname)
					}
				}

				err = encoder.Encode(message)
				if err != nil {
					log.Println("Error writing message to file:", err)
					file.Close()
					return
				}
			}

			if len(messages) < 100 {
				break
			}

			// Get the next page of messages
			lastMessageID := messages[len(messages)-1].ID

			log.Println("Fetching more 100 messages for room:", room.Title)
			messages, err = crawler.GetMessages(room.ID, 100, lastMessageID, "")
			if err != nil {
				log.Println("Error fetching messages for room:", room.ID, err)
				continue
			}
		}
		fmt.Fprintf(file, "  ]\n")
		fmt.Fprintf(file, "}\n")
		file.Close()
		log.Printf("Messages for room [%s] saved to [%s]\n", room.Title, messagesPath)
	}
}
