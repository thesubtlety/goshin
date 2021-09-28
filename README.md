# GoShIn

Goshin is a tool that makes it easier to leverage command injection. Copy the vulnerable request with Burp's or a browser's Copy as cURL and specify the injected parameter to be replaced. If the output is messy, use Inspect Output and Copy as Xpath for cleaner output. Uses `chzyer/readline` for a psuedo shell like experience with history and prompt shortcuts.

```
                _____ __    ____
    ____ _____ / ___// /_  /  _/___
   / __ '/ __ \\__ \/ __ \ / // __ \
  / /_/ / /_/ /__/ / / / // // / / /
  \__, /\____/____/_/ /_/___/_/ /_/
 /____/                 @thesubtlety

Usage: ./build/goshin-linux-amd64
  -H value
        header values, e.g. 'JSESSIONID: 1234', use additional -H for each header value
  -compressed
        flag ignored for ease of copy paste from browser's Copy as cURL
  -data string
        data to be POSTed
  -encoding string
        encoding (base64 or URL) (default "URL")
  -t string
        target host and URL to query, vuln param should be param=^CMD^
  -v    verbose mode
  -xpath string
        xPath selector, e.g '//*[@id="main_body"]/div/div/pre'

Examples
goshin -t http://localhost/exec?cmd=^CMD^
goshin -t http://localhost:8443/exec -data "{'exec': '^CMD^'}"
goshin -t https://localhost/exec?cmd=^CMD^' -H "Authorization: Basic asdf" -xpath '//*[@id="main_body"]/div/div/pre'

```

Tested with DVWA, YMMV.
