# == DO NOT EDIT THE FOLLOWING LINES for the Raspberry Pi Pico VS Code Extension to work ==
if(WIN32)
    set(USERHOME $ENV{USERPROFILE})
else()
    set(USERHOME $ENV{HOME})
endif()
set(sdkVersion 2.0.0)
set(toolchainVersion 13_2_Rel1)
set(picotoolVersion 2.0.0)
set(picoVscode ${USERHOME}/.pico-sdk/cmake/pico-vscode.cmake)
if (EXISTS ${picoVscode})
    include(${picoVscode})
endif()
# ====================================================================================
# == DO NOT EDIT THE FOLLOWING LINES for the Raspberry Pi Pico VS Code Extension to work ==
if(WIN32)
    set(USERHOME $ENV{USERPROFILE})
else()
    set(USERHOME $ENV{HOME})
endif()
set(sdkVersion 2.0.0)
set(toolchainVersion 13_2_Rel1)
set(picotoolVersion 2.0.0)
set(picoVscode ${USERHOME}/.pico-sdk/cmake/pico-vscode.cmake)
if (EXISTS ${picoVscode})
    include(${picoVscode})
endif()
# ====================================================================================
# == DO NOT EDIT THE FOLLOWING LINES for the Raspberry Pi Pico VS Code Extension to work ==
if(WIN32)
    set(USERHOME $ENV{USERPROFILE})
else()
    set(USERHOME $ENV{HOME})
endif()
set(sdkVersion 2.0.0)
set(toolchainVersion 13_2_Rel1)
set(picotoolVersion 2.0.0)
set(picoVscode ${USERHOME}/.pico-sdk/cmake/pico-vscode.cmake)
if (EXISTS ${picoVscode})
    include(${picoVscode})
endif()
# ====================================================================================
# == DO NOT EDIT THE FOLLOWING LINES for the Raspberry Pi Pico VS Code Extension to work ==
if(WIN32)
    set(USERHOME $ENV{USERPROFILE})
else()
    set(USERHOME $ENV{HOME})
endif()
set(sdkVersion 2.0.0)
set(toolchainVersion 13_2_Rel1)
set(picotoolVersion 2.0.0)
set(picoVscode ${USERHOME}/.pico-sdk/cmake/pico-vscode.cmake)
if (EXISTS ${picoVscode})
    include(${picoVscode})
endif()
# ====================================================================================
cmake_minimum_required(VERSION 3.13)

include(pico_sdk_import.cmake)

project(mr.duppl)

set(CMAKE_C_STANDARD 11)
set(CMAKE_CXX_STANDARD 23)

add_compile_definitions(
  USBD_PID=0x5052
  USBD_MANUFACTURER="@UTBDK"
  USBD_PRODUCT="Mr.Duppl"
  # Some rp2040 super mini boards may have trouble booting properly
	# due to bad componets used in the design.
  PICO_XOSC_STARTUP_DELAY_MULTIPLIER=32
)

pico_sdk_init()

add_executable(mr.duppl
  main.cc
  led.cc
  slip.cc
  usb.cc
)

pico_generate_pio_header(mr.duppl ${CMAKE_CURRENT_LIST_DIR}/usb_sniff.pio)
pico_generate_pio_header(mr.duppl ${CMAKE_CURRENT_LIST_DIR}/ws2812.pio)

target_link_libraries(mr.duppl
  pico_stdlib
  pico_multicore
  pico_stdio_usb
  hardware_pio
  hardware_dma
  hardware_irq
)

pico_add_extra_outputs(mr.duppl)
