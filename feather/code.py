import time
import traceback

import adafruit_usb_host_midi
import board
import busio
import digitalio
import neopixel
import usb

# NeoPixel colors
YELLOW = (255, 128, 0)
TEAL = (0, 200, 255)
GREENISH = (0, 255, 200)
PURPLE = (200, 0, 255)

STATUS_INIT = YELLOW
STATUS_LOOKING_FOR_DEVICE = TEAL
STATUS_TRYING_DEVICE = GREENISH
STATUS_FORWARDING_MSGS = PURPLE


class App:
    def __init__(self):
        self.neopixel = neopixel.NeoPixel(board.NEOPIXEL, 1)
        self.neopixel.brightness = 0.05
        self.status(STATUS_INIT)

    def status(self, color):
        """Set the NeoPixel color"""
        self.neopixel.fill(color)

    def main(self):
        dst = busio.UART(
            tx=board.D4,
            rx=board.D5,
            baudrate=31250,
            bits=8,
            parity=None,
            stop=1,
        )
        try:
            self._main(dst)
        finally:
            dst.deinit()

    def _main(self, dst):
        print("Looking for midi device")
        while True:
            self.status(STATUS_LOOKING_FOR_DEVICE)

            for device in usb.core.find(find_all=True):
                self.status(STATUS_TRYING_DEVICE)
                print(f"Trying {device.idVendor:04x}:{device.idProduct:04x}"
                      f": {device.manufacturer} {device.product}")

                try:
                    src = adafruit_usb_host_midi.MIDI(device)
                except ValueError:
                    continue

                self.status(STATUS_FORWARDING_MSGS)
                print("Forwarding messages")
                while True:
                    try:
                        buf = src.read(1)
                    except usb.core.USBError:
                        break
                    for c in buf:
                        print(hex(c))
                    dst.write(buf)

                print("Looking for midi device")


def main():
    led = digitalio.DigitalInOut(board.LED)
    led.direction = digitalio.Direction.OUTPUT
    while True:
        try:
            App().main()
        except:
            led.value = True
            raise  # XXX

        for _ in range(6):
            time.sleep(0.4)
            led.value = False
            time.sleep(0.4)
            led.value = True
        led.value = False


if __name__ == "__main__":
    main()
