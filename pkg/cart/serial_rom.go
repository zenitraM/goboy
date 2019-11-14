package cart

import "fmt"
import "log"
import "io"
import "time"
import "github.com/jacobsa/go-serial/serial"

// NewSerialROM returns a new ROM cartridge.
func NewSerialROM() BankingController {

    options := serial.OpenOptions{
      PortName: "/dev/ttyUSB0",
      BaudRate: 1100000,
      DataBits: 8,
      StopBits: 1,
      MinimumReadSize: 1,
    }

    // Open the port.
    port, err := serial.Open(options)
    if err != nil {
      log.Fatalf("serial.Open: %v", err)
	}
	port.Write([]byte{'G'})
	time.Sleep(100 * time.Microsecond)
	port.Write([]byte{'5', '\x00'})
	//port.Write([]byte{'M', '1', '\x00'})
	time.Sleep(100 * time.Microsecond)
	return &SerialROM{
		port: port,
		romState: make(map[uint16][]byte),
	}
}

// SerialROM uses a GBXCart to read ROM data from an actual GB cartridge.

// romState keeps a cache of ROMs contents as to not have to read over and over. It gets deleted, stupidly, on every write to RAM/ROM
// for mappers to work

type SerialROM struct {
	port io.ReadWriteCloser
	romState map[uint16][]byte
//	lastBuf []byte
//	lastBufPos uint16
}

// Read returns a value at a memory address in the ROM.
func (r *SerialROM) Read(address uint16) (ret byte) {
//	if address >= 0x4000 {
//	log.Printf("read byte at pos %02x", address)
//	}
//	if address == 0x4000 {
//		return 1
//	}
//	if address == 0x4001 {
//		return 0xad
//	}
//	if address == 0x4104 {
//		return 0x1
//	}
	posToRead := address-(address%64)
//	log.Printf("read request ROM %x", address)
	if _, ok := r.romState[posToRead]; !ok {
	//position is not in cache - read it!
		r.port.Write([]byte("0"))
		r.port.Write([]byte(fmt.Sprintf("A%04x\x00", posToRead)))
		r.port.Write([]byte{'R'})
		r.romState[posToRead] = make([]byte, 64)
		_, err := io.ReadFull(r.port, r.romState[posToRead])
		if err != nil {
		log.Fatalf("serial.Read: %v", err)
		}
		r.port.Write([]byte("0"))
	}
	ret = r.romState[posToRead][address%64]
	//delete(r.romState, posToRead)
	//log.Printf("read byte at pos %02x: %d", address, ret)

	return ret
}

// WriteROM would switch between cartridge banks, however a ROM cart does
// not support banking.
func (r *SerialROM) WriteROM(address uint16, value byte) {
	log.Printf("write to ROM %04x value %x", address, value)
	r.port.Write([]byte(fmt.Sprintf("B%x\x00", address)))
	time.Sleep(1000 * time.Microsecond)
	r.port.Write([]byte(fmt.Sprintf("B%d\x00", value)))
	time.Sleep(1000 * time.Microsecond)
//invalidate ROM caches
	r.romState = make(map[uint16][]byte)
}

// WriteRAM writes data to the cartridge RAM
func (r *SerialROM) WriteRAM(address uint16, value byte) {
	log.Printf("write to RAM %04x value %x", address, value)
	r.port.Write([]byte(fmt.Sprintf("B%x\x00", address)))
	time.Sleep(1000 * time.Microsecond)
	r.port.Write([]byte(fmt.Sprintf("B%d\x00", value)))
	time.Sleep(1000 * time.Microsecond)
//invalidate ROM caches
	r.romState = make(map[uint16][]byte)
}

// GetSaveData returns the save data for this banking controller to persist it to a file.
// We don't care about this, so we return a noop.
func (r *SerialROM) GetSaveData() []byte {
	return []byte{}
}

// LoadSaveData loads the save data into the cartridge. As RAM is not supported
// on this memory controller, this is a noop.
func (r *SerialROM) LoadSaveData([]byte) {}
