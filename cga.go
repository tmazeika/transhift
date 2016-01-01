package main

/*
#cgo LDFLAGS: -L./lib -lhash_2_cga
#include "./lib/hash_2_cga.h"
 */
import "C"

import (
    "crypto/rand"
    "crypto/sha1"
    "fmt"
    "io/ioutil"
    "os"
    "encoding/pem"
    "unsafe"
    "time"
)

const (
    Sec = 1
)

func main() {
    generateCga()
}

func generateCga() {
    file, err := os.Open("test/download/cert.pub")

    if err != nil {
        panic(err)
    }

    privB, err := ioutil.ReadAll(file)

    if err != nil {
        panic(err)
    }

    block, _ := pem.Decode(privB)

    pub := block.Bytes

    if err != nil {
        panic(err)
    }

    ptr := unsafe.Pointer(&pub[0])

    start := time.Now()

    var hash2 *C.u8 = C.generate_hash_2(4, 2, 512, (*C.u8) (ptr))

    end := time.Now()

    fmt.Println((*[1 << 30] C.u8) (unsafe.Pointer(hash2))[:14:14])

    fmt.Println("start - end = ", end.Sub(start).Nanoseconds())

    return

    modifier := make([]byte, 16)
    rand.Read(modifier)

    var stopSearch bool

    if Sec > 0 {
        var tries uint64

        blankAndPub := append(make([]byte, 9), pub...)

        search := func(start int, offset int, doneCh chan []byte, modifier []byte) {
            for i := 0; i < start; i++ {
                for j := len(modifier) - 1; j >= 0; j-- {
                    if modifier[j] == 255 {
                        modifier[j] = 0
                    } else {
                        modifier[j]++
                        break
                    }
                }
            }

            for ! stopSearch {
                step1Sum := sha1.Sum(append(modifier, blankAndPub...))

                hash2 := step1Sum[:14]

                allZero := true

                for v := range hash2[:(2 * Sec)] {
                    if v != 0 {
                        allZero = false
                        break
                    }
                }

                if allZero {
                    doneCh <- modifier
                    break
                }

                for j := 0; j < offset; j++ {
                    for i := len(modifier) - 1; i >= 0; i-- {
                        if modifier[i] == 255 {
                            modifier[i] = 0
                        } else {
                            modifier[i]++
                            break
                        }
                    }
                }

                tries++

                if tries % 1000000 == 0 {
//                    fmt.Printf("tried %d ... %.5f%%\n", tries, float64(tries) / math.Pow(2.0, 16.0 * float64(Sec)))
                    fmt.Printf("%d: current modifier: %x\n", start, modifier)
//                    fmt.Printf("sample sha1: %x\n", step1Sum)
                }
            }
        }

        doneCh := make(chan []byte)

        for i := 0; i < 10000; i++ {
            go search(i, 10000, doneCh, modifier)
        }

        modifier = <- doneCh
        stopSearch = true
    }

    var collision uint8

    var step5 []byte

    step5 = append(step5, modifier...)
    step5 = append(step5, []byte{0x04, 0x3e, 0x41, 0x06, 0xb0, 0x76, 0x2d, 0x0e}...)
    step5 = append(step5, byte(collision))
    step5 = append(step5, pub...)

    step5Sum := sha1.Sum(step5)
    hash1 := step5Sum[:8]

    interfaceIdent := hash1
    interfaceIdent[0] = ((interfaceIdent[0] & 0x1F) | Sec) ^ 0x06

    var step7 []byte

    step7 = append(step7, []byte{0x04, 0x3e, 0x41, 0x06, 0xb0, 0x76, 0x2d, 0x0e}...)
    step7 = append(step7, interfaceIdent...)

    var cga []byte

    cga = append(cga, modifier...)
    cga = append(cga, []byte{0x04, 0x3e, 0x41, 0x06, 0xb0, 0x76, 0x2d, 0x0e}...)
    cga = append(cga, byte(collision))
    cga = append(cga, pub...)

    fmt.Printf("cga: %x\n", cga)
}
