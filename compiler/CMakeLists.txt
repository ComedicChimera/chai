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
    "depm/mod_load.c"
    )
list(TRANSFORM SRCS PREPEND "${SRC_DIR}/")

# Add the TOMLC C file (b/c `tomlc99` doesn't have a CMakeLists.txt)
list(APPEND SRCS "${VENDOR_DIR}/tomlc/toml.c")

add_executable(${PROJECT_NAME} ${SRCS})

target_include_directories(${PROJECT_NAME} PRIVATE ${SRC_DIR})

# Small Utility Libraries
# -----------------------

# dirent.h
target_include_directories(${PROJECT_NAME} PRIVATE "${VENDOR_DIR}/dirent/include")

# tomlc (includes; C file added above)
target_include_directories(${PROJECT_NAME} PRIVATE "${VENDOR_DIR}/tomlc")

# cwalk
add_subdirectory("${VENDOR_DIR}/cwalk")
target_link_libraries(${PROJECT_NAME} cwalk)

