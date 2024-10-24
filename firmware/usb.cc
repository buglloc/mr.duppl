#include "usb.h"

#include "pico/multicore.h"
#include "hardware/pio.h"
#include "hardware/clocks.h"
#include "hardware/dma.h"
#include "hardware/irq.h"
#include "usb_sniff.pio.h"

using namespace DPPL;

namespace
{
  typedef struct {
    size_t Start;
    size_t Len;
  } PacketPos;

  constexpr uint32_t kQueueLen = 8192;
  constexpr uint32_t kCaptureBufLen = 8192;
  constexpr uint32_t kEOPMsg = 0xFFFFFFFF;
  
  constexpr uint8_t kUSBSync     = 0x80;
  constexpr uint8_t kUSBPidOut   = 0b0001;
  constexpr uint8_t kUSBPidIn    = 0b1001;
  constexpr uint8_t kUSBPidSof   = 0b0101;
  constexpr uint8_t kUSBPidSetup = 0b1101;
  constexpr uint8_t kUSBPidData0 = 0b0011;
  constexpr uint8_t kUSBPidData1 = 0b1011;
  constexpr uint8_t kUSBPidAck   = 0b0010;
  constexpr uint8_t kUSBPidNak   = 0b1010;
  constexpr uint8_t kUSBPidStall = 0b1110;
  constexpr uint8_t kUSBPidPre   = 0b1100;
  
  bool initialized_ = false;
  uint32_t dpPin_ = 11;
  queue_t queue_ = {};

  PIO capturePio_ = pio0;
  uint32_t captureBuf_[kCaptureBufLen] = {};
  uint32_t captureSm_ = 0;
  uint32_t captureDmaChan_ = 0;
}

namespace
{
  // Called when DMA completes transfer of data whose amount is specified in its setting
  void handle_dma_complete_interrupt()
  {
    dma_channel_acknowledge_irq0(captureDmaChan_);
    // Restart DMA
    dma_channel_set_write_addr(captureDmaChan_, captureBuf_, true);
  }

  void usb_sniff_program_init(uint32_t prodOffset)
  {
    // Get default configuration for a PIO state machine
    pio_sm_config conf = usb_sniff_program_get_default_config(prodOffset);
    // Input number 0 is assigned to USB D+ pin (GPIO number is dpPin).
    // Input number 1 is USB D- pin (GPIO number is dp_pin+1).
    sm_config_set_in_pins(&conf, dpPin_);
    // Right shift (LSB first), autopush when 8 bits are read
    sm_config_set_in_shift(&conf, true, true, 8);
    // Right shift (LSB first), no autopull
    sm_config_set_out_shift(&conf, true, false, 32);
    // 120 MHz clock (10 x 12 Mbps)
    sm_config_set_clkdiv(&conf, (float)clock_get_hz(clk_sys) / 120000000);
    // Because only RX FIFO is needed, two FIFOs are combined into single RX FIFO.
    sm_config_set_fifo_join(&conf, PIO_FIFO_JOIN_RX);

    // Allow PIO to use the specified pin
    pio_gpio_init(capturePio_, dpPin_);
    pio_gpio_init(capturePio_, dpPin_ + 1);
    // Speicify D+ and D- pins as input
    pio_sm_set_consecutive_pindirs(capturePio_, captureSm_, dpPin_, 2, false);

    // Initialize the state machine with the config created above
    pio_sm_init(capturePio_, captureSm_, prodOffset, &conf);

    // DMA configuration
    dma_channel_config chanConf = dma_channel_get_default_config(captureDmaChan_);
    // Always read from same address (RX FIFO)
    channel_config_set_read_increment(&chanConf, false);
    // Write address increases after writing each byte
    channel_config_set_write_increment(&chanConf, true);
    // Transfer 4 bytes at once
    channel_config_set_transfer_data_size(&chanConf, DMA_SIZE_32);
    // PIO SM requests DMA to transfer
    channel_config_set_dreq(&chanConf, pio_get_dreq(capturePio_, captureSm_, false));
    // Apply configuration to a DMA channel
    dma_channel_configure(
      captureDmaChan_, &chanConf,
      captureBuf_, &capturePio_->rxf[captureSm_], kCaptureBufLen,
      false  // Don't start now
    );

    // Interrupt when DMA transfer is finished
    // It is used to make DMA run forever and implement a ring buffer
    // DMA_IRQ_0 is fired when DMA completes
    dma_channel_set_irq0_enabled(captureDmaChan_, true);
    // Handler runs on current core
    irq_set_exclusive_handler(DMA_IRQ_0, handle_dma_complete_interrupt);
    // DMA interrupt has the highest priority
    irq_set_priority(DMA_IRQ_0, 0);
    irq_set_enabled(DMA_IRQ_0, true);
  }

