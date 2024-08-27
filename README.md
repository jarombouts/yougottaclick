# YOU GOTTA CLICK

Over one million boxes to flip! 
If you click a light square, it turns dark and you get one point. If you click a dark square, it goes light and your score resets.
Because the playing field is the same for everyone, you'd better watch out for other players on the field!

The idea for this site was stolen from the excellent onemillioncheckboxes.com made by <a href="https://onemillioncheckboxes.com">eieio</a>, and (poorly) recycled into this. 
I got quite obsessed by his <a href="https://eieio.games/essays/scaling-one-million-checkboxes">amazing writeup</a> on how he built One Million Checkboxes, and felt that I had to build something like this myself.
Also, I've wanted to learn Go for a while now, and this seemed like the perfect excuse to do so.

## GETTING STARTED

This repo takes an extremely... *ozempic* approach to architecture. Everything is contained in a single go binary. Just do `go run .` (or `go build` if you want a binary) to start the server, which listens on port 8008.


## WHAT'S HAPPENING

The server exposes several things:

- The frontend, a single file `./static/index.html`. It's terrible. I don't know any Golang, but my frontend skills are even worse. 
  LLM-du-jour did most of the work, and came up with many terrible ideas until I told it to piss off with the endless <div> soup.
  It's using a Canvas to draw all the squares, which is sort of messy but easy to understand. 
  Using it reminded me a bit of the Processing language for visual art, which is Very Fun.
- The current global state of all 1024x1024 boxes at /state, as a base64 encoded wad of binary
- A websocket connection that...  
  - Periodically sends a fresh global state object to the client (identical to what you GET at /state) 
  - Accepts flips from the frontend (you click a square and it asks the server to 'flip bit 1234 please')
  - Periodically sends state updates, diffs since the last full state update. These occur at a much higher rate, but they're smaller in size.

There's some heavy mutex-ing and goroutine-ing going on in the code... and it's definitely not very clean. 
But coming from Python at my day job, it's just wonderful how you get crazy performance and seamless concurrency with zero headaches 
(LOOKING AT YOU PYTHON `multiprocessing` OR `asyncio`)

Every 60 seconds, the current bitfield and the cumulative number of clicks are written to disk, so the state persists (somewhat; minus all the edits since saving of course) between restarts.
