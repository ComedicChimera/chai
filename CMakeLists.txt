cmake_minimum_required(
    VERSION 3.20.0
)

# Main project configuration
# --------------------------
project(chai)
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Sources
set(SRC_DIR "${CMAKE_CURRENT_SOURCE_DIR}/src")
set(SRCS 
    "main.cpp"
    "tokenize/scanner.cpp"
    "tokenize/patterns.cpp"
    "report/display.cpp"
    "depm/loader.cpp"
    )
list(TRANSFORM SRCS PREPEND "${SRC_DIR}/")
add_executable(${PROJECT_NAME} ${SRCS})

# Include/lib configuration
target_include_directories(${PROJECT_NAME} PRIVATE ${SRC_DIR})

# Vendor library directories
set(VENDOR_DIR "${CMAKE_CURRENT_SOURCE_DIR}/vendor")


# LLVM configuration
# ------------------
if (NOT DEFINED LLVM_DIR)
    set(LLVM_DIR "D:/DevTools/llvm/out")
endif()

if (DEFINED LLVM_BIN_DIR_SUFFIX)
    set(LLVM_BIN_DIR "${LLVM_DIR}/${LLVM_BIN_DIR_SUFFIX}")
else()
    set(LLVM_BIN_DIR "${LLVM_DIR}/Debug/bin")
endif()

if (WIN32)
    set(LLVM_CONFIG_PATH "${LLVM_BIN_DIR}/llvm-config.exe")
else()
    set(LLVM_CONFIG_PATH "${LLVM_BIN_DIR}/llvm-config")
endif()

# get the required link libraries
exec_program(${LLVM_CONFIG_PATH}
    ARGS "--libs"
    OUTPUT_VARIABLE LLVM_LIBS 
)
separate_arguments(LLVM_LIBS)
target_link_libraries(${PROJECT_NAME} PRIVATE ${LLVM_LIBS})
target_link_directories(${PROJECT_NAME} PRIVATE "${VENDOR_DIR}/lib")

# get the required system link libraries (eg. psapi.lib)
exec_program(${LLVM_CONFIG_PATH}
    ARGS "--system-libs"
    OUTPUT_VARIABLE LLVM_SYS_LIBS
)

separate_arguments(LLVM_SYS_LIBS)
target_link_libraries(${PROJECT_NAME} PRIVATE ${LLVM_SYS_LIBS})

# get the llvm include dir
exec_program(${LLVM_CONFIG_PATH}
    ARGS "--includedir"
    OUTPUT_VARIABLE LLVM_INCLUDE_PATH
)

target_include_directories(${PROJECT_NAME} PRIVATE ${LLVM_INCLUDE_PATH})
target_include_directories(${PROJECT_NAME} PRIVATE "${LLVM_DIR}/include")


# Other libraries
# ---------------

# TOML++
target_include_directories(${PROJECT_NAME} PRIVATE "${VENDOR_DIR}/tomlplusplus/include")
