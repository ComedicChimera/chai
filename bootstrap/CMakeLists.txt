cmake_minimum_required(
    VERSION 3.20.0
)

project(chaic)

# Collect all the source files.
set(SOURCES 
    "main.cpp" 
    "args.cpp"
    "report.cpp"
    "analyze.cpp" 
    "syntax/lexer.cpp" 
    "types/type_repr.cpp" 
    "types/type_eq.cpp" 
    "syntax/parser.cpp"
    "syntax/parse_top.cpp"
    "syntax/parse_common.cpp"
    "resolve/symbol_table.cpp"
    "types/type_store.cpp"
)

list(TRANSFORM SOURCES PREPEND "${CMAKE_CURRENT_SOURCE_DIR}/src/")

# Create our executable.
add_executable(${PROJECT_NAME} ${SOURCES})

# Set for C++ 20
target_compile_features(${PROJECT_NAME} PRIVATE cxx_std_20)

# Add our source include directory and the Boost include directory.
target_include_directories(${PROJECT_NAME} 
    PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/src"
    PRIVATE ${BOOST_DIR}
)

# Add Boost.
target_link_directories(${PROJECT_NAME}
    PRIVATE ${BOOST_LIB_DIR}
)

# Find LLVM.
find_package(LLVM REQUIRED CONFIG PATHS "${LLVM_INSTALL_PREFIX}/lib/cmake/llvm")

# Configure LLVM.
target_include_directories(${PROJECT_NAME} PRIVATE ${LLVM_INCLUDE_DIRS})

separate_arguments(LLVM_DEFINITIONS_LIST NATIVE_COMMAND ${LLVM_DEFINITIONS})
target_compile_options(${PROJECT_NAME} PRIVATE ${LLVM_DEFINITIONS_LIST})

llvm_map_components_to_libnames(LLVM_LIBS core support)

target_link_libraries(${PROJECT_NAME} ${LLVM_LIBS})
