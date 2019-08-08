package utils

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func CheckErr(err error) bool {

	//TODO откуда вызвали
	_, file, line, _ := runtime.Caller(1)
	if err != nil {
		if os.Getenv("ENV") == "DEV" {
			log.Printf("%10s: %s:%d\n", err, filepath.Base(file), line)
			return true
		} else {
			log.Panicf("%10s: %s:%d\n", err, filepath.Base(file), line)
			return true
		}
	}
	return false
}

func ErrLog(err error) bool {

	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("%10s: %s:%d\n", err, filepath.Base(file), line)
		return true
	}
	return false

	//if os.Getenv("ENV") == "DEV" {

}

func ErrPanic(err error) bool {

	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Panicf("%10s: %s:%d\n", err, filepath.Base(file), line)
	}
	return false
}
