# rave2gether
Little server software for a shared playlist.
Simply send a http post request with the following payload:
```
{
    "queries": ["Dicht & Ergreifend - Wandadoog"]
}
```
to the /api/queue endpoint to add a song to the queue.
The query can be either text (will be searched on youtube and downloaded) or a URL which will be tried to download as mp3 with ytdlp
## Prerequisits
ytdlp and ffmpeg
## Features
### Implemented
- Download songs and queue them
- Name of songs in queue
- Delete tracks from queue
- Up and down vote songs
- History of songs
### Planed
- User und Rights Managment
- Webinterface
