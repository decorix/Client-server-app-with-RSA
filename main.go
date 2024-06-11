package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
)

var accessUser string

func main() {
	a := app.New()
	w := a.NewWindow("Аутентификация по открытому ключу")
	w.Resize(fyne.NewSize(800, 600))

	w2 := a.NewWindow("Ресурсы (Режим пользователя)") // window for Auth User
	w2.Resize(fyne.NewSize(800, 600))

	w3 := a.NewWindow("Ресурсы (Режим администратора)") // window for Auth Admin
	w3.Resize(fyne.NewSize(800, 600))

	label := widget.NewLabel("Введите логин")
	entry := widget.NewEntry()

	label2 := widget.NewLabel("Зашифровать сообщение сервера")
	label3 := widget.NewLabel("Сообщение отсутствует")

	btnExitUser := widget.NewButton("Выйти", func() {
		w2.Hide()
		w.Show()
	})

	btnExitAdmin := widget.NewButton("Выйти", func() {
		w3.Hide()
		w.Show()
	})

	genPairEcKey()

	extractPublicKey("keypair.pem")

	btn := widget.NewButton("Войти", func() {
		data := entry.Text
		fmt.Println(data)

		// Отправляем логин на сервер
		login := entry.Text
		response, err := request("1", login, "public_key.pem")
		if err != nil {
			fmt.Println("Ошибка отправки логина:", err)
			return
		}
		parts := strings.Split(response, "::")

		randomMessage := parts[0]
		accessUser = parts[1]
		updateLabel(label3, randomMessage)
	})

	btn2 := widget.NewButton("Отправить логин и шифртекст", func() {
		//writeFile(label3.Text)

		singFile("keypair.pem")

		response, err := request("2", entry.Text, "file.signature")
		if err != nil {
			fmt.Println("Ошибка отправки шифртекста:", err)
			return
		}
		if response == "Аутентификация прошла успешно!" {
			if accessUser == "user" {
				w2.Show()
				w2.SetContent(container.NewVBox(
					btnExitUser))
				w.Hide()
			} else if accessUser == "admin" {
				w3.Show()
				w3.SetContent(container.NewVBox(
					btnExitAdmin))
				w.Hide()
			}

		}

	})

	w.SetContent(container.NewVBox(
		label, entry, btn, label2, label3, btn2))
	w.Show()
	w.SetMaster()
	a.Run()
}

func request(command string, login string, signFile string) (string, error) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		return "", fmt.Errorf("Ошибка подключения к серверу: %v", err)
	}
	defer conn.Close()

	fmt.Println("Успешное подключение к серверу")

	data, err := ioutil.ReadFile(signFile)
	if err != nil {
		fmt.Println("Ошибка чтения файла public_key.pem:", err)
	}

	_, err = conn.Write([]byte(command + "::" + login + "::" + string(data)))
	if err != nil {
		return "", fmt.Errorf("Ошибка отправки логина: %v", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("Ошибка чтения ответа от сервера: %v", err)
	}

	return string(buf[:n]), nil
}

func updateLabel(label *widget.Label, message string) {
	label.SetText(message)
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

func genPairEcKey() {
	cmd := exec.Command("openssl", "genpkey", "-algorithm", "EC", "-pkeyopt", "ec_paramgen_curve:secp521r1", "-out", "keypair.pem")
	err := cmd.Run()

	if err != nil {
		fmt.Println("Error: keypair.pem")
	}
	fmt.Println("Успешно: keypair.pem")

}

func extractPublicKey(keyPair string) {
	cmd := exec.Command("openssl", "pkey", "-in", keyPair, "-pubout", "-out", "public_key.pem")
	err := cmd.Run()

	if err != nil {
		fmt.Println("Error: public_key.pem")
	}
	fmt.Println("Успешно: public_key.pem")

}

func singFile(keyPair string) {
	cmd := exec.Command("openssl", "pkeyutl", "-sign", "-digest", "sha3-512", "-inkey", keyPair, "-in", "file.txt", "-rawin", "-out", "file.signature")
	err := cmd.Run()

	if err != nil {
		fmt.Println("Error: файл не подписан!")
	}

	fmt.Println("Файл подписан!")
}
