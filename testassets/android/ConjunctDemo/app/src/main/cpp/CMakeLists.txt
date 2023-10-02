cmake_minimum_required(VERSION 3.22.1)

# Keep the CACHE STRING so that we can pass them to CMakeLists through
# build.gradle as a -D flag
set(CONJUNCT_BIN_PATH "" CACHE STRING "")
set(CONJUNCT_FLAGS "" CACHE STRING "")

project(
  libconjunctdemo
  VERSION 1.0
  LANGUAGES CXX)

if (CONJUNCT_BIN_PATH STREQUAL "")
  message("CONJUNCT_BIN_PATH not set")
else()
  message("CONJUNCT_BIN_PATH is ${CONJUNCT_BIN_PATH}")
  set(CMAKE_C_COMPILER "${CONJUNCT_BIN_PATH}")
  set(CMAKE_CXX_COMPILER "${CONJUNCT_BIN_PATH}")
endif()

if (CONJUNCT_FLAGS STREQUAL "")
  message("CONJUNCT_FLAGS not set")
else()
  message("CONJUNCT_FLAGS is ${CONJUNCT_FLAGS}")
  set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${CONJUNCT_FLAGS}")
  set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${CONJUNCT_FLAGS}")
endif()

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

add_library(${PROJECT_NAME} SHARED native-lib.cpp)

set_property(TARGET ${PROJECT_NAME} PROPERTY CXX_STANDARD 17)

# This forces the compiler to make libxxx.so, not liblibxxx.so
set_target_properties(${PROJECT_NAME} PROPERTIES PREFIX "")

# Used for debugging
find_library(log-lib log)
target_link_libraries(${PROJECT_NAME} ${log-lib})
