package main

import (
    "cloud.google.com/go/storage"
    "context"
    "database/sql"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"

    _ "github.com/lib/pq"
)

var (
    db            *sql.DB
    bucketName    string
    filestorePath string
)

func main() {
    // Configuration from env
    dbUser := os.Getenv("DATABASE_USER")
    dbPassword := os.Getenv("DATABASE_PASSWORD")
    dbName := os.Getenv("DATABASE_NAME")
    dbHost := os.Getenv("DATABASE_HOST")
    bucketName = os.Getenv("BUCKET_NAME")
    filestorePath = os.Getenv("FILESTORE_PATH")

    connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
        dbHost, dbUser, dbPassword, dbName)
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatalf("DB connection error: %v", err)
    }
    defer db.Close()
    if err = db.Ping(); err != nil {
        log.Fatalf("DB ping error: %v", err)
    }

    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/files", filesHandler)
    http.HandleFunc("/upload", uploadHandler)
    http.HandleFunc("/bucket/list", bucketListHandler)
    http.HandleFunc("/bucket/upload", bucketUploadHandler)

    log.Println("Server on :8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "<h1>Hello from GKE WebApp!</h1>")
}

func filesHandler(w http.ResponseWriter, r *http.Request) {
    files, err := os.ReadDir(filestorePath)
    if err != nil {
        http.Error(w, "Failed to read filestore", http.StatusInternalServerError)
        return
    }
    fmt.Fprintln(w, "<h2>Filestore Files:</h2><ul>")
    for _, f := range files {
        fmt.Fprintf(w, "<li>%s</li>", f.Name())
    }
    fmt.Fprintln(w, "</ul>")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprintln(w, `<form enctype="multipart/form-data" method="post">`+
            `<input type="file" name="file" /><input type="submit" value="Upload" />`+
            `</form>`)
        return
    }
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Failed to get file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    dstPath := filepath.Join(filestorePath, header.Filename)
    dst, err := os.Create(dstPath)
    if err != nil {
        http.Error(w, "Failed to create file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()
    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, "Failed to save file", http.StatusInternalServerError)
        return
    }
    fmt.Fprintf(w, "Uploaded %s!", header.Filename)
}

func bucketListHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    client, err := storage.NewClient(ctx)
    if err != nil {
        http.Error(w, "Storage client error", http.StatusInternalServerError)
        return
    }
    defer client.Close()

    it := client.Bucket(bucketName).Objects(ctx, nil)
    fmt.Fprintln(w, "<h2>Bucket Objects:</h2><ul>")
    for {
        attrs, err := it.Next()
        if err == storage.Done {
            break
        }
        if err != nil {
            http.Error(w, "Listing failed", http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "<li>%s</li>", attrs.Name)
    }
    fmt.Fprintln(w, "</ul>")
}

func bucketUploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprintln(w, `<form enctype="multipart/form-data" method="post">`+
            `<input type="file" name="file" /><input type="submit" value="Upload to Bucket" />`+
            `</form>`)
        return
    }
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Failed to get file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    ctx := context.Background()
    client, err := storage.NewClient(ctx)
    if err != nil {
        http.Error(w, "Storage client error", http.StatusInternalServerError)
        return
    }
    defer client.Close()

    wc := client.Bucket(bucketName).Object(header.Filename).NewWriter(ctx)
    if _, err := io.Copy(wc, file); err != nil {
        http.Error(w, "Upload failed", http.StatusInternalServerError)
        return
    }
    wc.Close()
    fmt.Fprintf(w, "Uploaded %s to bucket!", header.Filename)
}
