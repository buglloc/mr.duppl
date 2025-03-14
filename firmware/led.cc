#include "led.h"

#include "hardware/pio.h"
#include "ws2812.pio.h"

using namespace DPPL;

namespace
{
  PIO ledPio_ = pio1;
  uint32_t ledSm_ = 0;
  bool initialized_ = false;

  static inline uint32_t urgb_u32(uint8_t r, uint8_t g, uint8_t b) {
    return
      ((uint32_t) (r) << 8) |
      ((uint32_t) (g) << 16) |
      (uint32_t) (b);
  }
}

pico_error_codes LED::Initialize(uint32_t pin)
{
  uint32_t offset = pio_add_program(ledPio_, &ws2812_program);
  ledSm_ = pio_claim_unused_sm(ledPio_, true);
  ws2812_program_init(ledPio_, ledSm_, offset, pin, 800000, false);

  initialized_ = true;
  Inactive();

  return PICO_OK;
}

void LED::Active()
{
  if (!initialized_) {
    return;
  }

  pio_sm_put_blocking(ledPio_, ledSm_, urgb_u32(0, 0, 0x20) << 8u);
}

void LED::Inactive()
{
  if (!initialized_) {
    return;
  }

  pio_sm_put_blocking(ledPio_, ledSm_, urgb_u32(0, 0x20,0) << 8u);
}
