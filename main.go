package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func (user *User) String() string {
	b, _ := json.Marshal(*user)
	return string(b)
}

func (user *User) Set(s string) error {
	return json.Unmarshal([]byte(s), user)
}

func parseArgs() Arguments {
	operation := flag.String("operation", "", "usage: -operation add|list|findById|remove")
	var user User
	flag.Var(&user, "item", "usage: -item {\"id\": \"1\", \"email\": \"email@test.com\", \"age\": 23}")
	fileName := flag.String("fileName", "", "usage: -fileName \"user.json\"")
	flag.Parse()

	return Arguments{"operation": *operation, "item": user.String(), "fileName": *fileName}
}

func overwriteFileWithNewUsers(file *os.File, fileCtx []User) error {
	res, err := json.Marshal(fileCtx)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	err2 := overwriteFileWithBytes(file, res)
	if err2 != nil {
		return err2
	}
	return nil
}

func overwriteFileWithBytes(file *os.File, content []byte) error {
	patherr := file.Truncate(0)
	if patherr != nil {
		return fmt.Errorf("%w", patherr)
	}
	_, seekErr := file.Seek(0, 0)
	if seekErr != nil {
		return fmt.Errorf("%w", seekErr)
	}
	code, err := file.Write(content)
	if code == 0 || err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func Perform(args Arguments, writer io.Writer) error {
	fileName := args["fileName"]
	if fileName == "" {
		return errors.New("-fileName flag has to be specified")
	}
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	var item string
	var id string
	switch operation := args["operation"]; operation {
	case "":
		return errors.New("-operation flag has to be specified")
	case "list":
		_, err := writer.Write(buf)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	case "add":
		item = args["item"]
		if item == "" {
			return errors.New("-item flag has to be specified")
		}
		var userItem User
		err3 := json.Unmarshal([]byte(item), &userItem)
		if err3 != nil {
			return fmt.Errorf("%w", err3)
		}

		fileCtx := make([]User, 0, 50)
		if len(buf) > 0 {
			err2 := json.Unmarshal(buf, &fileCtx)
			if err2 != nil {
				return fmt.Errorf("%w", err2)
			}
			
		}
		if len(fileCtx) > 0 {
			for _, u := range fileCtx {
				if u.Id == userItem.Id {
					writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", userItem.Id)))
				}
			}
		}
		fileCtx = append(fileCtx, userItem)
		overwriteErr := overwriteFileWithNewUsers(file, fileCtx)
		if overwriteErr != nil {
			return fmt.Errorf("%w", overwriteErr)
		}

	case "remove":
		id = args["id"]
		if id == "" {
			return errors.New("-id flag has to be specified")
		}
		fileCtx := make([]User, 0, 50)
		err2 := json.Unmarshal(buf, &fileCtx)
		if err2 != nil {
			return fmt.Errorf("%w", err2)
		}
		found := false
		for i, u := range fileCtx {
			if u.Id == id {
				found = true
				fileCtx = append(fileCtx[:i], fileCtx[i+1:]...)
				overwriteErr := overwriteFileWithNewUsers(file, fileCtx)
				if overwriteErr != nil {
					return fmt.Errorf("%w", overwriteErr)
				}
			}
		}
		if !found {
			writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))
		}
	case "findById":
		id = args["id"]
		if id == "" {
			return errors.New("-id flag has to be specified")
		}
		fileCtx := make([]User, 0, 50)
		err2 := json.Unmarshal(buf, &fileCtx)
		if err2 != nil {
			return fmt.Errorf("%w", err2)
		}
		found := false
		for _, u := range fileCtx {
			if u.Id == id {
				found = true
				writer.Write([]byte(u.String()))
			}
		}
		if !found {
			writer.Write([]byte(""))
		}
	default:
		return errors.New("Operation abcd not allowed!")
	}

	return nil
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
