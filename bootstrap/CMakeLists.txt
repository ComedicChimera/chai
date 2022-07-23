cmake_minimum_required(
    VERSION 3.20.0
)

project(chaic)

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
)

list(TRANSFORM SOURCES PREPEND "${CMAKE_CURRENT_SOURCE_DIR}/src/")

add_executable(${PROJECT_NAME} ${SOURCES})

target_compile_features(${PROJECT_NAME} PRIVATE cxx_std_20)

target_include_directories(${PROJECT_NAME} 
    PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/src"
    PRIVATE $ENV{BOOST_PATH}
)

target_link_directories(${PROJECT_NAME}
    PRIVATE "$ENV{BOOST_PATH}/stage/lib/"
)
