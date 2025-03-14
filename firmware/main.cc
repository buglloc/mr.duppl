#include <cstdint>
#include <stdio.h>
#include <string>
#include <string_view>

#include "hardware/clocks.h"
#include "pico/stdlib.h"

#include "usb.h"
#include "led.h"
#include "slip.h"


namespace
{
  constexpr uint32_t kDpPin = 7;
  constexpr uint32_t kLedPin = 16;

  constexpr uint32_t kBlinkTimeout = 150;

  constexpr size_t kUSBMaxPacketLen = 1028;
  constexpr size_t kIncommingCmdLen = 8;

  constexpr uint8_t kCmdStartCapture = 0x01;
  constexpr uint8_t kCmdStopCapture  = 0x02;

  DPPL::Slip::Stdin slipIn_;
  DPPL::Slip::Stdout slipOut_;
  bool sendPackets_ = false;
  bool usePacketFolding_ = false;

  void initializeStdio()
  {
    stdio_usb_init();
    // Disable CR/LF conversion built into stdio of Pico SDK.
    stdio_set_translate_crlf(&stdio_usb, false);
  }

  void processUsbPacket(const std::basic_string_view<uint8_t> packet)
  {
    slipOut_.WritePacket(packet);
  }

  void processCommandPacket(const std::basic_string_view<uint8_t> packet)
  {
    if (packet.empty()) {
      return;
    }

    switch (packet[0])
    {
    case kCmdStartCapture:
      sendPackets_ = true;
      if (packet.size() > 1) {
        usePacketFolding_ = packet[1] != 0x00;
      }
      break;

    case kCmdStopCapture:
      sendPackets_ = false;
      break;

    }
  }
}

int main()
{
  // Change system clock to 120 MHz (10 times the frequency of USB Full Speed)
  set_sys_clock_khz(120000, true);

  initializeStdio();
  DPPL::USB::Initialize(kDpPin);

  if (kLedPin > 0) {
    DPPL::LED::Initialize(kLedPin);
  }

  std::basic_string<uint8_t> usbPacket = {};
  usbPacket.reserve(kUSBMaxPacketLen);

  std::basic_string<uint8_t> inCommand = {};
  inCommand.reserve(kIncommingCmdLen);

  absolute_time_t packet_time = nil_time;
  while (true) {

    if (DPPL::USB::NextPacket(usePacketFolding_, usbPacket) && sendPackets_) {
      processUsbPacket(usbPacket);
      DPPL::LED::Active();
      packet_time = make_timeout_time_ms(kBlinkTimeout);
    }

    if (packet_time != nil_time && absolute_time_diff_us(get_absolute_time(), packet_time) < 0) {
      DPPL::LED::Inactive();
      packet_time = nil_time;
    }

    if (slipIn_.Read(inCommand)) {
      processCommandPacket(inCommand);
    }
  }
}
