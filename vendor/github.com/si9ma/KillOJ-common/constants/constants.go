package constants

import "time"

const ProjectName = "KillOJ"
const DefaultLang = "en"

// redis
const SubmitStatusKeyPrefix = "submit_status_"
const SubmitStatusTimeout = time.Hour // 1 hour

// sandbox
const JavaFile = "Main.java"
const GoFile = "main.go"
const CFile = "main.c"
const CppFile = "main.cpp"
const ExeFile = "Main"
