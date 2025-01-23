package main

import (
	"fmt"
	"net/http"
	"os"

	"gorant/templates"
	"gorant/upload"
)

func (k *keycloak) adminUploadFileHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Start uploading")
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024) // (32 * 2^20) + 1024 bytes
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "An error occurred!"))
			return
		}
		// Not using r.MultipartForm, because I've only 1 file for 1 input field. If I use r.MultipartForm, I'd need to do
		// mpf.File["upload"][0].Filename, mpf.File["upload"][0].Open() etc.

		uploadedFile, header, err := r.FormFile("upload")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer uploadedFile.Close()
		fileType, err := checkFileType(uploadedFile)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		fileName, thumbnailFileName, uniqueKey, err := bc.UploadToBucket(uploadedFile, header.Filename, fileType)
		if err != nil {
			fmt.Println("Upload issue!!! ", err)
		}
		fmt.Println(uniqueKey, thumbnailFileName)
		if r.Header.Get("Hx-Request") == "" {
			TemplRender(w, r, templates.UploadAdmin("testing upload", k.currentUser, nil))
			return
		}

		TemplRender(w, r, templates.SuccessfulUpload(fileName))
	})
}

func (k *keycloak) adminViewUploadHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_ENV") != "TRUE" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		f, err := bc.ListBucket()
		if err != nil {
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}
		TemplRender(w, r, templates.UploadAdmin("testing upload", k.currentUser, f))
	})
}

func adminViewFileHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := bc.PublicAccessDomain
		fileID := r.PathValue("fileID")
		TemplRender(w, r, templates.ViewFile(domain, fileID))
	})
}

func (k *keycloak) uploadTestFileHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Start uploading")
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024) // (32 * 2^20) + 1024 bytes
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			w.Header().Set("Hx-Redirect", "/error")
			return
		}
		// Not using r.MultipartForm, because I've only 1 file for 1 input field. If I use r.MultipartForm, I'd need to do
		// mpf.File["upload"][0].Filename, mpf.File["upload"][0].Open() etc.
		uploadedFile, _, err := r.FormFile("upload")
		if err != nil {
			fmt.Println("form file error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer uploadedFile.Close()
		_, err = checkFileType(uploadedFile)
		if err != nil {
			fmt.Println("check file type", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		fileName, err := upload.ToLocalWebp(uploadedFile)
		if err != nil {
			fmt.Println("Upload issue!!! ", err)
		}
		if r.Header.Get("Hx-Request") == "" {
			TemplRender(w, r, templates.UploadAdmin("testing upload", k.currentUser, nil))
			return
		}
		TemplRender(w, r, templates.SuccessfulTestUpload(fileName))
	})
}

func (k *keycloak) adminViewDuplicateFilesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_ENV") != "TRUE" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		files, err := upload.GetOrphanFilesDB()
		if err != nil {
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		TemplRender(w, r, templates.ViewOrphanFiles("View Orphan Files", k.currentUser, files))
	})
}

func (k *keycloak) adminDeleteDuplicateFilesHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_ENV") != "TRUE" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		files, err := upload.GetOrphanFilesDB()
		if err != nil {
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}
		if err := bc.DeleteBucketFiles(files); err != nil {
			fmt.Println(err)
			w.Header().Set("Hx-Redirect", "/error")
			return
		}
		if err := upload.DeleteOrphanFilesDB(files); err != nil {
			fmt.Println(err)
			w.Header().Set("Hx-Redirect", "/error")
			return
		}
		TemplRender(w, r, templates.Toast("success", "Deleted successfully!"))
	})
}
