package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jhillyerd/enmime"
	"github.com/mattn/godown"
	"github.com/nfnt/resize"
)

func fatal(s interface{}) {
	fmt.Fprintln(os.Stderr, s)
	os.Exit(1)
}

func msgSlug(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:20])
}

func saveJpeg(fname string, attachment *enmime.Part) error {
	img, _, err := image.Decode(bytes.NewReader(attachment.Content))
	if err != nil {
		return err
	}
	if img.Bounds().Dx() > 800 || img.Bounds().Dy() > 800 {
		img = resize.Resize(800, 0, img, resize.Lanczos3)
	}
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, nil)
}

func clean() error {
	commands := [][]string{
		{"git", "reset"},
		{"git", "checkout", "."},
		{"git", "reset", "--hard", "HEAD"},
		{"git", "clean", "-fdx"},
	}
	for _, args := range commands {
		err := exec.Command(args[0], args[1:]...).Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var mailserver string
	var accept string
	var sender string
	var repo string
	var usehtml bool
	flag.StringVar(&mailserver, "m", "localhost:25", "mail server")
	flag.StringVar(&accept, "a", "*", "accept e-mail from")
	flag.StringVar(&sender, "s", "moblog@example.com", "e-mail sender")
	flag.StringVar(&repo, "d", "/path/to/jekyll/blog", "repository of jekyll")
	flag.BoolVar(&usehtml, "h", false, "use html")
	flag.Parse()

	err := os.Chdir(repo)
	if err != nil {
		log.Fatalf("cannot chdir: %v", err)
	}

	err = clean()
	if err != nil {
		log.Fatalf("cannot git clean: %v", err)
	}

	env, err := enmime.ReadEnvelope(os.Stdin)
	if err != nil {
		log.Fatalf("cannot parse e-mail: %v", err)
	}

	addr, err := mail.ParseAddress(env.GetHeader("From"))
	if err != nil {
		log.Fatalf("cannot parse from address: %v", err)
	}

	if usehtml && env.HTML != "" {
		var buf bytes.Buffer
		if err := godown.Convert(&buf, strings.NewReader(env.HTML), nil); err == nil {
			env.Text = buf.String()
		}
	}
	body := strings.ReplaceAll(strings.ReplaceAll(env.Text, "\r", ""), "\n", "\n\n")
	text := fmt.Sprintf(`---
layout: post
title: %s
date: %s
---
%s
`, env.GetHeader("Subject"), time.Now().Format(`2006-01-02 15:04:05.999999999 -0700 MST`), body)

	slug := msgSlug(env.GetHeader("Subject"))
	n := 1
	for _, attachment := range env.Inlines {
		if !strings.HasPrefix(attachment.ContentType, "image/") {
			continue
		}
		file := filepath.ToSlash(filepath.Join("assets", slug+fmt.Sprintf("-%03d.jpg", n)))
		err = saveJpeg(file, attachment)
		if err != nil {
			log.Fatalf("cannot write attachment file: %v", err)
		}
		err = exec.Command("git", "add", file).Run()
		if err != nil {
			log.Fatalf("cannot execute git add: %v", err)
		}

		marker := "[image: " + attachment.FileName + "]"
		text = strings.ReplaceAll(text, marker, fmt.Sprintf(`![%s](%s)`, attachment.FileName, "/"+file))
		n++
	}
	for _, attachment := range env.OtherParts {
		if !strings.HasPrefix(attachment.ContentType, "image/") {
			continue
		}
		file := filepath.ToSlash(filepath.Join("assets", slug+fmt.Sprintf("-%03d.jpg", n)))
		err = saveJpeg(file, attachment)
		if err != nil {
			log.Fatalf("cannot write attachment file: %v", err)
		}
		err = exec.Command("git", "add", file).Run()
		if err != nil {
			log.Fatalf("cannot execute git add: %v", err)
		}

		text += fmt.Sprintf("![%s](%s)\n\n", attachment.FileName, "/"+file)
		n++
	}

	file := filepath.Join("_posts", time.Now().Format("2006-01-02")+"-"+slug+".md")
	err = ioutil.WriteFile(file, []byte(text), 0644)
	if err != nil {
		log.Fatalf("cannot create new entry: %v", err)
	}
	err = exec.Command("git", "add", file).Run()
	if err != nil {
		log.Fatalf("cannot execute git add: %v", err)
	}
	err = exec.Command("git", "commit", "--no-gpg-sign", "-a", "-m", "Add entry: "+env.GetHeader("Subject")).Run()
	if err != nil {
		log.Fatalf("cannot execute git commit: %v", err)
	}
	err = exec.Command("git", "push", "--force", "origin", "master").Run()
	if err != nil {
		log.Fatalf("cannot execute git push: %v", err)
	}

	from := mail.Address{Name: "うんこ", Address: sender}
	message := fmt.Sprintf(`To: %s
From: %s
Reference: %s
Subject: %s

投稿が完了しました
`, addr.String(), from.String(), env.GetHeader("Message-ID"), "RE: "+env.GetHeader("Subject"))

	err = smtp.SendMail(mailserver, nil, from.Address, []string{addr.Address}, []byte(message))
	if err != nil {
		log.Fatalf("cannot send e-mail: %v", err)
	}
}
