package model

// 1xx: compile error
const INNER_COMPILER_ERR = 101 // error from real compiler
const OUTER_COMPILER_ERR = 102 // error from our outer compiler
const COMPILE_TIMEOUT = 103

// [2-4]xx: container error
// 2xx: error from outermost process (the process to start container)
const RUNNER_ERR = 201

// 3xx: container error (error from container)
const CONTAINER_ERR = 301

// 4xx: run error (error from program run in container)
const APP_ERR = 401
const UNEXPECTED_RES_ERR = 402
const OUT_OF_MEMORY = 403
const RUN_TIMEOUT = 404
const BAD_SYSTEMCALL = 405
const NO_ENOUGH_PID = 406
const JAVA_SECURITY_MANAGER_ERR = 407
