# rave2gether
Little server software for a shared playlist.
Simply send a http post request with the following payload:
```
{
    "query": "Dicht & Ergreifend - Wandadoog"
}
```
to the /queue endpoint to add a song to the queue.
The query can be either text (will be searched on youtube and downloaded) or a URL which will be tried to download as mp3 with ytdlp
## Prerequisits
ytdlp and ffmpeg
## Features
### Implemented
- Download songs and queue them
### Planed
- Name of songs in queue
- Delete tracks from queue
- Up and down vote songs
- User und Rights Managment
- Webinterface
