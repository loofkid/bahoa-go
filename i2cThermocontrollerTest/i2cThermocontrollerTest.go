package main

import (
	"fmt"
	"log"
    "bytes"
	"github.com/d2r2/go-i2c"
    "encoding/binary"
)

func read_from_i2c() {

}

func control_pin(pin int, state bool) {

}

func main() {
    i2c, err := i2c.NewI2C(0x42, 1)
    if err != nil { log.Fatal(err) }
    defer i2c.Close()

    full_sum := 0

    start_buffer := make([]byte, 1)
    for start_buffer[0] != 0x69 {
        start_buffer = make([]byte, 1)
        i2c.ReadBytes(start_buffer)
    }
    full_sum += int(start_buffer[0])

    for i := 0; i < 4; i++ {
        buffer := make([]byte, 7)
        _, err = i2c.ReadBytes(buffer)
        if err != nil { log.Fatal(err) }
        checksum := 0
        for _, v := range buffer {
            checksum += int(v)
            full_sum += int(v)
        }
        checksum = checksum % 256
        fmt.Println(buffer)
        sum_buffer := make([]byte, 1)
        _, err = i2c.ReadBytes(sum_buffer)
        if err != nil { log.Fatal(err) }
        full_sum += int(sum_buffer[0])
        if int(sum_buffer[0]) == checksum {
            fmt.Println("Checksum OK")
        } else {
            fmt.Println("Checksum ERROR")
        }

        fmt.Println(string(buffer[0:2]))
        var float_value float32
        float_buffer := bytes.NewReader(buffer[2:7])
        err := binary.Read(float_buffer, binary.LittleEndian, &float_value)
        if err != nil {
            fmt.Println("binary.Read failed:", err)
        }
        fmt.Printf("Temperature C: %f\n", float_value)
        farenheit := float_value * 9 / 5 + 32
        fmt.Printf("Temperature F: %f\n", farenheit)
        if err != nil { log.Fatal(err) }
        if buffer[6] == 0x01 {
            fmt.Println("Connected")
        } else {
            fmt.Println("Not Connected")
        }
    }

    end_buffer := make([]byte, 1)
    _, err = i2c.ReadBytes(end_buffer)
    if err != nil { log.Fatal(err) }

    full_sum_buffer := make([]byte, 1)
    _, err = i2c.ReadBytes(full_sum_buffer)
    if err != nil { log.Fatal(err) }

    full_sum = full_sum % 256
    fmt.Println(full_sum)
    fmt.Println(full_sum_buffer[0])
    if full_sum == int(full_sum_buffer[0]) {
        fmt.Println("Full Checksum OK")
    } else {
        fmt.Println("Full Checksum ERROR")
    }
    // done1 := make(chan bool, 1)
    
}