  void usbReadLoop()
  {
    uint32_t progOffset = pio_add_program(capturePio_, &usb_sniff_program);

    usb_sniff_program_init(progOffset);

    // Start DMA
    dma_channel_start(captureDmaChan_);
    // Start the state machine
    pio_sm_set_enabled(capturePio_, captureSm_, true);

    uint32_t curPos = 0;
    uint32_t startPos = 0;

    while (true) {
      uint32_t eofPos = (((uint32_t*)(dma_hw->ch[captureDmaChan_].write_addr) - captureBuf_) % kCaptureBufLen);
      if (curPos == eofPos) {
        continue;
      }

      if (captureBuf_[curPos] == kEOPMsg) {
        // EOP is detected
        if (startPos != curPos) {
          PacketPos packet = {
            .Start = startPos
          };
          
          if (curPos > startPos) {
            packet.Len = curPos - startPos;
          } else {
            packet.Len = (kCaptureBufLen - startPos) + curPos;
          }

          // Copy packet_pos and send to Core 0
          queue_add_blocking(&queue_, &packet);
        }

        startPos = (curPos + 1) % kCaptureBufLen;
      }

      curPos = (curPos + 1) % kCaptureBufLen;
    }
  }
}

pico_error_codes USB::Initialize(uint32_t dpPin)
{
  queue_init(&queue_, sizeof(PacketPos), kQueueLen);
  captureSm_ = pio_claim_unused_sm(capturePio_, true);
  captureDmaChan_ = dma_claim_unused_channel(true);
  dpPin_ = dpPin;

  // Starting reading loop on the another core1
  multicore_launch_core1(usbReadLoop);

  return PICO_OK;
}

bool USB::NextPacket(bool packetFolding, std::basic_string<uint8_t>& out)
{
  static PacketPos packet = {};
  static uint8_t lastPid = 0x00;

  if (!queue_try_remove(&queue_, &packet)) {
    return false;
  }

  uint8_t firstByte = captureBuf_[packet.Start] >> 24;
  if (firstByte != kUSBSync || packet.Len < 3) {
    // skips invalid packets
    return USB::NextPacket(packetFolding, out);
  }

  uint8_t secondByte = captureBuf_[(packet.Start + 1) % kCaptureBufLen] >> 24;
  if (((~(secondByte >> 4)) & 0xF) != (secondByte & 0xF)) {
    // Skip invalid packet which has a broken PID byte (First 4 bits are not bit-inversion of the rest)
    return USB::NextPacket(packetFolding, out);
  }

  uint8_t curPid = secondByte & 0xF;
  if (packetFolding) {
    if (curPid == kUSBPidSof) {
      return USB::NextPacket(packetFolding, out);
    }

    if (curPid == lastPid) {
      switch (curPid) {
      case kUSBPidOut:
        [[fallthrough]];
      case kUSBPidIn:
        [[fallthrough]];
      case kUSBPidAck:
        [[fallthrough]];
      case kUSBPidNak:
        return USB::NextPacket(packetFolding, out);
      }
    }
  }
  lastPid = curPid;

  out.clear();
  for (size_t i = 1; i < packet.Len; ++i) {
    out += captureBuf_[(packet.Start + i) % kCaptureBufLen] >> 24;
  }

  return true;
}
