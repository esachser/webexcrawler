package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/esachser/webexcrawler"
)

func main() {

	maxRooms := 0
	flag.IntVar(&maxRooms, "rooms", 100, "Maximum number of rooms to fetch")

	output := ""
	flag.StringVar(&output, "output", "./webexmessages", "Output directory to save the rooms")

	onlyRooms := false
	flag.BoolVar(&onlyRooms, "onlyrooms", false, "Only fetch rooms and not messages")

	flag.Parse()

	// Initialize the Crawler
	crawler := webexcrawler.NewCrawler()

	// Get the rooms
	rooms, err := crawler.GetRooms(maxRooms)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the rooms
	for _, room := range rooms {
		fmt.Printf("Room ID: %s, Title: %s, Last Activity: %s\n", room.ID, room.Title, room.LastActivity)
	}

	err = os.MkdirAll(output, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		fmt.Println("Error creating output directory:", err)
		return
	}

	roomsPath := fmt.Sprintf("%s/rooms.json", output)
	file, err := os.Create(roomsPath)
	if err != nil {
		fmt.Println("Error creating rooms file:", err)
		return
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(rooms)
	if err != nil {
		fmt.Println("Error writing rooms to file:", err)
		file.Close()
		return
	}
	fmt.Printf("Rooms saved to %s\n", roomsPath)
	file.Close()

	if onlyRooms {
		fmt.Println("Only rooms were requested, exiting.")
		return
	}

	for _, room := range rooms {
		err = os.MkdirAll(fmt.Sprintf("%s/%s-%s/content", output, room.Title, room.ID), os.ModePerm)
		if err != nil && !os.IsExist(err) {
			fmt.Println("Error creating room directory:", err)
			return
		}

		messages, err := crawler.GetMessages(room.ID, 100, "")
		if err != nil {
			fmt.Println("Error fetching messages for room:", room.ID, err)
			continue
		}

		messagesPath := fmt.Sprintf("%s/%s-%s/messages.json", output, room.Title, room.ID)
		file, err = os.Create(messagesPath)
		if err != nil {
			fmt.Println("Error creating messages file:", err)
			return
		}

		fmt.Fprintf(file, "{\n")
		fmt.Fprintf(file, "  \"messages\": [\n")

		isFirst := true

		encoder = json.NewEncoder(file)
		encoder.SetIndent("    ", "  ")
		for len(messages) > 0 {
			for _, message := range messages {
				if !isFirst {
					fmt.Fprintf(file, "    ,")
				} else {
					fmt.Fprintf(file, "    ")
				}
				isFirst = false
				if len(message.Files) > 0 {
					// Gets the files to the disk
					for i, file := range message.Files {
						fname, bts, err := crawler.GetFile(file)
						if err != nil {
							fmt.Println("Error downloading file:", err)
							continue
						}
						err = os.WriteFile(fmt.Sprintf("%s/%s-%s/content/%s", output, room.Title, room.ID, fname), bts, 0644)
						if err != nil {
							fmt.Println("Error writing file to disk:", err)
							continue
						}
						message.Files[i] = fmt.Sprintf("./content/%s", fname)
					}
				}

				err = encoder.Encode(message)
				if err != nil {
					fmt.Println("Error writing message to file:", err)
					file.Close()
					return
				}
			}

			if len(messages) < 100 {
				break
			}

			// Get the next page of messages
			lastMessageID := messages[len(messages)-1].ID

			messages, err = crawler.GetMessages(room.ID, 100, lastMessageID)
			if err != nil {
				fmt.Println("Error fetching messages for room:", room.ID, err)
				continue
			}
		}
		fmt.Fprintf(file, "  ]\n")
		fmt.Fprintf(file, "}\n")
		file.Close()
		fmt.Printf("Messages for room %s saved to %s\n", room.ID, messagesPath)
	}
}
