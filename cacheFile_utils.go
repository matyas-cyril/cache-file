package cacheFile

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"syscall"
)

// isPathExist: Vérifier qu'un répertoire existe
func isPathExist(path string) error {

	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf(fmt.Sprintf("path %s is not a directory", path))
	}

	return nil
}

// isWritable: Vérifier que l'on peut écrire dans le dossier
func isWritable(path string) error {

	if err := syscall.Access(path, 0x2); err != nil {
		return fmt.Errorf(fmt.Sprintf("directory %s is not writable", path))
	}

	return nil
}

func makeDirectory(path string) error {

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil

}

func mapToByte(obj map[string][]byte) ([]byte, error) {

	var indexBuffer bytes.Buffer
	encoder := gob.NewEncoder(&indexBuffer)
	if err := encoder.Encode(obj); err != nil {
		return nil, err
	}
	return indexBuffer.Bytes(), nil
}

func byteToMap(obj []byte) (map[string][]byte, error) {

	indexBuffer := bytes.NewBuffer(obj)
	data := map[string][]byte{}
	decoder := gob.NewDecoder(indexBuffer)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func writeFile(fileName string, data []byte) error {

	if _, err := os.Stat(fileName); errors.Is(err, os.ErrExist) {
		if err := os.Remove(fileName); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.Write(data); err != nil {
		return err
	}

	return nil
}

func readFile(fileName string) ([]byte, error) {

	d, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return d, nil
}
