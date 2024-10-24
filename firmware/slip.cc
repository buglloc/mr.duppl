// Packetization for serial communicaton using byte stuffing method of Serial Line Internet Protocol (SLIP)
// Reference:
//  Serial Line Internet Protocol - Wikipedia, https://en.wikipedia.org/w/index.php?title=Serial_Line_Internet_Protocol&oldid=1064759583
//  RFC 1055 - Nonstandard for transmission of IP datagrams over serial lines: SLIP, https://datatracker.ietf.org/doc/html/rfc1055

#include "slip.h"
#include <utility>

#include "pico/stdio.h"


using namespace DPPL;

namespace
{
  enum SlipCode : uint8_t
  {
    SLIP_CODE_END      = 0xC0,
    SLIP_CODE_ESC      = 0xDB,
    SLIP_CODE_ESC_END  = 0xDC,
    SLIP_CODE_ESC_ESC  = 0xDD
  };
}

Slip::Stdin::Stdin()
{
  this->readBuf_.reserve(64);
}

bool Slip::Stdin::Read(std::basic_string<uint8_t>& out)
{
  this->Consume();
  if (!this->hasFrame_ || this->readBuf_.empty()) {
    return false;
  }

  out.clear();
  out.reserve(this->readBuf_.size());
  out.insert(out.begin(), this->readBuf_.begin(), this->readBuf_.end());
  return true;
}

void Slip::Stdin::Consume()
{
  if (this->hasFrame_ == true) {
    this->hasFrame_ = false;
    this->readBuf_.clear();
  }

  int input = getchar_timeout_us(0 /* no wait */);
  if (input < 0 || input > 255) {
    return;
  }

  uint8_t c = static_cast<uint8_t>(input);
  switch (c) {
  case SLIP_CODE_ESC:
    this->inEsc_ = true;    
    break;

  case SLIP_CODE_END:
    this->hasFrame_ = true;
    break;

  case SLIP_CODE_ESC_END:
    if (!this->inEsc_) {
      break;
    }

    this->readBuf_ += SLIP_CODE_END;
    this->inEsc_ = false;
    break;

  case SLIP_CODE_ESC_ESC:
    if (!this->inEsc_) {
      break;
    }

    this->readBuf_ += SLIP_CODE_ESC;
    this->inEsc_ = false;
    break;

  default:
    this->inEsc_ = false;
    this->readBuf_ += c;
    break;
  }
}

pico_error_codes Slip::Stdout::WritePacket(const std::basic_string_view<uint8_t> packet)
{
  for (uint8_t c : packet) {
    switch (c) {
    case SLIP_CODE_END:
      fputc(SLIP_CODE_ESC, stdout);
      fputc(SLIP_CODE_ESC_END, stdout);
      break;
    
    case SLIP_CODE_ESC:
      fputc(SLIP_CODE_ESC, stdout);
      fputc(SLIP_CODE_ESC_ESC, stdout);
      break;

    default:
      fputc(c, stdout);
      break;
    }
  }

  fputc(SLIP_CODE_END, stdout);
  fflush(stdout);
  return PICO_OK;
}
