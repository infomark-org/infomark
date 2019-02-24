// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package helper

import (
  "errors"
  "fmt"
  "io"
  "mime/multipart"
  "net/http"
  "os"
  "strconv"

  _ "github.com/lib/pq"
)

type FileCategory int32

const (
  AvatarCategory FileCategory = 0
  SheetCategory  FileCategory = 1
)

type FileManager interface {
  WriteToBody(w http.ResponseWriter) error
  WriteToDisk(req multipart.File) error
  GetContentType() (string, error)
  Path(fallback bool) bool
  Delete() error
  Exists() bool
}

type FileHandle struct {
  Category FileCategory
  ID       int64 // an unique identifier (e.g. from database)
}

func NewAvatarHandle(userID int64) *FileHandle {
  return &FileHandle{
    Category: AvatarCategory,
    ID:       userID,
  }
}

// Path returns a path without checking if it exists. If fallback is true,
// the method tries to use the default value.
func (f *FileHandle) Path(fallback bool) string {
  switch f.Category {
  case AvatarCategory:
    potentialPath := fmt.Sprintf("files/uploads/avatars/%s.jpg", strconv.FormatInt(f.ID, 10))
    if fallback {
      if _, err := os.Stat(potentialPath); os.IsNotExist(err) {
        return "files/uploads/avatars/default.jpg"
      }
    }
    return potentialPath

  case SheetCategory:
    return fmt.Sprintf("files/uploads/sheets/%s.zip", strconv.FormatInt(f.ID, 10))
  }
  return ""
}

func (f *FileHandle) Exists() bool {
  file, err := os.Open(f.Path(false))
  if err != nil {
    return false
  }
  defer file.Close()
  return true
}
func (f *FileHandle) Delete() error {
  return os.Remove(f.Path(false))
}

func (f *FileHandle) GetContentType() (string, error) {

  // Only the first 512 bytes are used to sniff the content type.
  buffer := make([]byte, 512)

  file, err := os.Open(f.Path(true))
  if err != nil {
    return "", err
  }
  defer file.Close()

  _, err = file.Read(buffer)
  if err != nil {
    return "", err
  }

  // Use the net/http package's handy DectectContentType function. Always returns a valid
  // content-type by returning "application/octet-stream" if no others seemed to match.
  contentType := http.DetectContentType(buffer)

  return contentType, nil
}

func (f *FileHandle) WriteToBody(w http.ResponseWriter) error {

  // check if file exists
  file, err := os.Open(f.Path(true))
  if err != nil {
    return err
  }
  defer file.Close()

  // prepare header
  fileType, err := f.GetContentType()
  if err != nil {
    return err
  }
  w.Header().Set("Content-Type", fileType)

  // return file
  _, err = io.Copy(w, file)
  if err != nil {
    return err
  }

  return nil
}

func (f *FileHandle) WriteToDisk(r *http.Request, fieldName string) error {

  // receive data from post request
  if err := r.ParseMultipartForm(32 << 20); err != nil {
    return err
  }

  // we are interested in the field "avatar_data"
  file, handler, err := r.FormFile(fieldName)
  if err != nil {
    return err
  }
  defer file.Close()

  givenContentType := handler.Header["Content-Type"][0]

  switch f.Category {
  case AvatarCategory:
    switch givenContentType {
    case "image/jpeg":
    case "image/jpg":

    default:
      return errors.New("We support JPG/JPEG only.")

    }
  }

  // try to open new file
  hnd, err := os.OpenFile(f.Path(false), os.O_WRONLY|os.O_CREATE, 0666)
  if err != nil {
    return err
  }
  defer hnd.Close()

  // copy file from request
  bt, err := io.Copy(hnd, file)
  fmt.Println(bt)
  if err != nil {
    return err
  }

  return nil
}
