package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

var login string
var users = map[string]string{
	"user1": "user",
	"user2": "user",
	"admin": "admin",
}

func main() {
	ln, err := net.Listen("tcp", ":8080") // Принимаем соединения (порт 8080)
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Сервер запущен и ожидает соединений...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка при принятии соединения:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	//defer conn.Close()

	fmt.Println("Успешное подключение клиента", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	for {
		message := make([]byte, 1024)
		n, err := reader.Read(message)
		if err != nil {
			fmt.Println("Ошибка чтения сообщения от клиента:", err)
			return
		}

		fmt.Println(string(message[:n]))

		command := strings.Split(string(message[:n]), "::")[0]
		login = strings.Split(string(message[:n]), "::")[1]
		fileData := strings.Split(string(message[:n]), "::")[2]

		//fmt.Println("Команда:", command)
		//fmt.Println("Логин:", data)
		//fmt.Println("Название файла:", fileName)

		if command == "1" {
			accessRights, ok := users[login]
			if !ok {
				// Если пользователь не найден
				//fmt.Println("Пользователь", login, "не найден")
				writeData(conn, "Пользователь "+login+" не найден")
			} else {
				// Если пользователь найден
				fmt.Println("Права доступа для пользователя", login, ":", accessRights)

				file, err := os.Create("public_key.pem")
				if err != nil {
					fmt.Println("Ошибка создания файла public_key.pem:", err)
					continue
				}
				defer file.Close()

				_, err = file.WriteString(fileData)
				if err != nil {
					fmt.Println("Ошибка записи данных в файл:", err)
					continue
				}

				fmt.Println("Логин:", login)
				randomMessage := generateRandomString(10)

				writeFile(randomMessage)
				writeData(conn, randomMessage+"::"+accessRights)
			}

		} else if command == "2" {
			file, err := os.Create("file.signature")
			if err != nil {
				fmt.Println("Ошибка создания файла file.signature:", err)
				continue
			}
			defer file.Close()

			_, err = file.WriteString(fileData)
			if err != nil {
				fmt.Println("Ошибка записи данных в файл:", err)
				continue
			}
			res := verify()
			fmt.Println("res:" + res)
			if strings.TrimSpace(res) == "Signature Verified Successfully" {
				writeData(conn, "Аутентификация прошла успешно!")
			} else {
				writeData(conn, "Аутентификация не пройдена!")
			}

		} else {
			fmt.Println("Неправильный формат сообщения")
		}
	}

}

func generateRandomString(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return ""
	}

	randomString := base64.URLEncoding.EncodeToString(randomBytes)

	return randomString
}

func writeData(conn net.Conn, data string) error {
	_, err := conn.Write([]byte(data))
	if err != nil {
		//fmt.Println("Ошибка отправки данных:", err)
		return err
	}
	return nil
}

func verify() string {
	cmd := exec.Command("openssl", "pkeyutl", "-verify", "-digest", "sha3-512", "-inkey", "public_key.pem",
		"-pubin", "-in", "file.txt", "-rawin", "-sigfile", "file.signature")

	var out bytes.Buffer

	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println("Ошибка выполнения команды:", err)
		return "error"
	}

	result := out.String()

	return result
}

func writeFile(message string) {
	file, err := os.Create("file.txt")
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(message)
	if err != nil {
		fmt.Println("Ошибка записи в файл:", err)
		return
	}

	fmt.Println("Строка успешно записана в файл.")
}
