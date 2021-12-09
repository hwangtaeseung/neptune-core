package common

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func RemoveFileExtension(inputFile string) string {
	ext := filepath.Ext(inputFile)
	if ext != "" {
		return strings.TrimSuffix(inputFile, filepath.Ext(inputFile))
	}
	return inputFile
}

func ToJson(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

func ToJsonAsString(object interface{}) string {
	jsonBytes, _ := ToJson(object)
	return string(jsonBytes)
}

func ToJsonBeautifully(object interface{}) string {
	jsonBytes, _ := ToJson(object)
	jsonString, _ := BeautifyJson(jsonBytes)
	return jsonString
}

func FromJson(byteJson []byte, object interface{}) error {
	return json.Unmarshal(byteJson, object)
}

func BeautifyJson(byteJson []byte) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, byteJson, "", "\t"); err != nil {
		return "", err
	}
	return string(prettyJSON.Bytes()), nil
}

func NVL(value interface{}, replace interface{}) interface{} {
	if value == nil {
		return replace
	}
	switch value.(type) {
	case string:
		if value == "" {
			return replace
		}
	}
	return value
}

func IF(is bool, trueValue interface{}, falseValue interface{}) interface{} {
	if is {
		return trueValue
	} else {
		return falseValue
	}
}

type GoRoutineFunc func(ctx context.Context) interface{}
func GoRoutineWithContext(ctx context.Context, callbacks ...GoRoutineFunc) ([]interface{}, error) {

	if callbacks == nil {
		return nil, errors.New("callbacks is nil")
	}

	done := make(chan interface{})
	defer close(done)

	// execute go routine
	for _, callback := range callbacks {
		callback := callback
		go func() {
			done <- callback(ctx)
		}()
	}

	// receive go routine return value
	var results []interface{}
	for range callbacks {
		select {
		case result := <-done:
			results = append(results, result)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return results, nil
}

func GetMd5OfFile(inputFile string) (string, error) {

	file, err := os.Open(inputFile)
	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("file close err : %v", err)
		}
	}()
	if err != nil {
		log.Printf("file (%v) open err : %v", inputFile, err)
		return "", err
	}

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)[:16]

	return hex.EncodeToString(hashInBytes), nil
}

func DurationToSecond(duration string) (sec float64) {
	durationArray := strings.Split(duration, ":")
	if len(durationArray) != 3 {
		return
	}
	hours, _ := strconv.ParseFloat(durationArray[0], 64)
	minutes, _ := strconv.ParseFloat(durationArray[1], 64)
	seconds, _ := strconv.ParseFloat(durationArray[2], 64)
	return hours * (60 * 60) + minutes * (60) + seconds
}

func RetryWrapper(callback func() (interface{}, error), retryCount int) (interface{}, error) {
	var err error
	var result interface{}
	for count := 0; count < retryCount; count++ {
		if result, err = callback(); err != nil {
			log.Printf("An error occurred. try again... (count:%v, error:%v)\n", count, err)
			// wait for a second
			time.Sleep(time.Second)
			continue
		} else {
			return result, nil
		}
	}
	return nil, err
}

func GetWhereIs() string {
	switch runtime.GOOS {
	case "windows":
		return "where"
	default:
		return "which"
	}
}

func GetCarriageReturn() byte {
	switch runtime.GOOS {
	case "windows":
		return '\r'
	default:
		return '\n'
	}
}

func StringToFile(fileName string, data string) error {
	return ioutil.WriteFile(fileName, []byte(data), 0644)
}

func ObjectToFile(fileName string, data interface{}) error {
	if objectBytes, err := ToJson(data); err != nil {
		return err
	} else {
		return ioutil.WriteFile(fileName, objectBytes, 0644)
	}
}

func JsonFileToObject(fileName string, object interface{}) error {
	if jsonBytes, err := ioutil.ReadFile(fileName); err != nil {
		return err
	} else {
		if err := FromJson(jsonBytes, object); err != nil {
			return err
		}
	}
	return nil
}

func QueryString(headers map[string]string) string {
	headerString := ""
	for key, value := range headers {
		headerString += fmt.Sprintf("%v=%v&", key, value)
	}
	return headerString
}

func GroupLoop(originArray []interface{}, concurrencyCount int, callback func(group []interface{}) error) error {
	count := len(originArray)
	groupArray := originArray
	for index := 0; index < count; index += concurrencyCount {
		if len(originArray) >= concurrencyCount {
			groupArray = originArray[:concurrencyCount]
			originArray = originArray[concurrencyCount:]
		} else {
			groupArray = originArray
		}
		if err := callback(groupArray); err != nil {
			log.Printf("loop callback error : %v\n", err)
			return err
		}
	}
	return nil
}