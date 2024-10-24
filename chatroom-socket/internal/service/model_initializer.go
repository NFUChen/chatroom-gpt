package service

//
//import (
//	"io"
//	"log"
//	"os"
//)
//
//var ModelScript string
//
//func init() {
//	file, err := os.Open("models.sql")
//	if err != nil {
//		log.Fatalf("failed to open SQL file: %v", err)
//	}
//	defer file.Close()
//	sqlBytes, err := io.ReadAll(file)
//	if err != nil {
//		log.Fatalf("failed to read SQL file: %v", err)
//	}
//
//	ModelScript = string(sqlBytes)
//	if len(ModelScript) == 0 {
//		log.Fatalf("failed to read SQL file")
//	}
//}
