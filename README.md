# KillOJ-sandbox

sandbox for KillOJ

## Install

```bash
make install
```

OR, Docker:

```bash
docker build -t si9ma/kbox:1.0 .
```

## Compile

```bash
kbox compile --src="main.c" --base-dir="/tmp/kbox" --lang="c" 
```

OR, Docker:

```bash
docker run --rm -v "$PWD":/tmp/kbox si9ma/kbox:1.0 compile --src="main.c" --base-dir="/tmp/kbox" --lang="c" 
```

## Run

```bash
kbox run --dir="/tmp/kbox" --cmd="/Main" --expected="hello" --input="hello" --timeout=1000 --memory=1000 --seccomp 
```

OR, Docker

```bash
docker run --rm -v "$PWD":/tmp/kbox si9ma/kbox:1.0 run --dir="/tmp/kbox" --cmd="/Main" --expected="hello" --input="hello" --timeout=1000 --memory=1000 --seccomp 
```

## TODO

- [x] container
- [ ] fix seccomp bug
