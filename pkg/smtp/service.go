// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smtp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/golang-jwt/jwt/v4"
)

type Config struct {
	Username string `env:"ILLA_MAIL_USERNAME" envDefault:"m17602200056@163.com"`
	Password string `env:"ILLA_MAIL_PASSWORD" envDefault:"ESXNALKGBIAZCSYO"`
	Host     string `env:"ILLA_MAIL_HOST" envDefault:"smtp.163.com"`
	Port     string `env:"ILLA_MAIL_PORT" envDefault:"465"`
	Secret   string `env:"ILLA_SECRET_KEY" envDefault:"ausNV5NJfVCrz3tPXtW2ZGGCpUuWFVQbikZ6d7FyOfpw9RcyLiNpqx4pJ6fSX9JXhMfmIupKKjQElURR"`
}

type SMTPServer struct {
	From     string
	Password string
	Host     string
	Port     string
	Secret   string
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}

func NewSMTPServer(cfg *Config) SMTPServer {
	return SMTPServer{
		From:     cfg.Username,
		Password: cfg.Password,
		Host:     cfg.Host,
		Port:     cfg.Port,
		Secret:   cfg.Secret,
	}
}

func (s *SMTPServer) NewVerificationCode(email string) (string, error) {
	// Authentication.
	auth := smtp.PlainAuth("", s.From, s.Password, s.Host)

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vCode := fmt.Sprintf("%06v", rnd.Int31n(1000000))

	header := make(map[string]string)
	header["From"] = "illa-builder" + "<" + s.From + ">"
	header["To"] = email
	header["Subject"] = "[Illa]: Verification Code"
	body := "Verification Code: " + vCode
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s:%s\r\n", k, v)
	}
	message += "\r\n" + body

	// Sending email.
	err := SendMailUsingTLS(s.Host+":"+s.Port, auth, s.From, email, []byte(message))
	if err != nil {
		return "", err
	}

	codeClaims := jwt.MapClaims{}
	codeClaims["id"] = vCode
	codeClaims["exp"] = time.Now().Add(60 * time.Second).Unix()
	codeClaims["iat"] = time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, codeClaims)
	codeToken, err := token.SignedString([]byte(s.Secret))
	if err != nil {
		return "", err
	}

	return codeToken, nil
}

func (s *SMTPServer) ValidateVerificationCode(codeToken, vCode string) (bool, error) {
	token, err := jwt.Parse(codeToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.Secret), nil
	})
	if err != nil {
		return false, err
	}

	var tokenCode string
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenCode = claims["id"].(string)
	} else {
		return false, errors.New("invalid token: token payload is invalid")
	}

	return tokenCode == vCode, err
}

func SendMailUsingTLS(addr string, auth smtp.Auth, from string, to string, msg []byte) (err error) {
	c, err := Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	tos := strings.Split(to, ";")
	for _, addr := range tos {
		if err = c.Rcpt(addr); err != nil {
			fmt.Print(err)
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}
