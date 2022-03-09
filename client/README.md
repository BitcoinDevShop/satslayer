## how to run

i'm using pnpm for no good reason, feel free to use npm instead:

`pnpm install`
then
`pnpm run dev`

the frontend has these hardcoded urls for the go servers:
```
const clientBaseUrl = "http://localhost:8080";
const serverBaseUrl = "http://localhost:8083";
socket = new WebSocket("ws://localhost:8083/subscribetx");
```