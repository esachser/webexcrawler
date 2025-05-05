# webexcrawler
Webex crawler to download messages from Webex

It's a simple crawler to get a backup of Webex messages.

Basically, what is done:
- Create a folder to store the messages
- List the N most recent rooms (direct, groups, teams)
  - Creates a file with the result (id, name)
- Creates a folder for each room using the id of the room
- For each Room
  - Gets the messages of the room and saves as JSON
  - Gets the file attachments and saves to a folder called content

