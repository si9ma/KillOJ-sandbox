package model

// 1xx: compile error
const INNER_COMPILER_ERR = 101 // error from real compiler
const OUTER_COMPILER_ERR = 102 // error from our outer compiler
const COMPILE_TIME_LIMIT_ERR = 103

// [2-4]xx: container error
// 2xx: error from outermost process (the process to start container)
const RUNNER_ERR = 201

// 3xx: container error (error from container)

// 4xx: run error (error from program run in container)
