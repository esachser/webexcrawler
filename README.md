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

## How to install

```bash
go install github.com/esachser/webexcrawler/cmd/webexcrawler@latest
```

Or download the executable from releases 

## How to use

First, get a Webex API Key [here](https://developer.webex.com/docs/getting-your-personal-access-token).

On the bash/cmd/powershell, create a env variable `WEBEX_APIKEY`.

On bash:
```bash
export WEBEX_APIKEY=<THE_APIKEY_YOU_GOT>
```

On Powershell
```powershell
$env:WEBEX_APIKEY="<THE_APIKEY_YOU_GOT>"
```

To show usage of the app:

```bash
webexcrawler -h
```

If you want first to "understand" the best value of rooms.

```bash
webexcrawler -onlyrooms -rooms <N>
```

Finally, to store the messages in a selected folder:

```bash
webexcrawler -output ./myselectedwebexfolder
```

```
Usage of webexcrawler:
  -after string
        Fetch messages after this date (YYYY-MM-DDTHH:MM:SSZ)
  -nofiles
        Do not download files
  -onlyrooms
        Only fetch rooms and not messages
  -output string
        Output directory to save the rooms (default "./webexmessages")
  -roomfile string
        File containing rooms to fetch messages for
  -rooms int
        Maximum number of rooms to fetch (default 100)
```

