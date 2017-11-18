package gpio

import (
	"strconv"
	"os"
)

const GPIOPath = "/sys/class/gpio"

type IOMode int

const (
	Input IOMode = iota
	Output
)

func fileExists(fileName string) bool {

	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

func writeFile(fileName, data string) (n int, err error) {

	f, err := os.OpenFile(fileName, os.O_WRONLY, 0)
	n = 0

	if err == nil {
		defer f.Close()
		n, err = f.WriteString(data)
	}

	return n, err
}

func OpenGPIO(no int, ioMode IOMode) (f *os.File, err error) {

	noStr := strconv.Itoa(no)

	if !fileExists(GPIOPath+"/gpio"+noStr+"/value") {
		if _, err := writeFile(GPIOPath+"/export", noStr+"\n") ; err != nil {
			return nil, err
		}
	}

	dir, flags := "in\n", os.O_RDONLY
	if ioMode == Output {
		dir, flags = "out\n", os.O_WRONLY
	}

	if _, err := writeFile(GPIOPath+"/gpio"+noStr+"/direction", dir) ; err != nil {
		return nil, err
	}

	file, err := os.OpenFile(GPIOPath+"/gpio"+noStr+"/value", flags, 0)

	return file, err
}

func ReadGPIO(f *os.File) int {
	b := []byte{'0', '\n'}
	f.ReadAt(b, 0)
	if b[0] == '0' {
		return 0
	} else {
		return 1
	}
}

func WriteGPIO(f *os.File, value int) {
	if value == 0 {
		f.WriteString("1\n")
	} else {
		f.WriteString("0\n")
	}
}

