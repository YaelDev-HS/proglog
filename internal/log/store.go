package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian // orden de los bytes
)

const (
	lenWidth = 8
)

type store struct {
	*os.File // incrustamos el tipo para que tenga acceso a todos los metodos de os.File
	mu       sync.Mutex
	buf      *bufio.Writer // buffer para escribir en el archivo, mucho mas eficiente que escribir directamente
	size     uint64        // registro del tamaño del archivo
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name()) // obtener informacion del archivo

	if err != nil {
		return nil, err
	}

	size := uint64(fi.Size()) // es importante para saber si ya tiene tamaño y empezar a escribir desde el final

	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	// evitar que haya condicion de carrera
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size // tamaño del archivo actual para saber su posicion

	lengthPrefixing := uint64(len(p)) // tiene una longitud de 8 bytes

	if err := binary.Write(s.buf, enc, lengthPrefixing); err != nil {
		return 0, 0, err
	}

	w, err := s.buf.Write(p) // escribimos el registro justo despues del prefijo
	if err != nil {
		return 0, 0, err
	}

	w += lenWidth // sumamos los bytes usados para el prefijo

	s.size += uint64(w)

	// retornamos el tamaño del registro, la posicion y el error
	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock() // evitar que un proceso se lea mientras se escribe

	if err := s.buf.Flush(); err != nil { // verificamos que todos los datos ya esten fisicamente en el disco
		return nil, err
	}

	size := make([]byte, lenWidth) // arreglo de bytes inicial para el prefijo

	// leemos el prefijo
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

	//enc.Uint64(size) convierte el prefijo a un uint64
	// creamos un arreglo de bytes con la longitud del registro en uint64
	b := make([]byte, enc.Uint64(size))

	// comenzamos a leer a partir del prefijo
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}

	return b, nil
}
