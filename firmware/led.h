#pragma once

#include <stdlib.h>
#include <stdint.h>
#include <string>
#include <string_view>

#include "pico/types.h"
#include "pico/error.h"


namespace DPPL::LED
{
  pico_error_codes Initialize(uint32_t pin);
  void Active();
  void Inactive();
}
