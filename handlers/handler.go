package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"test-technical-golang/databases"
	"test-technical-golang/models"

	"github.com/labstack/echo/v4"
)

func GenerateRandomKey() ([]byte, error) {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	return key, err
}

var globalKey []byte

func init() {
	key, err := GenerateRandomKey()
	if err != nil {
		panic(err)
	}
	globalKey = key
}

var users = map[string]string{
	"username@example.com": "password",
}

func Login(c echo.Context) error {
	email := c.FormValue("username@example.com")
	password := c.FormValue("password")

	if storedPassword, ok := users[email]; ok && storedPassword == password {
		c.SetCookie(&http.Cookie{
			Name:  "username@example.com",
			Value: email,
		})
		return c.Redirect(http.StatusSeeOther, "/input")
	}

	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"Error": "Invalid email or password",
	})
}

func ProcessForm(c echo.Context) error {
	if c.FormValue("generate") == "true" {

		for i := 0; i < 25; i++ {
			randomNumber := rand.Intn(8999999999) + 1000000000
			providers := []string{"XL", "AXIS", "Telkomsel", "Tri"}
			randomProvider := providers[rand.Intn(len(providers))]

			encryptedNumber, err := EncryptPhone(strconv.Itoa(randomNumber), globalKey)
			if err != nil {
				return err
			}

			formData := models.PhoneNumber{
				PhoneNumbers: encryptedNumber,
				Provider:     randomProvider,
			}

			if err := databases.DB.Create(&formData).Error; err != nil {
				log.Printf("Database error: %v", err)
				return err
			}
		}
		return c.Redirect(http.StatusSeeOther, "/output")
	}
	number := c.FormValue("phone_number")
	provider := c.FormValue("provider")

	encryptedNumber, err := EncryptPhone(number, globalKey)
	if err != nil {
		return err
	}

	formData := models.PhoneNumber{
		PhoneNumbers: encryptedNumber,
		Provider:     provider,
	}

	if err := databases.DB.Create(&formData).Error; err != nil {
		return err
	}

	return c.Render(http.StatusOK, "input.html", formData)
}

func ShowInputPage(c echo.Context) error {
	return c.Render(http.StatusOK, "input.html", nil)
}

func ShowOutputPage(c echo.Context) error {
	var allPhoneNumbers []models.PhoneNumber
	if err := databases.DB.Find(&allPhoneNumbers).Error; err != nil {
		return err
	}

	var oddPhoneNumbers []models.PhoneNumber
	var evenPhoneNumbers []models.PhoneNumber

	for _, phoneNumber := range allPhoneNumbers {
		if phoneNumber.PhoneNumbers == "" {
			continue
		}

		decryptedNumber, err := DecryptPhone(phoneNumber.PhoneNumbers, globalKey)
		if err != nil {
			return err
		}

		phoneNumber.PhoneNumbers = decryptedNumber

		number, err := strconv.Atoi(phoneNumber.PhoneNumbers)
		if err != nil {
			return err
		}

		if number%2 == 0 {
			evenPhoneNumbers = append(evenPhoneNumbers, phoneNumber)
		} else {
			oddPhoneNumbers = append(oddPhoneNumbers, phoneNumber)
		}
	}

	data := map[string]interface{}{
		"OddPhoneNumbers":  oddPhoneNumbers,
		"EvenPhoneNumbers": evenPhoneNumbers,
	}

	return c.Render(http.StatusOK, "output.html", data)

}

const (
	keySize = 32
	scryptN = 16384
	scryptR = 8
	scryptP = 1
)

func EncryptPhone(phoneNumber string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(phoneNumber), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptPhone(encryptedPhoneNumber string, key []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPhoneNumber)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func EditPhoneNumber(c echo.Context) error {
	phoneNumberID := c.FormValue("id")

	var phoneNumber models.PhoneNumber
	if err := databases.DB.First(&phoneNumber, phoneNumberID).Error; err != nil {
		return err
	}

	return c.Render(http.StatusOK, "edit.html", map[string]interface{}{
		"PhoneNumber": phoneNumber,
	})
}

func UpdatePhoneNumber(c echo.Context) error {
	phoneNumberID := c.FormValue("id")
	updatedPhoneNumber := c.FormValue("phone_number")
	updatedProvider := c.FormValue("provider")

	var phoneNumber models.PhoneNumber
	if err := databases.DB.First(&phoneNumber, phoneNumberID).Error; err != nil {
		return err
	}

	encryptedNumber, err := EncryptPhone(updatedPhoneNumber, globalKey)
	if err != nil {
		return err
	}

	phoneNumber.PhoneNumbers = encryptedNumber
	phoneNumber.Provider = updatedProvider

	if err := databases.DB.Save(&phoneNumber).Error; err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/output")
}

func DeletePhoneNumber(c echo.Context) error {
	phoneNumberID := c.FormValue("id")

	var phoneNumber models.PhoneNumber
	if err := databases.DB.First(&phoneNumber, phoneNumberID).Error; err != nil {
		return err
	}

	if err := databases.DB.Delete(&phoneNumber).Error; err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/output")
}
