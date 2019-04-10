# KillOJ-sandbox

sandbox for KillOJ

## Usage

```go
go build -o kbox
```

### Compile

```go
./kbox compile --lang="c" --src="main.c" --base-dir="./"
```

## Example

use log:

```go
./kbox --debug --log="./log" --log-format="json" compile --lang="c" --src="main.c" --base-dir="./" 
```

compile result:
```json
{
  "resType": "COMPILE",
  "status": 0,
  "msg": "compile success"
}
```

log:
```json
{
  "baseDir": "./",
  "lang": "c",
  "level": "info",
  "msg": "compile result",
  "result": {
    "resType": "COMPILE",
    "status": 0,
    "msg": "compile success"
  },
  "src": "main.c",
  "time": "2019-04-10T16:55:41+08:00",
  "timeout": "100000"
},
{
  "baseDir": "./",
  "lang": "",
  "level": "info",
  "msg": "compile result",
  "result": {
    "resType": "COMPILE",
    "status": 1,
    "errno": 102,
    "msg": "./kbox: \"compile\" require parameter:--lang"
  },
  "src": "main.c",
  "time": "2019-04-10T16:56:53+08:00",
  "timeout": "100000"
}

```


