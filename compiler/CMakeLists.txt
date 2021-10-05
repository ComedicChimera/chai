cmake_minimum_required(
    VERSION 3.20.0
)

project(chai)

# Configuration Variables
set(SRC_DIR "${CMAKE_CURRENT_SOURCE_DIR}/src")

#  Main Application Configuration
set(SRCS 
    "main.c"
    "compile.c"
    "syntax/lexer.c"
    "report/report.c"
    )
list(TRANSFORM SRCS PREPEND "${SRC_DIR}/")

add_executable(${PROJECT_NAME} ${SRCS})

target_include_directories(${PROJECT_NAME} PRIVATE ${SRC_DIR})
