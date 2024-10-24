#pragma once

#include <stdlib.h>
#include <stdint.h>
#include <string>
#include <string_view>

#include "pico/error.h"
#include "usb.h"


namespace DPPL::Slip
{
  class Stdin
  {
  public:
    explicit Stdin();
    bool Read(std::basic_string<uint8_t>& out);

  protected:
    void Consume();

  protected:
    std::basic_string<uint8_t> readBuf_ = {};
    bool hasFrame_ = false;
    bool inEsc_ = false;
  };

  class Stdout
  {
  public:
    explicit Stdout() = default;

    pico_error_codes WritePacket(const std::basic_string_view<uint8_t> packet);
  };
}
