cmake_minimum_required(
    VERSION 3.20.0
)

project(chai)

# Configuration Variables
# -----------------------

set(SRC_DIR "${CMAKE_CURRENT_SOURCE_DIR}/src")
set(VENDOR_DIR "${CMAKE_CURRENT_SOURCE_DIR}/vendor")

# Main Application Configuration
# ------------------------------

set(SRCS 
    "main.c"
    "compile.c"
    "syntax/lexer.c"
    "syntax/parser.c"
    "report/report.c"
    "depm/pkg_map.c"
    "depm/source.c"
    )
list(TRANSFORM SRCS PREPEND "${SRC_DIR}/")

add_executable(${PROJECT_NAME} ${SRCS})

target_include_directories(${PROJECT_NAME} PRIVATE ${SRC_DIR})

# Small Utility Libraries
# -----------------------

# dirent.h
target_include_directories(${PROJECT_NAME} PRIVATE "${VENDOR_DIR}/dirent/include")

