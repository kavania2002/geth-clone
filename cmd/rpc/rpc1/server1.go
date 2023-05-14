package main

import (
    // "bytes"
    "fmt"
    "log"
    "time"

    "github.com/torquem-ch/mdbx-go/mdbx"
)

func main() {
    env, err1 := mdbx.NewEnv()
    fmt.Println("Environment Created ", env, err1)

    if err1 != nil {
        fmt.Println("Cannot Open Environment")
    }

    err := env.Open("/home/kavania2002/.ethereum/geth/chaindata", 0, 0664)
    defer env.Close()
    if err != nil {
        fmt.Println("Cannot Use Open function")
    }

    err = env.View(func(txn *mdbx.Txn) error {
        // fmt.Println("GET INITIATED")
        dbi, err := txn.OpenRoot(0)
        if err != nil {
            return err
        }

        cur := mdbx.CreateCursor()
        err = cur.Bind(txn, dbi)

        key, value, err := cur.Get(nil, nil, mdbx.First)
        if err != nil {
            log.Fatal(err)
        }

        for {
            key, value, err := cur.Get(key, value, mdbx.Next)
            // if bytes.Equal(key[:prefixLength], prefix) == false {
            // fmt.Println("Not FOund")
            // break
            // }
            if err != nil {
                if err == mdbx.NotFound {
                    break // No more records
                }
                log.Fatal(err)
            }

            // Process the key and value
            fmt.Printf("Key: %s, Value: %s\n", key, value)
            time.Sleep(3 * time.Second)
        }

        // fmt.Println("YOHO")

        return nil
    })
}