package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

var (
	port     = flag.String("port", ":8888", "access port")
	dist_dir = flag.String("dir", "/srv/Downloads", "upload")
)
var uploadTemplate, _ = template.ParseFiles("upload.html")

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		uploadTemplate.Execute(w, nil)
	} else {
		part_reader, err := r.MultipartReader()
		if err != nil {
			log.Println("get file:", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Println("start copy")
		var file_part *multipart.Part
		for {
			if file_part, err = part_reader.NextPart(); err != nil {
				if err == io.EOF {
					err = nil
				}
				break
			}
			if file_part.FormName() == "file" {
				if err = write_file(file_part); err != nil {
					break
				}
			}
			file_part.Close()
		}
		if err != nil {
			log.Println("write file:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/upload", 302)
	}
}
func write_file(part *multipart.Part) error {
	dir_name := *dist_dir +
		"/" + time.Now().Format("2006-01-02")
	file_name := dir_name + "/" + part.FileName()
	if err := os.Mkdir(dir_name, 0755); err != nil {
		if os.IsNotExist(err) {
			return err
		}
	}
	if fd, err := os.Open(file_name); err == nil {
		fd.Close()
		return nil
	}
	var err error
	var newfile *os.File
	if newfile, err = os.Create(file_name); err != nil {
		return err
	}
	log.Println("create", file_name)
	defer newfile.Close()
	buf := make([]byte, 1024*1024)
	for {
		n, err := part.Read(buf)
		newfile.Write(buf[:n])
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			os.Remove(file_name)
			log.Print("remove", file_name)
			break
		}
	}
	return err
}
func main() {
	flag.Parse()
	http.HandleFunc("/upload", upload)
	http.ListenAndServe(*port, nil)
}
