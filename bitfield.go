package main

import (
	"encoding/binary"
	"log"
	"math/bits"
	"os"
)

func countOnes(bitfield []byte) int64 {
	count := 0
	for _, b := range bitfield {
		count += bits.OnesCount8(b)
	}
	return int64(count)
}

func saveBitfield() error {
	mutex.Lock()
	defer mutex.Unlock()

	file, err := os.Create("bitfield.dat")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(bitfield)
	if err != nil {
		return err
	}

	buffer := make([]byte, binary.MaxVarintLen64)
	nWritten := binary.PutVarint(buffer, clicks)
	_, err = file.Write(buffer[:nWritten])
	if err != nil {
		return err
	}

	log.Printf("bitfield saved successfully; %d cumulative clicks", clicks)

	return nil
}

func loadBitfield() error {
	mutex.Lock()
	defer mutex.Unlock()

	file, err := os.Open("bitfield.dat")
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, no need to load anything
			return nil
		}
		return err
	}
	defer file.Close()

	_, err = file.Read(bitfield)
	if err != nil {
		return err
	}
	log.Println("... read bitfield")

	buffer := make([]byte, binary.MaxVarintLen64)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}
	log.Println("... read clicks")
	clicks, _ = binary.Varint(buffer)

	hot = countOnes(bitfield)

	log.Printf("Finished loading existing bitfield; %d cumulative clicks", clicks)
	return nil
}
